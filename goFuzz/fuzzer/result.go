package fuzzer

import (
	"fmt"
	"log"
	"strings"
)

type RunResult struct {
	RetInput  *Input
	RetRecord *Record
}

func NewRunResult() *RunResult {
	return &RunResult{
		RetInput:  nil,
		RetRecord: nil,
	}
}

func CheckBugFromStdout(content string) (numBug int) {

	text := strings.Split(content, "\n")

	for _, oneLine := range text {
		if oneLine == "-----New Bug:" {
			numBug++
		}
	}

	return
}

// HandleRunResult handle the result for the given runTask
func HandleRunResult(runTask *RunTask, result *RunResult, fuzzCtx *FuzzContext) error {
	log.Printf("[Task %s] handling result", runTask.id)
	retRecord := result.RetRecord
	stage := runTask.stage

	if stage == InitStage {
		// If we are handling the output from InitStage
		// Generate all possible deter_inputs based on the current retInput. Only changing one select per time
		if result.RetInput == nil {
			return fmt.Errorf("input should not be empty")
		}

		deterInputs := Deterministic_enumerate_input(result.RetInput)
		log.Printf("[Task %s] generated %d inputs by deterministic enumeration", runTask.id, len(deterInputs))

		for idx, deterInput := range deterInputs {

			// Create multiple entries based on deterministic enumeration
			err := fuzzCtx.EnqueueQueryEntry(&FuzzQueryEntry{
				Stage:     DeterStage,
				CurrInput: deterInput,
				Idx:       idx,
			})
			if err != nil {
				log.Panicln(err)
			}
		}
	} else if stage == DeterStage {
		// If we are handling the output from DeterStage
		recordHash := HashOfRecord(retRecord)
		currentFuzzEntry := runTask.entry
		/* See whether the current deter_input trigger a new record. If yes, save the record hash and the input to the queue. */
		if _, exist := fuzzCtx.allRecordHashMap[recordHash]; exist == false {
			curScore := ComputeScore(fuzzCtx.mainRecord, retRecord)
			currentFuzzEntry.ExecutionCount = 1
			currentFuzzEntry.BestScore = curScore
			currentFuzzEntry.CurrInput = runTask.input
			currentFuzzEntry.CurrRecordHashSlice = []string{recordHash}

			// After running at DeterStage, move to CalibStage to run more times
			currentFuzzEntry.Stage = CalibStage
			fuzzCtx.EnqueueQueryEntry(currentFuzzEntry)
			fuzzCtx.allRecordHashMap[recordHash] = struct{}{}
		}
	} else if stage == CalibStage {
		// If we are handling the output from CalibStage
		recordHash := HashOfRecord(retRecord)
		currentEntry := runTask.entry
		if FindRecordHashInSlice(recordHash, currentEntry.CurrRecordHashSlice) == false {
			currentEntry.CurrRecordHashSlice = append(currentEntry.CurrRecordHashSlice, recordHash)
		}

		if _, exist := fuzzCtx.allRecordHashMap[recordHash]; exist == false {
			fuzzCtx.allRecordHashMap[recordHash] = struct{}{}
		}
		curScore := ComputeScore(fuzzCtx.mainRecord, retRecord)
		if curScore > currentEntry.BestScore {
			currentEntry.BestScore = curScore
		}

		// After calibration, we can move stage to RandStage
		currentEntry.Stage = RandStage
		currentEntry.ExecutionCount += 1
		fuzzCtx.EnqueueQueryEntry(currentEntry)

	} else if stage == RandStage {
		// If we are handling the output from RandStage
		recordHash := HashOfRecord(retRecord)
		if _, exist := fuzzerContext.allRecordHashMap[recordHash]; exist == false { // Found a new input with unique record!!!
			curScore := ComputeScore(fuzzerContext.mainRecord, retRecord)
			newEntry := &FuzzQueryEntry{
				IsFavored:           false,
				ExecutionCount:      1,
				BestScore:           curScore,
				CurrInput:           runTask.input,
				CurrRecordHashSlice: []string{recordHash},
				Stage:               RandStage,
				Idx:                 0,
			}
			fuzzCtx.EnqueueQueryEntry(newEntry)
			fuzzerContext.allRecordHashMap[recordHash] = struct{}{}
		}
	} else {
		return fmt.Errorf("found incorrect stage [%s]", stage)
	}

	return nil
}
