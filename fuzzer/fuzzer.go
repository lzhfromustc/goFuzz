package fuzzer

import (
	"fmt"
	"gotest/config"
	"os"
	"os/exec"
	"time"
)

func Run(strTestName string, input Input) (retInput Input, retRecord Record) {
	boolFirstRun := input.Note == NoteEmpty
	// Create the input file into disk
	CreateInput(input)

	// Run the test
	cmd := exec.Command("export", "GOPATH=" + config.StrProjectGOPATH)
	err := cmd.Run()
	if err != nil {
		fmt.Println("The export of GOPATH fails:", err)
		return
	}
	cmd = exec.Command("cd", config.StrTestPath)
	err = cmd.Run()
	if err != nil {
		fmt.Println("cd to the test path fails:", err)
		return
	}
	cmd = exec.Command("go", "test", "-run", strTestName)
	err = cmd.Run()
	if err != nil {
		fmt.Println("The go test command fails:", err)
		return
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