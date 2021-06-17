package fuzzer

import (
	"crypto/rand"
	"fmt"
	"goFuzz/config"
	"log"
	"math/big"
	"os"
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
	mutateMethod, err := rand.Int(rand.Reader, big.NewInt(1))
	if err != nil {
		fmt.Println("Crypto/rand returned non-nil errors: ", err)
	}
	return int(mutateMethod.Int64())
}

func Random_Mutate_Input(input *Input) (reInput *Input) {
	/* TODO:: In the current stage, I am not mutating the delayMS number!!! */
	reInput = copyInput(input)
	reInput.SelectDelayMS += 500 // TODO:: we may need to tune the two numbers here
	if reInput.SelectDelayMS > 5000 {
		reInput.SelectDelayMS = 500
	}
	mutateMethod := Get_Random_Int_With_Max(4)
	switch mutateMethod {
	case 0:
		/* Mutate one select per time */
		mutateWhichSelect := Get_Random_Int_With_Max(len(reInput.VecSelect))
		mutateToWhatValue := Get_Random_Int_With_Max(reInput.VecSelect[mutateWhichSelect].IntNumCase)
		reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue

	case 1:
		/* Mutate two select per time */
		for mutateIdx := 0; mutateIdx < 2; mutateIdx++ {
			mutateWhichSelect := Get_Random_Int_With_Max(len(reInput.VecSelect))
			mutateToWhatValue := Get_Random_Int_With_Max(reInput.VecSelect[mutateWhichSelect].IntNumCase)
			reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue
		}

	case 2:
		/* Mutate three select per time */
		for mutateIdx := 0; mutateIdx < 3; mutateIdx++ {
			mutateWhichSelect := Get_Random_Int_With_Max(len(reInput.VecSelect))
			mutateToWhatValue := Get_Random_Int_With_Max(reInput.VecSelect[mutateWhichSelect].IntNumCase)
			reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue
		}

	case 3:
		/* Mutate random number of select. */ // TODO:: Not sure whether it is necessary. Just put it here now.
		mutateChance := Get_Random_Int_With_Max(len(reInput.VecSelect))
		for mutateIdx := 0; mutateIdx < mutateChance; mutateIdx++ {
			mutateWhichSelect := Get_Random_Int_With_Max(len(reInput.VecSelect))
			mutateToWhatValue := Get_Random_Int_With_Max(reInput.VecSelect[mutateWhichSelect].IntNumCase)
			reInput.VecSelect[mutateWhichSelect].IntPrioCase = mutateToWhatValue
		}

	default:
		/* ??? ERROR ??? */
		fmt.Println("Random Mutate Input is not mutating.")
	}
	return
}

func SetDeadline() {
	go func() {
		time.Sleep(config.FuzzerDeadline)
		fmt.Println("The checker has been running for", config.FuzzerDeadline, ". Now force exit")
		os.Exit(1)
	}()
}

// Fuzz is the main entry for fuzzing
func Fuzz(tests []string, customCmds []string, numOfWorkers int) {

	log.Printf("Tests going to be run: %s", tests)
	log.Printf("Custom Commands going to be run: %s", customCmds)
	log.Printf("Number of workers: %d", numOfWorkers)
	InitWorkers(numOfWorkers, fuzzerContext)

	for _, test := range tests {
		e := NewInitStageFuzzQueryEntryWithTestname(test)
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
			log.Println("Fuzzing queue is empty, waiting 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		err = HandleFuzzQueryEntry(e, fuzzerContext)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
