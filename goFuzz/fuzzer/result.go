package fuzzer

import (
	"fmt"
	"log"
	"strings"
)

type RunResult struct {
	RetInput       *Input
	RetRecord      *Record
	StdoutFilepath string
	BugIDs         []string
	// Executed operation ID
	opIDs []string
}

func CheckBugFromStdout(content string) (numBug int) {

	text := strings.Split(content, "\n")

	for _, oneLine := range text {
		if oneLine == "-----New Blocking Bug:" {
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

	// Check any unique bugs found
	numOfBugs := 0
	for _, bugID := range result.BugIDs {
		if !fuzzCtx.HasBugID(bugID) {
			fuzzCtx.AddBugID(bugID, result.StdoutFilepath)
			numOfBugs += 1
		}
	}

	if numOfBugs != 0 {
		fuzzCtx.IncNumOfBugsFound(uint64(numOfBugs))
	}

	// record & print case coverage

	// If init stage, initailize track with total case combination
	if stage == InitStage {

		err := RecordTotalCases(testID2cases, runTask.src, result.RetInput.VecSelect)
		if err != nil {
			log.Printf("[Task %s][ignored] RecordTotalCases failed: %v", runTask.id, err)
		}

	}

	var triggeredSelects []SelectInput
	if stage == InitStage {
		triggeredSelects = result.RetInput.VecSelect
	} else {
		triggeredSelects = runTask.input.VecSelect
	}

	err := RecordTriggeredCase(testID2cases, runTask.src, triggeredSelects)
	if err != nil {
		log.Printf("[Task %s][ignored] RecordTriggeredCase failed: %v", runTask.id, err)
	} else {
		cov, err := GetCumulativeTriggeredCaseCoverage(testID2cases, runTask.src)
		if err != nil {
			log.Printf("[Task %s][ignored] GetCumulativeTriggeredCaseCoverage failed: %v", runTask.id, err)
		} else {
			log.Printf("[Task %s] cumulative case coverage: %.2f%%", runTask.id, cov*100)
		}
	}

	// echo primitive operation coverage if it has
	if OpCover != "" {
		// calculate and print how many operation happened in the single run
		report := GetCurrOpIDCoverageReport(opID2Type, result.opIDs)
		PrintCurrOpIDCovReport(totalReport, report)

		// calculate and print how many operation happened in the total
		UpdateTriggeredOpID(triggeredOpID, result.opIDs)
		cumulativeReport := GetTriggeredOpIDCoverageReport(opID2Type, triggeredOpID)
		PrintTriggeredOpIDCovReport(totalReport, cumulativeReport)
	}

	log.Printf("[Task %s] has %d bug(s), %d unique bug(s)", runTask.id, len(result.BugIDs), numOfBugs)

	if stage == InitStage {
		// If we are handling the output from InitStage
		// Generate all possible deter_inputs based on the current retInput. Only changing one select per time
		if result.RetInput == nil {
			return fmt.Errorf("input should not be empty")
		}

		deterInputs := Deterministic_enumerate_input(result.RetInput)
		log.Printf("[Task %s] generated %d inputs by deterministic enumeration", runTask.id, len(deterInputs))

		for _, deterInput := range deterInputs {

			// Create multiple entries based on deterministic enumeration
			err := fuzzCtx.EnqueueQueryEntry(&FuzzQueryEntry{
				Stage:     DeterStage,
				CurrInput: deterInput,
				Idx:       fuzzCtx.NewFuzzQueryEntryIndex(),
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
		if _, exist := fuzzCtx.allRecordHashMap[recordHash]; !exist {
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
		if !FindRecordHashInSlice(recordHash, currentEntry.CurrRecordHashSlice) {
			currentEntry.CurrRecordHashSlice = append(currentEntry.CurrRecordHashSlice, recordHash)
		}

		if _, exist := fuzzCtx.allRecordHashMap[recordHash]; !exist {
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
		if _, exist := fuzzerContext.allRecordHashMap[recordHash]; !exist { // Found a new input with unique record!!!
			curScore := ComputeScore(fuzzerContext.mainRecord, retRecord)
			newEntry := &FuzzQueryEntry{
				IsFavored:           false,
				ExecutionCount:      1,
				BestScore:           curScore,
				CurrInput:           runTask.input,
				CurrRecordHashSlice: []string{recordHash},
				Stage:               RandStage,
				Idx:                 fuzzCtx.NewFuzzQueryEntryIndex(),
			}
			fuzzCtx.EnqueueQueryEntry(newEntry)
			fuzzerContext.allRecordHashMap[recordHash] = struct{}{}
		}
	} else {
		return fmt.Errorf("found incorrect stage [%s]", stage)
	}

	return nil
}
