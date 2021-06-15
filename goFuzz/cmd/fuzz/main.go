package main

import (
	"flag"
	"goFuzz/fuzzer"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	workerInputChan  chan *fuzzer.Input = make(chan *fuzzer.Input)
	fuzzingQueue                        = []fuzzer.FuzzQueryEntry{}
	mainRecord                          = fuzzer.EmptyRecord()
	allRecordHashMap                    = make(map[string]struct{})
	deteredTestname                     = make(map[string]bool)
)

// init setups the log for the fuzzer
func parseFlag() {
	file, err := os.OpenFile("fuzzer.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	w := io.MultiWriter(file, os.Stdout)
	log.SetOutput(w)

	// Parse input
	pTargetGoModDir := flag.String("goModDir", "", "Directory contains Go Mod file")
	pTargetTestFunc := flag.String("testFunc", "", "Optional, if you only want to test single function in unit test")
	pOutputDir := flag.String("outputDir", "", "Full path of the output file")
	pModeGlobalTuple := flag.Bool("globalTuple", false, "Whether prev_location is global or per channel")
	maxParallel := flag.Int("parallel", 1, "Specified the maximum subroutine number for fuzzing.")

	flag.Parse()

	fuzzer.TargetTestFunc = *pTargetTestFunc
	fuzzer.OutputDir = *pOutputDir
	fuzzer.TargetGoModDir = *pTargetGoModDir
	fuzzer.GlobalTuple = *pModeGlobalTuple
	fuzzer.MaxParallel = *maxParallel

	if fuzzer.OutputDir == "" {
		log.Fatal("-outputDir is required")
	}

	if fuzzer.TargetGoModDir == "" {
		log.Fatal("-goModDir is required")
	}
}

// startWorkers starts maxParallel workers working on workChan.
func startWorkers(maxParallel int, workChan chan *fuzzer.Input) {
	var wg sync.WaitGroup

	for i := 0; i < maxParallel; i++ {
		wg.Add(1)

		// Start worker
		go func(i int) {
			log.Printf("[Worker %d] Started", i)
			defer wg.Done()
			for {
				select {
				// Receive input
				case input := <-workChan:
					log.Printf("[Worker %d] Working on %s\n", i, input.TestName)
					retOutput, err := fuzzer.Run(input)
					if err != nil {
						log.Printf("[Worker %d] [Test %s] Error: %s\n", i, input.TestName, err.Error())
						continue
					}
					// Handle output
					handleRetOutput(retOutput)
				case <-time.After(60 * time.Second):
					log.Printf("[Worker %d] Timeout. Exiting.", i)
					break
				}
			}

		}(i)
	}

	wg.Wait()
}

func main() {
	parseFlag()

	var err error
	fuzzer.SetDeadline()

	/* Begin running specific number of worker subroutines. Blocked before we send them inputs from the main routine. */
	go startWorkers(fuzzer.MaxParallel, workerInputChan)

	var tests []string
	if fuzzer.TargetTestFunc != "" {
		tests = append(tests, fuzzer.TargetTestFunc)
	} else {
		tests, err = fuzzer.ListTestsInPackage(fuzzer.TargetGoModDir, "./...")
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Tests going to be run: %s", tests)

	for _, pTestName := range tests {
		//  Run the instrumented program for one time,
		//  generate original input and record
		var emptyInput = fuzzer.EmptyInput()
		emptyInput.TestName = pTestName
		// Give one job to the worker. And receive results.
		workerInputChan <- emptyInput
	}

	/* Infinite loop for the main fuzzing */
	mainRandomLoopIdx := 0
	for {
		log.Println("Beginning main random fuzzing loop idx: ", mainRandomLoopIdx)
		if len(fuzzingQueue) == 0 {
			log.Println("Fuzzing queue is empty, waiting 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		// TODO:: Maybe we should cull the queue first. (Or maybe after the calibration?)
		// TODO:: Prioritize queue based on scores.
		fuzzingQueueLen := len(fuzzingQueue)
		for fuzzingIdx := 0; fuzzingIdx < fuzzingQueueLen; fuzzingIdx++ {
			currentEntry := fuzzingQueue[fuzzingIdx]
			/* Calibrate case. */
			for i := 0; i < 1; i++ {
				// TODO:: Maybe we can have multiple times of Calibrate case here. Since the retRecord might not be completely stable.
				// TODO:: If the seed case has already been calibrated, maybe we can skip the duplicated calibrate case.
				// TODO:: There seems to be no way to get an error message from the Run func?
				// TODO:: Set calibration_failed to the queue entry if calibration failed (fuzz.Run() failed)
				currentEntry.CurrentInput.Stage = "calib"
				workerInputChan <- currentEntry.CurrentInput

				/* Next, Random fuzzing */
				// TODO:: Random fuzzing with dynamic energy. (Maybe depends on the scores, executionCount etc)
				currentFuzzingEnergy := 100
				for randFuzzIdx := 0; randFuzzIdx < currentFuzzingEnergy; randFuzzIdx++ {
					currentMutatedInput := fuzzer.Random_Mutate_Input(currentEntry.CurrentInput)
					currentMutatedInput.Stage = "rand"
					workerInputChan <- currentMutatedInput
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
	}
}
