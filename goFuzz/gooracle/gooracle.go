package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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
var StrTestFile string

var chEnforceCheck chan struct{}
var chDelayCheckSign chan struct{}
var intDelayCheckCounter int
const (
	DelayCheckModPerTime int = 0 // Check bugs every DelayCheckMS Milliseconds
	DelayCheckModCount int = 1 // Check bugs when runtime.EnqueueCheckEntry is called DelayCheckCountMax times
)

// config
var DelayCheckMod int = DelayCheckModPerTime
var DelayCheckMS int = 1000
var DelayCheckCountMax int = 10

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
	StrTestName = runtime.MyCaller(1)
	if indexDot := strings.Index(StrTestName, "."); indexDot > -1 {
		StrTestName = StrTestName[indexDot+1:]
	}
	_, StrTestFile, _, _ = runtime.Caller(2)
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
	//OpenOutputFile() // No need

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
	} else {
		BoolFirstRun = false
	}

	MapInput = ParseInputStr(text)
	if MapInput == nil {
		fmt.Println("Error when parsing input during text start: MapInput is nil")
	}

	if runtime.BoolDelayCheck {
		chEnforceCheck = make(chan struct{})
		chDelayCheckSign = make(chan struct{}, 10)
		if DelayCheckMod == DelayCheckModCount {
			runtime.FnCheckCount = DelayCheckCounterFN
		}
		go CheckBugClient()
		go CheckBugServer()
	}
}

// An endless loop that checks bug
func CheckBugServer() {
	defer close(chEnforceCheck)
	for {
		boolBreakLoop := false
		switch DelayCheckMod {
		case DelayCheckModPerTime:
			select {
			case <-time.After(time.Millisecond * time.Duration(DelayCheckMS)):
			case <-chEnforceCheck:
				boolBreakLoop = true
			}
		case DelayCheckModCount:
			select {
			case <-chDelayCheckSign:
			case <-chEnforceCheck:
				boolBreakLoop = true
			}
		}

		for len(runtime.VecCheckEntry) > 0 {
			checkEntry := runtime.DequeueCheckEntry()
			if atomic.LoadUint32(&checkEntry.Uint32NeedCheck) == 1 {
				runtime.CheckBlockBug(checkEntry.CS)
			}
		}
		if boolBreakLoop {
			break
		}
	}
}

func CheckBugClient() {

}

func DelayCheckCounterFN() {
	if DelayCheckMod == DelayCheckModCount {
		intDelayCheckCounter++ // no need to worry about data race, since runtime.MuCheckEntry is held
		if intDelayCheckCounter > DelayCheckCountMax {
			intDelayCheckCounter = 0
			select { // the channel has buffer, so default should be rarely chosen
			case chDelayCheckSign <- struct{}{}:
			default:
			}
		}
	}
}

func AfterRun() {
	switch StrTestMod {
	case "TestOnce": // Run all unit tests once, and print a file containing each test's name, # of select visited
		AfterRunTestOnce()
	default: // Normal fuzzing
		AfterRunFuzz()
	}
}

func AfterRunTestOnce() {
	strOutputPath := os.Getenv("OutputFullPath")
	out, err := os.OpenFile(strOutputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
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
	return "\n" + StrTestFile + ":" + StrTestName + ":" + strconv.Itoa(int(runtime.ReadSelectCount()))
}

func AfterRunFuzz() {
	PrintNumTimeoutSelect() // Print the number of selects, number of timeout selects and not in input selects

	if runtime.BoolDelayCheck { // This is the end of program, need to do all delayed bug detection
		chEnforceCheck <- struct{}{}
		<-chEnforceCheck // wait for checking to finish
	}

	// if this is the first run, create input file using runtime's global variable
	if BoolFirstRun {
		CreateInput()
	}

	// create output file using runtime's global variable
	CreateRecordFile()

	// print bug info
	str, foundBug := runtime.TmpDumpBlockingInfo()
	if foundBug {
		fmt.Println(str)
	}

	// dump operation records
	opFile := os.Getenv("GF_OP_COV_FILE")
	if opFile != "" {
		err := dumpOpRecordsToFile(opFile, opRecords)
		if err != nil {
			// print to error
			println(err)
		}
	}


}
