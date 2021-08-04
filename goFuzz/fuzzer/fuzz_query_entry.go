package fuzzer

import (
	"fmt"
	"log"
	"math/rand"
)

var (
	test2HighScore = make(map[*GoTest]int)
	cusTest2HighScore = make(map[string]int)
)

// FuzzQueryEntry records the multiple run results for an input and history score
// Notes:
//   1. An input can be run multiple times
//   2. An input will be randomly mutated to run
type FuzzQueryEntry struct {
	Idx                 uint64
	Stage               FuzzStage
	IsFavored           bool
	BestScore           int
	ExecutionCount      int
	CurrInput           *Input
	CurrRecordHashSlice []string
}

func (e *FuzzQueryEntry) String() string {
	return fmt.Sprintf("[input %s][stage %s][idx %d]", e.CurrInput, e.Stage, e.Idx)
}

func NewInitStageFuzzQueryEntryWithGoTest(test *GoTest) *FuzzQueryEntry {
	return &FuzzQueryEntry{
		Stage: InitStage,
		CurrInput: &Input{
			Note:      NotePrintInput,
			GoTestCmd: test,
		},
	}
}

func NewInitStageFuzzQueryEntryWithCustomCmd(customCmd string) *FuzzQueryEntry {
	return &FuzzQueryEntry{
		Stage: InitStage,
		CurrInput: &Input{
			Note:      NotePrintInput,
			CustomCmd: customCmd,
		},
	}
}

// HandleFuzzQueryEntry will handle a single entry from fuzzCtx's fuzzingQueue
// Notes:
//   1. e is expected to be dequeue from fuzzCtx's fuzzingQueue
func HandleFuzzQueryEntry(e *FuzzQueryEntry, fuzzCtx *FuzzContext) error {
	// TODO: better way to print FuzzQueryEntry, maybe ID or string of input?
	log.Printf("handle entry: %s\n", e)

	if shouldDropFuzzQueryEntry(fuzzCtx, e) {
		return nil
	}

	var runTasks []*RunTask

	if e.Stage == InitStage {
		// If stage is InitStage, input's note will be PrintInput and gooracle will record select choices
		t, err := NewRunTask(e.CurrInput, e.Stage, e.Idx, e.ExecutionCount, e)
		if err != nil {
			return err
		}
		runTasks = append(runTasks, t)
	} else if e.Stage == DeterStage {
		// If stage is InitStage, input's note will be not PrintInput and expect to have some select choice enforcement
		t, err := NewRunTask(e.CurrInput, e.Stage, e.Idx, e.ExecutionCount, e)
		if err != nil {
			return err
		}
		runTasks = append(runTasks, t)
	} else if e.Stage == CalibStage {
		t, err := NewRunTask(e.CurrInput, e.Stage, e.Idx, e.ExecutionCount, e)
		if err != nil {
			return err
		}
		runTasks = append(runTasks, t)
	} else if e.Stage == RandStage {
		/* Determine whether we should skip the current entry based on score. */
		isSkip := false
		if e.CurrInput.GoTestCmd != nil { // This is Go Test command.
			/* If we have seen this unit test before. */
			if prevHighestScore, ok := test2HighScore[e.CurrInput.GoTestCmd]; ok {
				if prevHighestScore <= e.BestScore {
					test2HighScore[e.CurrInput.GoTestCmd] = e.BestScore
				} else {
					/* Current score smaller than the highest score, skip by chances. */
					randNum := rand.Int31n(101)
					skipPos := (prevHighestScore - e.BestScore) / prevHighestScore
					if randNum < int32(skipPos) {
						isSkip = true
					}
				}
			} else {
				/* If we haven't seen this unit test before. */
				test2HighScore[e.CurrInput.GoTestCmd] = e.BestScore
			}
		} else if e.CurrInput.CustomCmd != "" { // This is custom test command. Similar to the previous case.
			if prevHighestScore, ok := cusTest2HighScore[e.CurrInput.CustomCmd]; ok {
				/* If we have seen this unit test before. */
				if prevHighestScore <= e.BestScore {
					cusTest2HighScore[e.CurrInput.CustomCmd] = e.BestScore
				} else {
					/* Current score smaller than the highest score, skip by chances. */
					randNum := rand.Int31n(101)
					skipPos := (prevHighestScore - e.BestScore) / prevHighestScore
					if randNum < int32(skipPos) {
						isSkip = true
					}
				}
			} else {
				/* If we haven't seen this unit test before. */
				cusTest2HighScore[e.CurrInput.CustomCmd] = e.BestScore
			}
		}
		if isSkip == true {
			log.Printf("[%s] randomly skipped", e)
			// if skip, simply add entry to the tail
			// TODO:: Should we add the skipped query to the tail??? s
			fuzzCtx.EnqueueQueryEntry(e)
			return nil
		}
		// energy is too large
		currentFuzzingEnergy := (e.BestScore / 10) + 1
		generatedSelectsHash := make(map[string]bool)
		execCount := e.ExecutionCount
		log.Printf("[%+v] randomly mutate with energy %d", *e, currentFuzzingEnergy)
		for randFuzzIdx := 0; randFuzzIdx < currentFuzzingEnergy; randFuzzIdx++ {
			randomInput, err := RandomMutateInput(e.CurrInput)
			if err != nil {
				log.Printf("[%s] randomly mutate input fail: %s, continue", e, err)
				continue
			}
			selectsHash := GetHashOfSelects(randomInput.VecSelect)
			if _, exist := generatedSelectsHash[selectsHash]; exist {
				log.Printf("[%s][%d] skip generated input because of duplication", e, randFuzzIdx)
				continue
			}
			generatedSelectsHash[selectsHash] = true
			log.Printf("[%s][%d] successfully generate input", e, randFuzzIdx)
			t, err := NewRunTask(randomInput, e.Stage, e.Idx, execCount, e)
			if err != nil {
				return err
			}
			runTasks = append(runTasks, t)
			execCount += 1
		}
		e.ExecutionCount = execCount
		fuzzCtx.EnqueueQueryEntry(e)
	} else {
		return fmt.Errorf("incorrect stage found: %s", e.Stage)
	}

	for _, t := range runTasks {
		fuzzCtx.runTaskCh <- t
	}

	return nil

}

// shouldDropFuzzQueryEntry return true if given fuzz entry need to be dropped
func shouldDropFuzzQueryEntry(fuzzCtx *FuzzContext, e *FuzzQueryEntry) bool {
	// only check when it is in rand stage
	if e.Stage != RandStage {
		return false
	}
	return ShouldSkipInput(fuzzCtx, e.CurrInput)
}
