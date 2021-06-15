package fuzzer

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"goFuzz/config"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

func getInputFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "input"))
}

func getOutputFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "stdout"))
}

func getRecordFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "record"))
}

// createOutputDir create an folder to contain/record all information about each run
func createOutputDir(input *Input, baseOutputDir string) (string, error) {
	dir := path.Join(baseOutputDir, input.GetID())
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return dir, os.MkdirAll(dir, os.ModePerm)
	} else {
		// return any other error if occurs
		return "", err
	}
}

func Run(input *Input) (*RunOutput, error) {
	var err error
	if input.TestName == "Empty" || input.TestName == "" {
		return nil, fmt.Errorf("the Run command in the fuzzer receive an input without input.TestName")
	}

	curRunOutputDir, err := createOutputDir(input, OutputDir)

	boolFirstRun := input.Note == NotePrintInput

	gfInputFp, err := getInputFilePath(curRunOutputDir)
	if err != nil {
		return nil, err
	}

	gfOutputFp, err := getOutputFilePath(curRunOutputDir)
	if err != nil {
		return nil, err
	}
	gfRecordFp, err := getRecordFilePath(curRunOutputDir)
	if err != nil {
		return nil, err
	}

	// Create the input file into disk
	SerializeInput(input, gfInputFp)

	// Run the test
	//err := os.Setenv("GOPATH", config.StrProjectGOPATH)
	//if err != nil {
	//	return nil, fmt.Errorf("the export of GOPATH fails: %s", err)
	//}

	err = os.Setenv("GF_RECORD_FILE", gfRecordFp)
	if err != nil {
		return nil, fmt.Errorf("the export of GF_RECORD_FILE fails: %s", err)
	}

	err = os.Setenv("GF_INPUT_FILE", gfInputFp)
	if err != nil {
		return nil, fmt.Errorf("the export of GF_INPUT_FILE fails: %s", err)
	}

	if GlobalTuple {
		err = os.Setenv("BitGlobalTuple", "1")
	} else {
		err = os.Setenv("BitGlobalTuple", "0")
	}
	if err != nil {
		return nil, fmt.Errorf("the export of TestPath fails: %s", err)
	}

	cmd := exec.Command("go", "test", "-run", input.TestName, "./...")
	cmd.Dir = TargetGoModDir
	var errBuf, stdOutBuf bytes.Buffer
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()

	outputF, err := os.Create(gfOutputFp)
	defer outputF.Close()
	outputW := bufio.NewWriter(outputF)
	outputW.Write(stdOutBuf.Bytes())
	outputW.Flush()

	if err != nil {
		return nil, fmt.Errorf("go test command fails: %s", err)
	}

	outputNumBug := ParseOutputFile(stdOutBuf.String())
	if outputNumBug != 0 {
		log.Println("Found a new bug. Now exit")
		os.Exit(1)
	}

	// Read the newly printed input file if this is the first run
	var retInput *Input
	if boolFirstRun {
		bytes, err := ioutil.ReadFile(gfInputFp)
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
	retInput.TestName = input.TestName
	// Read the printed record file
	b, err := ioutil.ReadFile(gfRecordFp)
	if err != nil {
		return nil, err
	}
	retRecord, err := ParseRecordFile(string(b))

	if err != nil {
		return nil, err
	}

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
