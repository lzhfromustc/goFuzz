package fuzzer

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"goFuzz/config"
	"os"
	"strconv"
	"strings"
)

type Input struct {
	Note           string
	TestName       string
	SelectDelayMS  int    // How many milliseconds a select will wait for the prioritized case
	Stage          string // "unknown", "deter", "calib" or "rand"
	VecSelect      []SelectInput
	InputFilePath  string
	RecordFilePath string
	OutputFilePath string
}

type SelectInput struct {
	StrFileName string
	IntLineNum  int
	IntNumCase  int
	IntPrioCase int
}

const (
	NotePrintInput string = "PrintInput"
	NoteEmptyName  string = "Empty"
	InputFileName  string = "myinput.txt"
)

func EmptyRunOutput() *RunOutput {
	return &RunOutput{
		RetInput:  nil,
		RetRecord: nil,
		Stage:     "Unknown",
	}
}

func EmptyInput() *Input {
	return &Input{
		TestName:      NoteEmptyName,
		Note:          NotePrintInput,
		SelectDelayMS: 0,
		VecSelect:     nil,
		Stage:         "unknown",
	}
}

// SerializeInput dump input as string to file
func SerializeInput(input *Input, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	str := StrOfInput(input)
	w.WriteString(str)
	return nil
}

func StrOfInput(input *Input) (retStr string) {
	retStr = ""

	// The first line is a note. Could be empty.
	if input.Note == NotePrintInput {
		retStr += NotePrintInput + "\n"
	} else {
		retStr += "\n"
	}

	// The second line is how many milliseconds to wait
	retStr += strconv.Itoa(input.SelectDelayMS) + "\n"

	// From the third line, each line corresponds to a select
	for _, selectInput := range input.VecSelect {
		// filename:linenum:totalCaseNum:chooseCaseNum
		str := selectInput.StrFileName + ":" + strconv.Itoa(selectInput.IntLineNum)
		str += ":" + strconv.Itoa(selectInput.IntNumCase)
		str += ":" + strconv.Itoa(selectInput.IntPrioCase)
		str += "\n"
		retStr += str
	}
	return
}

func HashOfRecord(record *Record) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", record))) // TODO: we may need to replace `record` with `StrOfRecord(record)`
	return fmt.Sprintf("%x", h.Sum(nil))
}

func FindRecordHashInSlice(recordHash string, recordHashSlice []string) bool {
	for _, searchRecordHash := range recordHashSlice {
		if recordHash == searchRecordHash {
			return true
		}
	}
	return false
}

func FileNameOfInput() string {
	return config.StrTestPath + "/" + InputFileName
}

func ParseInputFile(content string) (*Input, error) {
	var err error

	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return nil, errors.New("Input has less than 2 lines")
	}

	newInput := &Input{
		Note:          lines[0],
		SelectDelayMS: 0,
		VecSelect:     []SelectInput{},
	}

	strDelayMS := lines[1]
	newInput.SelectDelayMS, err = strconv.Atoi(strDelayMS)
	if err != nil {
		return nil, err
	}

	// Skip line 1 (PrintInput) and line 2 (select time out)
	selectInfos := lines[2:]
	for _, eachLine := range selectInfos {
		if eachLine == "" {
			continue
		}
		selectInput, err := ParseSelectInput(eachLine)
		if err != nil {
			return nil, err
		}

		newInput.VecSelect = append(newInput.VecSelect, *selectInput)
	}

	return newInput, nil
}

// ParseSelectInput parses the each select in input file
// which has format filename:linenum:totalCaseNum:chooseCaseNum
func ParseSelectInput(line string) (*SelectInput, error) {
	var err error
	selectInput := SelectInput{}
	vecStr := strings.Split(line, ":")
	if len(vecStr) != 4 {
		return nil, fmt.Errorf("expect number of components: 4, actual: %d", len(vecStr))
	}
	selectInput.StrFileName = vecStr[0]
	if selectInput.IntLineNum, err = strconv.Atoi(vecStr[1]); err != nil {
		return nil, fmt.Errorf("incorrect format at line number")
	}
	if selectInput.IntNumCase, err = strconv.Atoi(vecStr[2]); err != nil {
		return nil, fmt.Errorf("incorrect format at number of cases")
	}
	if selectInput.IntPrioCase, err = strconv.Atoi(vecStr[3]); err != nil {
		return nil, fmt.Errorf("incorrect format at priority case")
	}
	return &selectInput, nil
}

//func GenInputs(input Input) []Input {
//	// for each select in input, generate a different file where other selects remain the same,
//	// but one select prioritize the next case
//	// Then for each different file, copy it multiple times, where SelectDelayMS are different
//	vecInputs := []Input{}
//	for i, selectInput := range input.VecSelect {
//		for _, delayMS := range config.FuzzerSelectDelayVector {
//			newInput := copyInput(input)
//			newInput.SelectDelayMS = delayMS
//			newSelectInput := copySelectInput(selectInput)
//			if newSelectInput.IntPrioCase < newSelectInput.IntNumCase - 1 {
//				newSelectInput.IntPrioCase += 1
//			} else { // if this is the last case, go back to zero
//				newSelectInput.IntPrioCase = 0
//			}
//			newInput.VecSelect[i] = newSelectInput
//
//			vecInputs = append(vecInputs, newInput)
//		}
//	}
//
//	return vecInputs
//}
//
//func InsertWorklist(newInputs, workList []Input, indexInsertAfter int) []Input {
//	result := []Input{}
//	if indexInsertAfter == -1 {
//		for _, newInput := range newInputs {
//			result = append(result, newInput)
//		}
//	}
//	for i, oldInput := range workList {
//		result = append(result, oldInput)
//		if i == indexInsertAfter {
//			for _, newInput := range newInputs {
//				result = append(result, newInput)
//			}
//		}
//	}
//	return result
//}
//
//func InsertWorklistScore(curScore int, numNewInputs int, workListScore []int, indexInsertAfter int) []int {
//	result := []int{}
//	if indexInsertAfter == -1 {
//		for i := 0; i < numNewInputs; i++ {
//			result = append(result, curScore)
//		}
//	}
//	for i, oldScore := range workListScore {
//		result = append(result, oldScore)
//		if i == indexInsertAfter {
//			for i := 0; i < numNewInputs; i++ {
//				result = append(result, curScore)
//			}
//		}
//	}
//	return result
//}
//
func copySelectInput(sI SelectInput) SelectInput {
	return SelectInput{
		StrFileName: sI.StrFileName,
		IntLineNum:  sI.IntLineNum,
		IntNumCase:  sI.IntNumCase,
		IntPrioCase: sI.IntPrioCase,
	}
}

func copyInput(input *Input) *Input {
	newInput := &Input{
		Note:          input.Note,
		TestName:      input.TestName,
		SelectDelayMS: input.SelectDelayMS,
		VecSelect:     []SelectInput{},
	}
	for _, selectInput := range input.VecSelect {
		newInput.VecSelect = append(newInput.VecSelect, copySelectInput(selectInput)) // TODO:: Here, the original is append(..., selectInput), not the copy of it. A bug?
	}
	return newInput
}
