package fuzzer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"
)

func Deterministic_enumerate_input(input *Input) (reInputSlice []*Input) {

	for idx_vec_select, select_input := range input.VecSelect {
		for i := 0; i < select_input.IntNumCase; i++ {
			var tmp_input *Input
			tmp_input = copyInput(input)
			tmp_input.Note = ""
			tmp_input.VecSelect[idx_vec_select].IntPrioCase = i
			tmp_input.SelectDelayMS = 500 // TODO:: We may need to tune the number here
			reInputSlice = append(reInputSlice, tmp_input)
		}
	}
	return
}

func Get_Random_Int_With_Max(max int) int {
	mutateMethod, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		fmt.Println("Crypto/rand returned non-nil errors: ", err)
	}
	return int(mutateMethod.Int64())
}

// RandomMutateInput generates a new input by randomly mutating select choices within given input
// Notes:
//   RandomMutateInput will fail if input's VecSelect is empty
func RandomMutateInput(input *Input) (*Input, error) {
	numOfSelects := len(input.VecSelect)
	if numOfSelects == 0 {
		return nil, errors.New("cannot randomly mutate an input with empty VecSelect")
	}
	reInput := copyInput(input)
	reInput.SelectDelayMS += 500 // TODO:: we may need to tune the two numbers here
	if reInput.SelectDelayMS > 5000 {
		reInput.SelectDelayMS = 500
	}
	mutateMethod := Get_Random_Int_With_Max(2)

	switch mutateMethod {
	case 0:
		/* Mutate one select per time */
		mutateWhichSelect := Get_Random_Int_With_Max(numOfSelects)
		numOfSelectCases := reInput.VecSelect[mutateWhichSelect].IntNumCase
		if numOfSelectCases == 0 {
			return nil, fmt.Errorf("cannot randomly mutate an input with zero number of cases in select %d", mutateWhichSelect)
		}
		mutateToWhatValue := Get_Random_Int_With_Max(numOfSelectCases)

		reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue
	case 1:
		/* Mutate random number of select. */
		mutateChance := Get_Random_Int_With_Max(numOfSelects)
		for mutateIdx := 0; mutateIdx < mutateChance; mutateIdx++ {
			mutateWhichSelect := Get_Random_Int_With_Max(numOfSelects)
			numOfSelectCases := reInput.VecSelect[mutateWhichSelect].IntNumCase
			if numOfSelectCases == 0 {
				return nil, fmt.Errorf("cannot randomly mutate an input with zero number of cases in select %d", mutateWhichSelect)
			}
			mutateToWhatValue := Get_Random_Int_With_Max(numOfSelectCases)

			reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue
		}

	default:
		return nil, fmt.Errorf("cannot randomly mutate an input with non-exist mutate method %d", mutateMethod)
	}
	return reInput, nil
}

// Fuzz is the main entry for fuzzing
func Fuzz(tests []*GoTest, customCmds []string, numOfWorkers int) {
	for _, test := range tests {
		log.Printf("tests going to be fuzzed: %v from package %s", test.Func, test.Package)
	}
	log.Printf("custom commands going to be fuzzed: %s", customCmds)
	log.Printf("# of workers: %d", numOfWorkers)
	InitWorkers(numOfWorkers, fuzzerContext)

	for _, test := range tests {
		e := NewInitStageFuzzQueryEntryWithGoTest(test)
		fuzzerContext.EnqueueQueryEntry(e)
	}

	for _, cmd := range customCmds {
		e := NewInitStageFuzzQueryEntryWithCustomCmd(cmd)
		fuzzerContext.EnqueueQueryEntry(e)
	}

	for {
		e, err := fuzzerContext.DequeueQueryEntry()
		if err != nil {
			log.Println(err)
			continue
		}
		if e == nil {
			log.Println("queue is empty, wait 10 seconds and retry")
			time.Sleep(10 * time.Second)
			continue
		}
		err = HandleFuzzQueryEntry(e, fuzzerContext)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
