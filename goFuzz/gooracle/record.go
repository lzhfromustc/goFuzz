package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
)

const MapOpLength int = 65536 // 2^16
const MapChLength int = 65536 // 2^16


func XOR(a, b [2]byte) [2]byte {
	a[0] ^= b[0]
	a[1] ^= b[1]
	return a
}

const (
	NotePrintInput string = "PrintInput"
	InputFileName  string = "myinput.txt"
	RecordFileName        = "myrecord.txt"
	OutputFileName        = "myoutput.txt"
	ErrFileName           = "myerror.txt"
	RecordSplitter        = "-----"
)

var StrTestpath string
var BoolFirstRun bool = false

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
		BoolFirstRun = true
	}

	if BoolFirstRun { // if input is empty, then this is the first run. Let runtime know
		runtime.GenFirstInput = true
		return
	} else { // if input is not empty, store this input into runtime
		runtime.GenFirstInput = false
		runtime.MapInput = ParseInputStr(text)
		if runtime.MapInput == nil {
			fmt.Println("Error when parsing input during text start: runtime.MapInput is nil")
		}
		return
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
	str, foundBug := runtime.DumpBlockingInfo()
	if foundBug {
		fmt.Println("-----New Bug:")
		fmt.Println(str)
	}
	CloseOutputFile()
}