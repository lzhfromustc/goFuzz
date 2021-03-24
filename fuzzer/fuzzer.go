package fuzzer

import (
	"bytes"
	"fmt"
	"goFuzz/config"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Run(strTestName string, input Input) (retInput Input, retRecord Record) {
	boolFirstRun := input.Note == NoteEmpty
	// Create the input file into disk
	CreateInput(input)

	// Load the output file before running the test
	outputNumBugBefore := ParseOutputFile()

	// Run the test
	err := os.Setenv("GOPATH", config.StrProjectGOPATH)
	if err != nil {
		fmt.Println("The export of GOPATH fails:", err)
		return
	}
	err = os.Setenv("TestPath", config.StrTestPath)
	if err != nil {
		fmt.Println("The export of TestPath fails:", err)
		return
	}
	err = os.Setenv("OutputFullPath", config.StrOutputFullPath)
	if err != nil {
		fmt.Println("The export of OutputFullPath fails:", err)
		return
	}
	if config.BoolGlobalTuple {
		err = os.Setenv("BitGlobalTuple", "1")
	} else {
		err = os.Setenv("BitGlobalTuple", "0")
	}
	if err != nil {
		fmt.Println("The export of TestPath fails:", err)
		return
	}
	strRelativePath := strings.TrimPrefix(config.StrTestPath, config.StrProjectGOPATH + "/src/")
	cmd := exec.Command("go", "test", strRelativePath, "-run", strTestName)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		fmt.Println("The go test command fails:", err)
		return
	}
	fmt.Println("Output of unit test:")
	fmt.Println("test out:", outb.String(), "\ntest err:", errb.String())

	// If the output file is longer, it means we found a new bug
	outputNumBugAfter := ParseOutputFile()
	if outputNumBugAfter > outputNumBugBefore {
		fmt.Println("Found a new bug. Now exit")
		os.Exit(1)
	}


	// Read the newly printed input file if this is the first run
	if boolFirstRun {
		retInput = ParseInputFile()
	} else {
		retInput = EmptyInput()
	}
	// Read the printed record file
	retRecord = ParseRecordFile()
	return
}

func SetDeadline() {
	go func(){
		time.Sleep(config.FuzzerDeadline)
		fmt.Println("The checker has been running for",config.FuzzerDeadline,". Now force exit")
		os.Exit(1)
	}()
}

func PopWorklist(workList *[]Input) (result Input, numFile int) {
	result = (*workList)[0]
	if len(*workList) == 1 {
		*workList = nil
		return result, 0
	} else {
		(*workList) = (*workList)[1:]
		return result, len(*workList)
	}
}