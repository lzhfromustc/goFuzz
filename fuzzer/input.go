package fuzzer

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"goFuzz/config"
	"os"
	"strconv"
	"strings"
)

type Input struct {
	Note string
	SelectDelayMS int // How many milliseconds a select will wait for the prioritized case
	VecSelect []SelectInput
}

type SelectInput struct {
	StrFileName string
	IntLineNum int
	IntNumCase int
	IntPrioCase int
}

const (
	NoteEmpty string = "Empty"

	InputFileName string = "myinput.txt"
)

func EmptyInput() Input {
	return Input{
		Note:          NoteEmpty,
		SelectDelayMS: 0,
		VecSelect:     nil,
	}
}

func CreateInput(input Input) {
	out, err := os.Create(FileNameOfInput())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfInput())
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	str := StrOfInput(input)
	w.WriteString(str)
}

func StrOfInput(input Input) (retStr string) {
	if input.Note == NoteEmpty {
		retStr = NoteEmpty
		return
	}

	// The first line is how many seconds to wait
	retStr = strconv.Itoa(input.SelectDelayMS) + "\n"

	// Each line corresponds to a select
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

func HashOfInput(input Input) string {
	strOfInput := StrOfInput(input)
	h := sha256.New()
	h.Write([]byte(strOfInput))
	hash := string(h.Sum(nil))
	return hash
}

func FileNameOfInput() string {
	return config.StrTestPath + "/" + InputFileName
}

func ParseInputFile() (retInput Input) {
	retInput = EmptyInput()

	// The input being parsed shouldn't be empty
	file, err := os.Open(FileNameOfInput())
	if err != nil {
		fmt.Println("Failed to open input file:", FileNameOfInput())
		return
	}
	defer file.Close()

	var text []string

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	if len(text) == 0 {
		fmt.Println("Input is empty:", FileNameOfInput())
		return
	}

	newInput := Input{
		Note:          "",
		SelectDelayMS: 0,
		VecSelect:     []SelectInput{},
	}

	strDelayMS := text[0]
	newInput.SelectDelayMS, err = strconv.Atoi(strDelayMS)
	if err != nil {
		fmt.Println("The first line of input is not a number:", strDelayMS)
		return
	}

	for i, eachLine := range text {
		if i == 0 {
			continue
		}
		if eachLine == "" {
			continue
		}
		selectInput := SelectInput{}
		// filename:linenum:totalCaseNum:chooseCaseNum
		vecStr := strings.Split(eachLine, ":")
		if len(vecStr) != 4 {
			fmt.Println("One line in input has incorrect format:", eachLine, "\tLine:", i)
			return
		}
		selectInput.StrFileName = vecStr[0]
		if selectInput.IntLineNum, err = strconv.Atoi(vecStr[1]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return
		}
		if selectInput.IntNumCase, err = strconv.Atoi(vecStr[2]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return
		}
		if selectInput.IntPrioCase, err = strconv.Atoi(vecStr[3]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return
		}
		newInput.VecSelect = append(newInput.VecSelect, selectInput)
	}

	retInput = newInput
	return
}

func GenInputs(input Input) []Input {
	// for each select in input, generate a different file where other selects remain the same,
	// but one select prioritize the next case
	// Then for each different file, copy it multiple times, where SelectDelayMS are different
	vecInputs := []Input{}
	for i, selectInput := range input.VecSelect {
		for _, delayMS := range config.FuzzerSelectDelayVector {
			newInput := copyInput(input)
			newInput.SelectDelayMS = delayMS
			newSelectInput := copySelectInput(selectInput)
			if newSelectInput.IntPrioCase < newSelectInput.IntNumCase - 1 {
				newSelectInput.IntPrioCase += 1
			} else { // if this is the last case, go back to zero
				newSelectInput.IntPrioCase = 0
			}
			newInput.VecSelect[i] = newSelectInput

			vecInputs = append(vecInputs, newInput)
		}
	}

	return vecInputs
}

func InsertWorklist(newInputs, workList []Input, indexInsertAfter int) []Input {
	result := []Input{}
	if indexInsertAfter == -1 {
		for _, newInput := range newInputs {
			result = append(result, newInput)
		}
	}
	for i, oldInput := range workList {
		result = append(result, oldInput)
		if i == indexInsertAfter {
			for _, newInput := range newInputs {
				result = append(result, newInput)
			}
		}
	}
	return result
}

func InsertWorklistScore(curScore int, numNewInputs int, workListScore []int, indexInsertAfter int) []int {
	result := []int{}
	if indexInsertAfter == -1 {
		for i := 0; i < numNewInputs; i++ {
			result = append(result, curScore)
		}
	}
	for i, oldScore := range workListScore {
		result = append(result, oldScore)
		if i == indexInsertAfter {
			for i := 0; i < numNewInputs; i++ {
				result = append(result, curScore)
			}
		}
	}
	return result
}

func copySelectInput(sI SelectInput) SelectInput {
	return SelectInput{
		StrFileName: sI.StrFileName,
		IntLineNum:  sI.IntLineNum,
		IntNumCase:  sI.IntNumCase,
		IntPrioCase: sI.IntPrioCase,
	}
}

func copyInput(input Input) Input {
	newInput :=  Input{
		Note:          input.Note,
		SelectDelayMS: input.SelectDelayMS,
		VecSelect:     []SelectInput{},
	}
	for _, selectInput := range input.VecSelect {
		newInput.VecSelect = append(newInput.VecSelect, selectInput)
	}
	return newInput
}