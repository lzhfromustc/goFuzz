package main

import (
	"flag"
	"fmt"
	"goFuzz/config"
	"goFuzz/fuzzer"
	"sync"
)

func main() {

	fuzzer.SetDeadline()

	// Parse input
	pProjectPath := flag.String("path","","Full path of the target project")
	pProjectGOPATH := flag.String("GOPATH","","GOPATH of the target project")
	//pTestName := flag.String("test","","Function name of the unit test")
	pOutputFullPath := flag.String("output","","Full path of the output file")
	pModeGlobalTuple := flag.Bool("globalTuple", false, "Whether prev_location is global or per channel")
	maxParallel := flag.Int("maxparallel", 1, "Specified the maximum subroutine number for fuzzing.")

	flag.Parse()

	config.StrTestPath = *pProjectPath
	config.StrProjectGOPATH = *pProjectGOPATH
	//config.StrTestName = *pTestName
	config.StrOutputFullPath = *pOutputFullPath
	config.BoolGlobalTuple = *pModeGlobalTuple

	var wg sync.WaitGroup
	workerInputChan := make(chan *fuzzer.Input)
	workerOutputChan := make(chan *fuzzer.RunOutput)


	/* Begin running specific number of worker subroutines. Blocked before we send them inputs from the main routine. */
	for i := 0; i < *maxParallel; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for currentInput := range workerInputChan {
				fmt.Println("Working with go subroutine: ", i)
				retOutput := fuzzer.Run(currentInput)
				workerOutputChan <- retOutput
			}
		}(i)
	}

	/* TODO:: Not finished in this part!!!
	In this part, we should implement an algorithm that can iterate all the possible unit test inside our target program.
	Right now, we are using an ad-hoc pre-defined slice pTestnameList for the DEMO purpose.
	 */
	var pTestNameList []string
	pTestNameList = append(pTestNameList, "TestF1")

	/* Standard initialize the necessary data structure. */
	var fuzzingQueue []fuzzer.FuzzQueryEntry
	var mainRecord = fuzzer.EmptyRecord()
	allRecordHashMap := make(map[string]struct{})

	for _, pTestName := range pTestNameList {
		//  Run the instrumented program for one time,
		//  generate original input and record
		var emptyInput = fuzzer.EmptyInput()
		emptyInput.TestName = pTestName
		// Give one job to the worker. And receive results.
		workerInputChan <- emptyInput
		retOutput := <- workerOutputChan

		retInput := retOutput.RetInput

		/* Generate all possible deter_inputs based on the current retInput. Only changing one select per time. */
		deterInputSlice := fuzzer.Deterministic_enumerate_input(retInput)

		/* Now we deterministically enumerate one select per time, iterating through all the pTestName and all selects */
		for  _, deterInput := range deterInputSlice {
			deterInput.Stage = "deter"  // The current stage of fuzzing is "deterministic fuzzing".
			for {
				select {
					case workerInputChan <- deterInput:
						fmt.Println("Deter: Insert an input to the workers. ")
						break
					case retOutput := <- workerOutputChan:
						if retOutput == nil || retOutput.RetInput == nil || retOutput.RetRecord == nil {
							// TODO:: I found some empty retInput again. Should I be worry?
							fmt.Println("Error!!! Empty retInput!!!")
							continue
						}
						fmt.Println("Deter: Reading outputs from the workers. ")
						/* We don't need to care about the running stage here. It would always be "deter". */
						retInputInDeter := retOutput.RetInput
						retRecord := retOutput.RetRecord
						fuzzer.HandleRunOutput(retInputInDeter, retRecord, retOutput.Stage, nil, mainRecord, fuzzingQueue, allRecordHashMap)
					default:
						//fmt.Println("Waiting for the worker to complete their jobs. ")
						continue
				}
			}
		}
	}

	/* Infinite loop for the main fuzzing */
	mainRandomLoopIdx := 0
	for {
		fmt.Println("Beginning main random fuzzing loop idx: ", mainRandomLoopIdx)
		if len(fuzzingQueue) == 0 {
			fmt.Println("Fuzzing Queue is nil (no components). Some error occurs. ")
			continue
		}
		// TODO:: Maybe we should cull the queue first. (Or maybe after the calibration?)
		// TODO:: Prioritize queue based on scores.
		fuzzingQueueLen := len(fuzzingQueue)
		for fuzzingIdx := 0; fuzzingIdx < fuzzingQueueLen; fuzzingIdx++ {
			currentEntry := fuzzingQueue[fuzzingIdx]
			/* Calibrate case. */
			for i := 0; i < 1; i++ {   // TODO:: Maybe we can have multiple times of Calibrate case here. Since the retRecord might not be completely stable.
				// TODO:: If the seed case has already been calibrated, maybe we can skip the duplicated calibrate case.
				// TODO:: There seems to be no way to get an error message from the Run func?
				// TODO:: Set calibration_failed to the queue entry if calibration failed (fuzz.Run() failed)
				currentEntry.CurrentInput.Stage = "calib"
				for {
					select {
					case workerInputChan <- currentEntry.CurrentInput:
						fmt.Println("Calib: Insert an input to the workers. ")
						break
					case retOutput := <- workerOutputChan:
						if retOutput == nil || retOutput.RetInput == nil || retOutput.RetRecord == nil {
							// TODO:: I found some empty retInput again. Should I be worry?
							fmt.Println("Error!!! Empty retInput!!!")
							continue
						}
						fmt.Println("Calib or Rand: Reading outputs from the workers. ")
						retInput := retOutput.RetInput
						retRecord := retOutput.RetRecord
						fuzzer.HandleRunOutput(retInput, retRecord, retOutput.Stage, &currentEntry, mainRecord, fuzzingQueue, allRecordHashMap)
					default:
						//fmt.Println("Waiting for the worker to complete their jobs. ")
						continue
					}
				}
			}

			/* Next, Random fuzzing */
			// TODO:: Random fuzzing with dynamic energy. (Maybe depends on the scores, executionCount etc)
			currentFuzzingEnergy := 100
			for randFuzzIdx := 0; randFuzzIdx < currentFuzzingEnergy; randFuzzIdx++ {
				currentMutatedInput := fuzzer.Random_Mutate_Input(currentEntry.CurrentInput)
				currentMutatedInput.Stage = "rand"
				for {
					select {
					case workerInputChan <- currentMutatedInput:
						fmt.Println("Rand: Insert an input to the workers. ")
						break
					case retOutput := <- workerOutputChan:
						if retOutput == nil || retOutput.RetInput == nil || retOutput.RetRecord == nil {
							// TODO:: I found some empty retInput again. Should I be worry?
							fmt.Println("Error!!! Empty retInput!!!")
							continue
						}
						fmt.Println("Calib or Rand: Reading outputs from the workers. ")
						retInput := retOutput.RetInput
						retRecord := retOutput.RetRecord
						fuzzer.HandleRunOutput(retInput, retRecord, retOutput.Stage, &currentEntry, mainRecord, fuzzingQueue, allRecordHashMap)
					default:
						//fmt.Println("Waiting for the worker to complete their jobs. ")
						continue
					}
				}
			}
		}
		mainRandomLoopIdx++
	}

	//
	//// fuzzing loop
	//for len(workList) > 0 {
	//	// Pop the first element of worklist
	//	curInput, _ := fuzzer.PopWorklist(&workList)
	//	workListScore = workListScore[1:]
	//
	//	// Check if input has been executed
	//	hash := fuzzer.HashOfInput(curInput)
	//	if _, executed := mapExecutedHash[hash]; executed {
	//		continue
	//	}
	//	mapExecutedHash[hash] = struct{}{}
	//
	//	// Run new input
	//	_, curRecord := fuzzer.Run(config.StrTestName, curInput)
	//	curScore := fuzzer.ComputeScore(mainRecord, curRecord)
	//	mainRecord = fuzzer.UpdateMainRecord(mainRecord, curRecord)
	//
	//	// Based on current input, generate multiple new inputs
	//	// and put them into the workList by curScore
	//	newInputs := fuzzer.GenInputs(curInput)
	//
	//	if len(workListScore) == 0 {
	//		workList = newInputs
	//		workListScore = []int{}
	//		for i := 0; i < len(workList); i++ {
	//			workListScore = append(workListScore, curScore)
	//		}
	//	} else {
	//		var indexInsertAfter int // insert the new inputs after this index. -1 stands for insert before the head
	//		for i := -1; i < len(workListScore); i++ {
	//			if i == -1 {
	//				if curScore >= workListScore[0] {
	//					indexInsertAfter = i
	//				}
	//			} else if i == len(workListScore) - 1 {
	//				if curScore <= workListScore[len(workListScore) - 1] {
	//					indexInsertAfter = i
	//				}
	//			} else {
	//				if workListScore[i] >= curScore && workListScore[i + 1] <= curScore {
	//					indexInsertAfter = i
	//				}
	//			}
	//		}
	//		workList = fuzzer.InsertWorklist(newInputs, workList, indexInsertAfter)
	//		workListScore = fuzzer.InsertWorklistScore(curScore, len(newInputs), workListScore, indexInsertAfter)
	//	}
	//}

	return

}

