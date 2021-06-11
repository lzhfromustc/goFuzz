package fuzzer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"goFuzz/config"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"
)

type FuzzQueryEntry struct {
	IsFavored              bool
	BestScore              int // TODO:: I only save the BestScore for the CurrentInput, is that enough?
	ExecutionCount         int
	IsCalibrateFail        bool
	CurrentInput           *Input
	CurrentRecordHashSlice []string
	/* Add more features to the queue if necessary. */
}

func Deterministic_enumerate_input(input *Input) (reInputSlice []*Input) {
	var tmp_input *Input
	for idx_vec_select, select_input := range input.VecSelect {
		for i := 0; i < select_input.IntNumCase; i++ {
			tmp_input = copyInput(input)
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

func Run(input *Input) (*RunOutput, error) {
	if input.TestName == "Empty" || input.TestName == "" {
		return nil, fmt.Errorf("the Run command in the fuzzer receive an input without input.TestName")
	}
	strTestName := input.TestName
	boolFirstRun := input.Note == NotePrintInput
	// Create the input file into disk
	SerializeInput(input, input.InputFilePath)

	// Run the test
	err := os.Setenv("GOPATH", config.StrProjectGOPATH)
	if err != nil {
		return nil, fmt.Errorf("the export of GOPATH fails: %s", err)
	}
	err = os.Setenv("TestPath", config.StrTestPath)
	if err != nil {
		return nil, fmt.Errorf("the export of TestPath fails: %s", err)
	}
	err = os.Setenv("OutputFullPath", input.OutputFilePath)
	if err != nil {
		return nil, fmt.Errorf("the export of OutputFullPath fails: %s", err)
	}
	if config.BoolGlobalTuple {
		err = os.Setenv("BitGlobalTuple", "1")
	} else {
		err = os.Setenv("BitGlobalTuple", "0")
	}
	if err != nil {
		return nil, fmt.Errorf("the export of TestPath fails: %s", err)
	}
	strRelativePath := strings.TrimPrefix(config.StrTestPath, config.StrProjectGOPATH+"/src/")
	cmd := exec.Command("go", "test", strRelativePath, "-run", strTestName) // TODO: Consider handling the case that strTestName isn't a unit test
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	fmt.Println("Output of unit test:") // this output is meaningless. It just prints things to indicate whether the unit test passed or not. This has nothing to do with whether a bug is triggered
	fmt.Println("test out:", outb.String(), "\ntest err:", errb.String())
	if err != nil {
		return nil, fmt.Errorf("go test command fails: %s", err)
	}

	// If the output file is longer, it means we found a new bug
	bytes, err := ioutil.ReadFile(input.OutputFilePath)
	if err != nil {
		return nil, err
	}

	outputNumBug := ParseOutputFile(string(bytes))
	if outputNumBug != 0 {
		fmt.Println("Found a new bug. Now exit")
		os.Exit(1)
	}

	// Read the newly printed input file if this is the first run
	retInput := EmptyInput()
	if boolFirstRun {
		bytes, err := ioutil.ReadFile(input.InputFilePath)
		if err != nil {
			return nil, err
		}

		retInput, err = ParseInputFile(string(bytes))
		if err != nil {
			return nil, err
		}

	} else {
		retInput = EmptyInput()
	}
	// Save the current TestName to the retInput.
	retInput.TestName = strTestName
	// Read the printed record file
	bytes, err = ioutil.ReadFile(input.RecordFilePath)
	if err != nil {
		return nil, err
	}
	retRecord := ParseRecordFile(string(bytes))

	retOutput := EmptyRunOutput()
	retOutput.RetInput = retInput
	retOutput.RetRecord = retRecord
	/* Pass the stage information to the output, otherwise, when the main routine receive the output,
	it does not know the context fo the executions. */
	retOutput.Stage = input.Stage
	return retOutput, nil
}

func SetDeadline() {
	go func() {
		time.Sleep(config.FuzzerDeadline)
		fmt.Println("The checker has been running for", config.FuzzerDeadline, ". Now force exit")
		os.Exit(1)
	}()
}

func HandleRunOutput(retInput *Input, retRecord *Record, stage string, currentEntry *FuzzQueryEntry, mainRecord *Record, fuzzingQueue *[]FuzzQueryEntry, allRecordHashMap map[string]struct{}) {
	if stage == "calib" {
		if len(retInput.VecSelect) == 0 { // TODO:: Should we ignore the output that contains no VecSelects entry?
			return
		}
		recordHash := HashOfRecord(retRecord)
		if FindRecordHashInSlice(recordHash, currentEntry.CurrentRecordHashSlice) == false {
			currentEntry.CurrentRecordHashSlice = append(currentEntry.CurrentRecordHashSlice, recordHash)
		}
		if _, exist := allRecordHashMap[recordHash]; exist == false {
			allRecordHashMap[recordHash] = struct{}{}
		}
		curScore := ComputeScore(mainRecord, retRecord)
		if curScore > currentEntry.BestScore {
			currentEntry.BestScore = curScore
		}

	} else if stage == "deter" {
		if len(retInput.VecSelect) == 0 { // TODO:: Should we ignore the output that contains no VecSelects entry?
			return
		}
		recordHash := HashOfRecord(retRecord)
		/* See whether the current deter_input trigger a new record. If yes, save the record hash and the input to the queue. */
		if _, exist := allRecordHashMap[recordHash]; exist == false {
			curScore := ComputeScore(mainRecord, retRecord)
			currentFuzzEntry := FuzzQueryEntry{
				IsFavored:              false,
				ExecutionCount:         1,
				BestScore:              curScore,
				CurrentInput:           retInput,
				CurrentRecordHashSlice: []string{recordHash},
			}
			*fuzzingQueue = append(*fuzzingQueue, currentFuzzEntry)
			allRecordHashMap[recordHash] = struct{}{}
		} else {
			return
		}

	} else if stage == "rand" {
		recordHash := HashOfRecord(retRecord)
		if _, exist := allRecordHashMap[recordHash]; exist == false { // Found a new input with unique record!!!
			curScore := ComputeScore(mainRecord, retRecord)
			currentFuzzEntry := FuzzQueryEntry{
				IsFavored:              false,
				ExecutionCount:         1,
				BestScore:              curScore,
				CurrentInput:           retInput, // TODO:: Should we save ori_input or retInput???
				CurrentRecordHashSlice: []string{recordHash},
			}
			*fuzzingQueue = append(*fuzzingQueue, currentFuzzEntry)
			allRecordHashMap[recordHash] = struct{}{}
		} else {
			return
		} // This mutation does not create new record. Discarded.
		currentEntry.ExecutionCount += 1
	}
}

//func PopWorklist(workList *[]Input) (result Input, numFile int) {
//	result = (*workList)[0]
//	if len(*workList) == 1 {
//		*workList = nil
//		return result, 0
//	} else {
//		(*workList) = (*workList)[1:]
//		return result, len(*workList)
//	}
//}
