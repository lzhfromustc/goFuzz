package fuzzer

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
)

var (
	recordHashMapLock sync.RWMutex
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
	if runTask.entry != nil && runTask.entry.PrevID != "" {
		log.Printf("[Worker %s][Task %s][PrevID %s] handling result", workerID, runTask.id, runTask.entry.PrevID)
	} else {
		log.Printf("[Worker %s][Task %s][PrevID %s] handling result", workerID, runTask.id, runTask.entry.PrevID)
	}
	src := runTask.input.Src()

	if result.IsTimeout {
		log.Printf( "[Worker %s][Task %s] found timeout", workerID, runTask.id)
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
		cov, combCov, err := GetCumulativeTriggeredCaseCoverage(src)
		if err != nil {
			log.Printf("[Worker %s][Task %s][ignored] GetCumulativeTriggeredCaseCoverage failed: %v", workerID, runTask.id, err)
		} else {
			log.Printf("[Worker %s][Task %s] cumulative case coverage: %.2f%%, case combination coverage %.2f%%", workerID, runTask.id, cov*100, combCov*100)
			fuzzCtx.UpdateTargetMaxCaseCov(src, cov)
		}
	}

	// echo primitive operation coverage if it has
	if OpCover != "" {
		// calculate and print how many operation happened in the single run
		report := GetCurrOpIDCoverageReport(result.opIDs)
		PrintCurrOpIDCovReport(totalReport, report)

		// calculate and print how many operation happened in the total
		UpdateTriggeredOpID(result.opIDs)
		cumulativeReport := GetTriggeredOpIDCoverageReport()
		PrintTriggeredOpIDCovReport(totalReport, cumulativeReport)
	}

	log.Printf("[Worker %s][Task %s] has %d bug(s), %d unique bug(s)", workerID, runTask.id, len(result.BugIDs), numOfBugs)

	// Increase queue entry execution count.
	runTask.entry.ExecutionCount++

	if stage == InitStage {
		// If we are handling the output from InitStage
		// Generate all possible deter_inputs based on the current retInput. Only changing one select per time
		if result.RetInput == nil {
			return fmt.Errorf("input should not be empty")
		}

		var log2RetRecord = retRecord
		for key, element := range log2RetRecord.MapTupleRecord{
			log2element := int(math.Log2(float64(element)))
			if log2element > 0 {
				log2RetRecord.MapTupleRecord[key] = log2element
			} else {
				log2RetRecord.MapTupleRecord[key] = 0
			}
		}

		ComputeScore(fuzzCtx.mainRecord, retRecord, result, runTask.id, "")

		deterInputs := Deterministic_enumerate_input(result.RetInput)
		log.Printf("[Worker %s][Task %s] generated %d inputs for deter stage", workerID, runTask.id, len(deterInputs))

		for _, deterInput := range deterInputs {

			// Create multiple entries based on deterministic enumeration
			err := fuzzCtx.EnqueueQueryEntry(&FuzzQueryEntry{
				Stage:     DeterStage,
				CurrInput: deterInput,
				Idx:       fuzzCtx.NewFuzzQueryEntryIndex(),
				PrevID:    runTask.id,
				IsFavored: false,
				ExecutionCount: 1,
				Depth:	   1,
			})
			if err != nil {
				log.Panicln(err)
			}
		}
		/* We only want to apply InitStage once. Next time we see this in the queue, treat it as RandStage */
		runTask.entry.Stage = RandStage

	} else if stage == DeterStage {
		if retRecord != nil {
			// If we are handling the output from DeterStage
			var log2RetRecord = retRecord
			for key, element := range log2RetRecord.MapTupleRecord{
				log2element := int(math.Log2(float64(element)))
				if log2element > 0 {
					log2RetRecord.MapTupleRecord[key] = log2element
				} else {
					log2RetRecord.MapTupleRecord[key] = 0
				}
			}
			recordHash := HashOfRecord(log2RetRecord)
			/* See whether the current deter_input trigger a new record. If yes, save the record hash and the input to the queue. */
			recordHashMapLock.Lock()
			if _, exist := fuzzCtx.allRecordHashMap[recordHash]; !exist {
				if runTask.entry != nil && runTask.entry.PrevID != "" {
					ComputeScore(fuzzCtx.mainRecord, retRecord, result, runTask.id, runTask.entry.PrevID)
				} else {
					log.Printf("Error, cannot find previous ID in Deter. ")
				}
				currentFuzzEntry := &FuzzQueryEntry{
					IsFavored:           false,
					ExecutionCount:      1,
					BestScore:           0,
					CurrInput:           runTask.input,
					CurrRecordHashSlice: []string{recordHash},
					Stage:               CalibStage,
					Idx:                 fuzzCtx.NewFuzzQueryEntryIndex(),
					PrevID:              runTask.id,
					Depth:               runTask.entry.Depth + 1,
				}

				// After running at DeterStage, move to CalibStage to run more times
				fuzzCtx.EnqueueQueryEntry(currentFuzzEntry)
				fuzzCtx.allRecordHashMap[recordHash] = struct{}{}
			}
			recordHashMapLock.Unlock()
		}

		/* We only want to apply DeterStage once. Next time we see this in the queue, treat it as RandStage */
		runTask.entry.Stage = RandStage

	} else if stage == CalibStage {
		if retRecord != nil {
			// If we are handling the output from CalibStage
			var log2RetRecord = retRecord
			for key, element := range log2RetRecord.MapTupleRecord{
				log2element := int(math.Log2(float64(element)))
				if log2element > 0 {
					log2RetRecord.MapTupleRecord[key] = log2element
				} else {
					log2RetRecord.MapTupleRecord[key] = 0
				}
			}
			recordHash := HashOfRecord(log2RetRecord)
			currentEntry := &FuzzQueryEntry{
				IsFavored:           false,
				ExecutionCount:      1,
				CurrInput:           runTask.input,
				Stage:               RandStage,
				Idx:                 fuzzCtx.NewFuzzQueryEntryIndex(),
				PrevID:              runTask.id,
				Depth:               runTask.entry.Depth + 1,
			}
			if !FindRecordHashInSlice(recordHash, currentEntry.CurrRecordHashSlice) {
				currentEntry.CurrRecordHashSlice = append(currentEntry.CurrRecordHashSlice, recordHash)
			}

			recordHashMapLock.Lock()
			if _, exist := fuzzCtx.allRecordHashMap[recordHash]; !exist {
				fuzzCtx.allRecordHashMap[recordHash] = struct{}{}
			}
			recordHashMapLock.Unlock()

			curScore := 0
			if runTask.entry != nil && runTask.entry.PrevID != "" {
				curScore = ComputeScore(fuzzCtx.mainRecord, retRecord, result, runTask.id, runTask.entry.PrevID)
			} else {
				log.Printf("Error, cannot find previous ID in Calib. ")
			}
			if curScore > currentEntry.BestScore {
				currentEntry.BestScore = curScore
			}

			// After calibration, we can move stage to RandStage
			fuzzCtx.EnqueueQueryEntry(currentEntry)
		}

		/* We only want to apply CalibStage once. Next time we see this in the queue, treat it as RandStage */
		runTask.entry.Stage = RandStage

	} else if stage == RandStage {
	//if stage == InitStage || stage == RandStage || stage == DeterStage || stage == CalibStage {
		var input *Input
		// If we are handling the output from RandStage
		if stage == InitStage && result.RetInput == nil {
			return fmt.Errorf("input should not be empty")
		}
		if stage == InitStage {
			input = result.RetInput
		} else {
			input = runTask.input
		}


		if retRecord != nil {
			var log2RetRecord = retRecord
			for key, element := range log2RetRecord.MapTupleRecord{
				log2element := int(math.Log2(float64(element)))
				if log2element > 0 {
					log2RetRecord.MapTupleRecord[key] = log2element
				} else {
					log2RetRecord.MapTupleRecord[key] = 0
				}
			}
			recordHash := HashOfRecord(log2RetRecord)
			recordHashMapLock.Lock()
			if _, exist := fuzzerContext.allRecordHashMap[recordHash]; !exist { // Found a new input with unique record!!!
				curScore := 0
				if runTask.entry != nil && runTask.entry.PrevID != "" {
					ComputeScore(fuzzCtx.mainRecord, retRecord, result, runTask.id, runTask.entry.PrevID)
				} else {
					log.Printf("Error, cannot find previous ID in Rand. ")
				}
				newEntry := &FuzzQueryEntry{
					IsFavored:           false,
					ExecutionCount:      1,
					BestScore:           curScore,
					CurrInput:           input,
					CurrRecordHashSlice: []string{recordHash},
					Stage:               RandStage,
					Idx:                 fuzzCtx.NewFuzzQueryEntryIndex(),
					PrevID:              runTask.id,
					Depth:               runTask.entry.Depth + 1,
				}
				fuzzCtx.EnqueueQueryEntry(newEntry)
				fuzzerContext.allRecordHashMap[recordHash] = struct{}{}
			}
			recordHashMapLock.Unlock()
		}
	} else {
		return fmt.Errorf("[Worker %s] found incorrect stage [%s]", workerID, stage)
	}

	return nil
}
