package fuzzer

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// Input contains all information about
// 1. need to be passed to fuzz target by environment variables
// 2. how to trigger fuzz target
type Input struct {
	// First line of input file, string `PrintInput` for recording input (used by gooracle),
	// otherwise it is just a placeholder for the first line of file to make sure file has correct format.
	Note string

	// If we are running fuzz target by go test command
	GoTestCmd *GoTest

	// If we are running fuzz target by custom command
	CustomCmd string

	// How many milliseconds a select will wait for the prioritized case
	SelectDelayMS int

	// Select choice need to be forced during runtime
	VecSelect []SelectInput
}

func (i *Input) String() string {
	if i.GoTestCmd != nil {
		return fmt.Sprintf("[pkg %s][func %s]", i.GoTestCmd.Package, i.GoTestCmd.Func)
	} else if i.CustomCmd != "" {
		return i.CustomCmd
	}

	return "empty"
}

// Src returns the which source, either pkg + test or custom command, the input will trigger
func (i *Input) Src() string {
	var src string

	if i.GoTestCmd != nil {
		if i.GoTestCmd.Package != "" {
			// Whole package URL might be too long for filesystem, so only use last component of the package URL.
			basePkg := path.Base(i.GoTestCmd.Package)
			src = fmt.Sprintf("%s-%s", basePkg, i.GoTestCmd.Func)
		} else {
			src = i.GoTestCmd.Func
		}
	} else if i.CustomCmd != "" {
		src = i.CustomCmd
	}

	return src
}

func ShouldSkipInput(fuzzCtx *FuzzContext, i *Input) bool {
	if i == nil {
		return true
	}
	src := i.Src()

	// drop it if list of selects has been run more than 3 times
	if GetCountsOfSelects(src, i.VecSelect) > 5 {
		log.Printf("drop %s since the select inputs run more than 3 times", i)
		return false
	}

	// drop it if it's source already timeout more than 3 times
	fuzzCtx.timeoutTargetsLock.RLock()
	timeoutCnt, exist := fuzzCtx.timeoutTargets[src]
	fuzzCtx.timeoutTargetsLock.RUnlock()
	if exist {
		if timeoutCnt > 3 {
			log.Printf("drop %s since it has timeout more than 3 times", i)
			return true
		}
	}

	if SkipIntegration == true {
		if i.GoTestCmd != nil && (strings.Contains(i.GoTestCmd.Func, "integration") || strings.Contains(i.GoTestCmd.Func, "Integration")) {
			log.Printf("drop %s since it is a integration unit test", i)
			return true
		} else if i.CustomCmd != "" && (strings.Contains(i.CustomCmd, "integration") || strings.Contains(i.CustomCmd, "Integration")) {
			log.Printf("drop %s since it is a integration unit test", i)
			return true
		}
	}

	return false
}

type GoTest struct {
	// If test should be triggered from compiled binary file
	Bin string

	// Test function name
	Func string

	// Which package the test function located (if bin has set, package is bin file name since test bin is compiled by package level)
	Package string
}

type SelectInput struct {
	StrFileName string
	IntLineNum  int
	IntNumCase  int
	IntPrioCase int
}

func GetHashOfSelects(selects []SelectInput) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", selects)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

const (
	NotePrintInput string = "PrintInput"
)

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

func copySelectInput(sI SelectInput) SelectInput {
	return SelectInput{
		StrFileName: sI.StrFileName,
		IntLineNum:  sI.IntLineNum,
		IntNumCase:  sI.IntNumCase,
		IntPrioCase: sI.IntPrioCase,
	}
}

func copyGoTest(gt *GoTest) *GoTest {
	if gt == nil {
		return nil
	}

	return &GoTest{
		Func:    gt.Func,
		Package: gt.Package,
	}

}

func copyInput(input *Input) *Input {
	newInput := &Input{
		Note:          input.Note,
		GoTestCmd:     copyGoTest(input.GoTestCmd),
		CustomCmd:     input.CustomCmd,
		SelectDelayMS: input.SelectDelayMS,
		VecSelect:     []SelectInput{},
	}
	for _, selectInput := range input.VecSelect {
		newInput.VecSelect = append(newInput.VecSelect, copySelectInput(selectInput))
	}
	return newInput
}
