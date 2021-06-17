package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

const (
	NotePrintInput string = "PrintInput"
	InputFileName  string = "myinput.txt"
	RecordFileName        = "myrecord.txt"
	OutputFileName        = "myoutput.txt"
	ErrFileName           = "myerror.txt"
	RecordSplitter        = "-----"
)

var StrTestPath string
var BoolFirstRun bool = true
var StrTestMod string
var StrTestName string

func BeforeRun() {
	StrTestMod = os.Getenv("TestMod")
	switch StrTestMod {
	case "TestOnce": // Run all unit tests once, and print a file containing each test's name, # of select visited
		BeforeRunTestOnce()
	default: // Normal fuzzing
		BeforeRunFuzz()
	}
}

func BeforeRunTestOnce() {
	StrTestPath = os.Getenv("TestPath")
	StrTestName = runtime.MyCaller()
	runtime.BoolSelectCount = true
}

func BeforeRunFuzz() {
	StrBitGlobalTuple := os.Getenv("BitGlobalTuple")
	if StrBitGlobalTuple == "1" {
		runtime.BoolRecordPerCh = false
	} else {
		runtime.BoolRecordPerCh = true
	}
	StrTestPath = os.Getenv("TestPath")
	//StrTestPath ="/data/ziheng/shared/gotest/gotest/src/gotest/testdata/toyprogram"

	// Create an output file and bound os.Stdout to it
	OpenOutputFile()

	// read input file
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

	if len(text) > 0 && text[0] == NotePrintInput {
		runtime.RecordSelectChoice = true
	}

	MapInput = ParseInputStr(text)
	if MapInput == nil {
		fmt.Println("Error when parsing input during text start: MapInput is nil")
	}
}

func AftterRun() {
	switch StrTestMod {
	case "TestOnce": // Run all unit tests once, and print a file containing each test's name, # of select visited
		AfterRunTestOnce()
	default: // Normal fuzzing
		AfterRunFuzz()
	}
}

func AfterRunTestOnce() {
	strOutputPath := os.Getenv("OutputFullPath")
	out, err := os.Create(strOutputPath)
	if err != nil {
		fmt.Println("Failed to create file:", strOutputPath)
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	w.WriteString(StrTestNameAndSelectCount())
}

func StrTestNameAndSelectCount() string {
	return "\n" + StrTestName + ":" + strconv.Itoa(int(runtime.ReadSelectCount()))
}

func AfterRunFuzz() {

	// if this is the first run, create input file using runtime's global variable
	if BoolFirstRun {
		CreateInput()
	}

	// create output file using runtime's global variable
	CreateRecordFile()

	// print bug info
	str, foundBug := runtime.TmpDumpBlockingInfo()
	if foundBug {
		fmt.Println("-----New Bug:")
		fmt.Println(str)
	}
	CloseOutputFile()
}
