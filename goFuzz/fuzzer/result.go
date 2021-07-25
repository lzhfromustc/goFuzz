package fuzzer

import (
	"context"
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
	opIDs     []string
	IsTimeout bool
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
func HandleRunResult(ctx context.Context, runTask *RunTask, result *RunResult, fuzzCtx *FuzzContext) error {
	workerID := ctx.Value("workerID").(string)

	log.Printf("[Worker %s][Task %s] handling result", workerID, runTask.id)
	src := runTask.input.Src()

	if result.IsTimeout {
		fuzzCtx.RecordTargetTimeoutOnce(src)
	}

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

	fuzzCtx.UpdateTargetStage(src, stage)

	// record & print case coverage
	// If init stage, initailize track with total case combination
	if stage == InitStage {

		err := RecordTotalCases(src, result.RetInput.VecSelect)
		if err != nil {
			log.Printf("[Worker %s][Task %s][ignored] RecordTotalCases failed: %v", workerID, runTask.id, err)
		}

	}

	if retRecord != nil {
		log.Printf("[Worker %s][Task %s] number of tuple: %d", workerID, runTask.id, len(result.RetRecord.MapTupleRecord))
	}

	var triggeredSelects []SelectInput
	if stage == InitStage {
		triggeredSelects = result.RetInput.VecSelect
	} else {
		triggeredSelects = runTask.input.VecSelect
	}

	err := RecordTriggeredCase(src, triggeredSelects)
	if err != nil {
		log.Printf("[Worker %s][Task %s][ignored] RecordTriggeredCase failed: %v", workerID, runTask.id, err)
	} else {
		cov, err := GetCumulativeTriggeredCaseCoverage(src)
		if err != nil {
			log.Printf("[Worker %s][Task %s][ignored] GetCumulativeTriggeredCaseCoverage failed: %v", workerID, runTask.id, err)
		} else {
			log.Printf("[Worker %s][Task %s] cumulative case coverage: %.2f%%", workerID, runTask.id, cov*100)
			fuzzCtx.UpdateTargetMaxCaseCov(src, cov)
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

	log.Printf("[Worker %s][Task %s] has %d bug(s), %d unique bug(s)", workerID, runTask.id, len(result.BugIDs), numOfBugs)

	if stage == InitStage {
		// If we are handling the output from InitStage
		// Generate all possible deter_inputs based on the current retInput. Only changing one select per time
		if result.RetInput == nil {
			return fmt.Errorf("input should not be empty")
		}

		deterInputs := Deterministic_enumerate_input(result.RetInput)
		log.Printf("[Worker %s][Task %s] generated %d inputs for deter stage", workerID, runTask.id, len(deterInputs))

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
		if retRecord != nil {
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
		}

	} else if stage == CalibStage {
		if retRecord != nil {
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
		}

	} else if stage == RandStage {
		if retRecord != nil {
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
		}
	} else {
		return fmt.Errorf("[Worker %s] found incorrect stage [%s]", workerID, stage)
	}

	return nil
}
