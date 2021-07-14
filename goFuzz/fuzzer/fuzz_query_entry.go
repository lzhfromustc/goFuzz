package fuzzer

import (
	"fmt"
	"log"
	"math/rand"
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
	IsCalibrateFail     bool
	CurrInput           *Input
	CurrRecordHashSlice []string
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
	log.Printf("handle entry: %+v\n", *e)

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
		randNum := rand.Int31n(101)
		if e.BestScore < int(randNum) {
			log.Printf("[%+v] randomly skipped", *e)
			// if skip, simply add entry to the tail
			fuzzCtx.EnqueueQueryEntry(e)
			return nil
		}
		currentFuzzingEnergy := e.BestScore
		execCount := e.ExecutionCount
		log.Printf("[%+v] randomly mutate with energy %d", *e, currentFuzzingEnergy)
		for randFuzzIdx := 0; randFuzzIdx < currentFuzzingEnergy; randFuzzIdx++ {
			currentMutatedInput, err := RandomMutateInput(e.CurrInput)
			if err != nil {
				log.Printf("[%+v] randomly mutate input fail: %s, continue", *e, err)
				continue
			}
			t, err := NewRunTask(currentMutatedInput, e.Stage, e.Idx, execCount, e)
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
