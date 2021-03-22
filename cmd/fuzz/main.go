package main

import (
	"flag"
	"goFuzz/config"
	"goFuzz/fuzzer"
)

func main() {

	fuzzer.SetDeadline()

	// Parse input
	pProjectPath := flag.String("path","","Full path of the target project")
	pProjectGOPATH := flag.String("GOPATH","","GOPATH of the target project")
	pTestName := flag.String("test","","Function name of the unit test")

	flag.Parse()

	config.StrTestPath = *pProjectPath
	config.StrProjectGOPATH = *pProjectGOPATH
	config.StrTestName = *pTestName

	// Run the instrumented program for one time,
	//  generate original input and record
	emptyInput := fuzzer.EmptyInput()
	input, mainRecord := fuzzer.Run(config.StrTestName, emptyInput)

	mapExecutedHash := make(map[string]struct{})

	workList := []fuzzer.Input{input}
	workListScore := []int{0}


	// fuzzing loop
	for len(workList) > 0 {
		// Pop the first element of worklist
		curInput, _ := fuzzer.PopWorklist(&workList)
		workListScore = workListScore[1:]

		// Check if input has been executed
		hash := fuzzer.HashOfInput(curInput)
		if _, executed := mapExecutedHash[hash]; executed {
			continue
		}
		mapExecutedHash[hash] = struct{}{}

		// Run new input
		_, curRecord := fuzzer.Run(config.StrTestName, curInput)
		curScore := fuzzer.ComputeScore(mainRecord, curRecord)
		mainRecord = fuzzer.UpdateMainRecord(mainRecord, curRecord)

		// Based on current input, generate multiple new inputs
		// and put them into the workList by curScore
		newInputs := fuzzer.GenInputs(curInput)

		if len(workListScore) == 0 {
			workList = newInputs
			workListScore = []int{}
			for i := 0; i < len(workList); i++ {
				workListScore = append(workListScore, curScore)
			}
		} else {
			var indexInsertAfter int // insert the new inputs after this index. -1 stands for insert before the head
			for i := -1; i < len(workListScore); i++ {
				if i == -1 {
					if curScore >= workListScore[0] {
						indexInsertAfter = i
					}
				} else if i == len(workListScore) - 1 {
					if curScore <= workListScore[len(workListScore) - 1] {
						indexInsertAfter = i
					}
				} else {
					if workListScore[i] >= curScore && workListScore[i + 1] <= curScore {
						indexInsertAfter = i
					}
				}
			}
			workList = fuzzer.InsertWorklist(newInputs, workList, indexInsertAfter)
			workListScore = fuzzer.InsertWorklistScore(curScore, len(newInputs), workListScore, indexInsertAfter)
		}
	}

	return

}

