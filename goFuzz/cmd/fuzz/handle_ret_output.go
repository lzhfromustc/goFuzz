package main

import (
	"fmt"
	"goFuzz/fuzzer"
)

func handleRetOutput(retOutput *fuzzer.RunOutput) {
	if retOutput == nil || retOutput.RetInput == nil || retOutput.RetRecord == nil {
		// TODO:: I found some empty retInput again. Should I be worry?
		fmt.Println("Error!!! Empty retInput!!!")
		return
	}
	fmt.Println("Deter: Reading outputs from the workers. ")
	/* We don't need to care about the running stage here. It would always be "deter". */
	retInput := retOutput.RetInput
	retRecord := retOutput.RetRecord
	fuzzer.HandleRunOutput(retInput, retRecord, retOutput.Stage, nil, mainRecord, &fuzzingQueue, allRecordHashMap)

	exists := deteredTestname[retInput.TestName]

	if !exists {
		/* Generate all possible deter_inputs based on the current retInput. Only changing one select per time. */
		deterInputSlice := fuzzer.Deterministic_enumerate_input(retInput)

		/* Now we deterministically enumerate one select per time, iterating through all the pTestName and all selects */
		for _, deterInput := range deterInputSlice {
			deterInput.Stage = "deter" // The current stage of fuzzing is "deterministic fuzzing".
			workerInputChan <- deterInput
		}

		deteredTestname[retInput.TestName] = true
	}

}
