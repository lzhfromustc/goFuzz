package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
)

const (
	NotePrintInput string = "PrintInput"
	InputFileName  string = "myinput.txt"
	RecordFileName        = "myrecord.txt"
	OutputFileName        = "myoutput.txt"
	ErrFileName           = "myerror.txt"
	RecordSplitter        = "-----"
)

var StrTestpath string
var BoolFirstRun bool = true

func BeforeRun() {
	StrBitGlobalTuple := os.Getenv("BitGlobalTuple")
	if StrBitGlobalTuple == "1" {
		runtime.BoolRecordPerCh = false
	} else {
		runtime.BoolRecordPerCh = true
	}
	StrTestpath = os.Getenv("TestPath")
	//StrTestpath ="/data/ziheng/shared/gotest/gotest/src/gotest/testdata/toyprogram"

	// Create an output file and bound os.Stdout to it
	//OpenOutputFile()

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

func AfterRun() {

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
	//CloseOutputFile()
}

func StoreOpInfo(strOpType string, uint16OpID uint16) {
	runtime.StoreChOpInfo(strOpType, uint16OpID)
}
