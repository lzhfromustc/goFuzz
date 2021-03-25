package main

import (
	"flag"
	"fmt"
	"goFuzz/config"
	"goFuzz/fuzzer"
)

func main() {

	fuzzer.SetDeadline()

	// Parse input
	pProjectPath := flag.String("path","","Full path of the target project")
	pProjectGOPATH := flag.String("GOPATH","","GOPATH of the target project")
	//pTestName := flag.String("test","","Function name of the unit test")
	pOutputFullPath := flag.String("output","","Full path of the output file")
	pModeGlobalTuple := flag.Bool("globalTuple", false, "Whether prev_location is global or per channel")

	flag.Parse()

	config.StrTestPath = *pProjectPath
	config.StrProjectGOPATH = *pProjectGOPATH
	//config.StrTestName = *pTestName
	config.StrOutputFullPath = *pOutputFullPath
	config.BoolGlobalTuple = *pModeGlobalTuple

	/* TODO:: Not finished in this part!!!
	In this part, we should implement an algorithm that can iterate all the possible unit test inside our target program.
	Right now, we are using an ad-hoc pre-defined slice pTestnameList for the DEMO purpose.
	 */
	var pTestNameList []string
	pTestNameList = append(pTestNameList, "TestF1")

	/* Standard initialize the necessary data structure. */
	var fuzzingQueue []fuzzer.FuzzQueryEntry
	var mainRecord = fuzzer.EmptyRecord()
	var allRecordHashSlice []string

	for _, pTestName := range pTestNameList {
		//  Run the instrumented program for one time,
		//  generate original input and record
		var emptyInput = fuzzer.EmptyInput()
		emptyInput.TestName = pTestName
		retInput, _ := fuzzer.Run(emptyInput)

		/* Generate all possible deter_inputs based on the current retInput. Only changing one select per time. */
		deterInputSlice := fuzzer.Deterministic_enumerate_input(retInput)

		/* Now we deterministically enumerate one select per time, iterating through all the pTestName and all selects */
		for _, deterInput := range deterInputSlice {
			retInput, retRecord := fuzzer.Run(deterInput)
			if len(retInput.VecSelect) == 0 {   // TODO:: Should we ignore the output that contains no VecSelects entry?
				// TODO:: In the toy project, we notice that the retInput.VecSelect is nil.
				continue
			}
			recordHash := fuzzer.HashOfRecord(retRecord)
			/* See whether the current deter_input trigger a new record. If yes, save the record hash and the input to the queue. */
			if fuzzer.FindRecordHashInSlice(recordHash, allRecordHashSlice) == false {
				curScore := fuzzer.ComputeScore(mainRecord, retRecord)
				currentFuzzEntry := fuzzer.FuzzQueryEntry{
					IsFavored:              false,
					ExecutionCount:         1,
					BestScore:              curScore,
					CurrentInput:           retInput,
					CurrentRecordHashSlice: []string{recordHash},
				}
				fuzzingQueue = append(fuzzingQueue, currentFuzzEntry)
				allRecordHashSlice = append(allRecordHashSlice, recordHash)
			} else {
				continue
			}
		}
	}

	/* Infinite loop for the main fuzzing */
	mainRandomLoopIdx := 0
	for {
		fmt.Println("Beginning main random fuzzing loop idx: ", mainRandomLoopIdx)
		if len(fuzzingQueue) == 0 {
			fmt.Println("Fuzzing Queue is nil (no components). Some error occurs. ")
			break
		}
		for _, currentEntry := range fuzzingQueue{
			// TODO:: Maybe we should cull the queue first. (Or maybe after the calibration?)

			/* Calibrate case. */
			for i := 0; i < 1; i++ {   // TODO:: Maybe we can have multiple times of Calibrate case here. Since the retRecord might not be completely stable.
				// TODO:: There seems to be no way to get an error message from the Run func?
				// TODO:: Set calibration_failed to the queue entry if calibration failed (fuzz.Run() failed)
				retInput, retRecord := fuzzer.Run(currentEntry.CurrentInput)
				if len(retInput.VecSelect) == 0 {   // TODO:: Should we ignore the output that contains no VecSelects entry?
					continue
				}
				recordHash := fuzzer.HashOfRecord(retRecord)
				if fuzzer.FindRecordHashInSlice(recordHash, currentEntry.CurrentRecordHashSlice) == false {
					currentEntry.CurrentRecordHashSlice = append(currentEntry.CurrentRecordHashSlice, recordHash)
				}
				if fuzzer.FindRecordHashInSlice(recordHash, allRecordHashSlice) == false {
					allRecordHashSlice = append(allRecordHashSlice, recordHash)
				}
				curScore := fuzzer.ComputeScore(mainRecord, retRecord)
				if curScore > currentEntry.BestScore {
					currentEntry.BestScore = curScore
				}
			}

			/* Next, Random fuzzing */
			// TODO:: Random fuzzing with dynamic energy. (Maybe depends on the scores, executionCount etc)
			currentFuzzingEnergy := 100
			for randFuzzIdx := 0; randFuzzIdx < currentFuzzingEnergy; randFuzzIdx++ {
				currentMutatedInput := fuzzer.Random_Mutate_Input(currentEntry.CurrentInput)
				retInput, retRecord := fuzzer.Run(currentMutatedInput)
				recordHash := fuzzer.HashOfRecord(retRecord)
				if fuzzer.FindRecordHashInSlice(recordHash, allRecordHashSlice) == false {   // Found a new input with unique record!!!
					curScore := fuzzer.ComputeScore(mainRecord, retRecord)
					currentFuzzEntry := fuzzer.FuzzQueryEntry{
						IsFavored:              false,
						ExecutionCount:         1,
						BestScore:              curScore,
						CurrentInput:           retInput,
						CurrentRecordHashSlice: []string{recordHash},
					}
					fuzzingQueue = append(fuzzingQueue, currentFuzzEntry)
					allRecordHashSlice = append(allRecordHashSlice, recordHash)
				} else {continue}  // This mutation does not create new record. Discarded.
				currentEntry.ExecutionCount += 1
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

