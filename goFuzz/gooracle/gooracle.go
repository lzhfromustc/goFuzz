package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

type OracleEntry struct {
	WgCheckBug *sync.WaitGroup
}

func BeforeRun() *OracleEntry {
	StrTestMod = os.Getenv("TestMod")
	switch StrTestMod {
	case "TestOnce": // Run all unit tests once, and print a file containing each test's name, # of select visited
		return BeforeRunTestOnce()
	default: // Normal fuzzing
		return BeforeRunFuzz()
	}
}

func BeforeRunTestOnce() *OracleEntry {
	StrTestPath = os.Getenv("TestPath")
	StrTestName = runtime.MyCaller(1)
	if indexDot := strings.Index(StrTestName, "."); indexDot > -1 {
		StrTestName = StrTestName[indexDot+1:]
	}
	_, StrTestFile, _, _ = runtime.Caller(2)
	runtime.BoolSelectCount = true
	return &OracleEntry{WgCheckBug: &sync.WaitGroup{}}
}

func BeforeRunFuzz() (result *OracleEntry) {
	result = &OracleEntry{WgCheckBug: &sync.WaitGroup{}}
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

	CheckBugStart(result)
	return
}

// Only enables oracle
func LightBeforeRun() *OracleEntry {
	entry := &OracleEntry{WgCheckBug: &sync.WaitGroup{}}
	CheckBugStart(entry)
	return entry
}

// Start the endless loop that checks bug. Should be called at the beginning of unit test
func CheckBugStart(entry *OracleEntry) {
	go CheckBugLate()
	if runtime.BoolDelayCheck {
		chEnforceCheck = make(chan struct{})
		chDelayCheckSign = make(chan struct{}, 10)
		if DelayCheckMod == DelayCheckModCount {
			runtime.FnCheckCount = DelayCheckCounterFN
		}
		go CheckBugRun(entry)
	}
}

// An endless loop that checks bug. Exits when the unit test ends
func CheckBugRun(entry *OracleEntry) {
	entry.WgCheckBug.Add(1)
	defer entry.WgCheckBug.Done()

	boolBreakLoop := false
	for {
		switch DelayCheckMod {
		case DelayCheckModPerTime:
			select {
			case <-time.After(time.Millisecond * time.Duration(DelayCheckMS)):
			case <-chEnforceCheck:
				if runtime.BoolDebug {
					fmt.Printf("Check bugs at the end of unit test\n")
				}
				boolBreakLoop = true
			}
		case DelayCheckModCount:
			select {
			case <-chDelayCheckSign:
			case <-chEnforceCheck:
				if runtime.BoolDebug {
					fmt.Printf("Check bugs at the end of unit test\n")
				}
				boolBreakLoop = true
			}
		}

		enqueueAgain := [][]runtime.PrimInfo{}
		for len(runtime.VecCheckEntry) > 0 {
			checkEntry := runtime.DequeueCheckEntry()
			if runtime.BoolDebug {
				print("Dequeueing:")
				for _, C := range checkEntry.CS {
					if ch, ok := C.(*runtime.ChanInfo); ok {
						print("\t", ch.StrDebug)
					}
				}
				println()
			}
			if atomic.LoadUint32(&checkEntry.Uint32NeedCheck) == 1 {
				if runtime.CheckBlockBug(checkEntry.CS) == false { // CS needs to be checked again in the future
					enqueueAgain = append(enqueueAgain, checkEntry.CS)
				}
			}
		}
		for _, CS := range enqueueAgain {
			runtime.EnqueueCheckEntry(CS)
		}
		if boolBreakLoop {
			break
		}
	}
}

func CheckBugLate() {
	time.Sleep(2 * time.Minute) // Before the deadline we set for unit test in fuzzer/run.go, check once again

	if runtime.BoolDebug {
		fmt.Printf("Check bugs after 2 minutes\n")
	}

	for len(runtime.VecCheckEntry) > 0 {
		checkEntry := runtime.DequeueCheckEntry()
		if runtime.BoolDebug {
			print("Dequeueing:")
			for _, C := range checkEntry.CS {
				if ch, ok := C.(*runtime.ChanInfo); ok {
					print("\t", ch.StrDebug)
				}
			}
			println()
		}
		if atomic.LoadUint32(&checkEntry.Uint32NeedCheck) == 1 {
			if runtime.CheckBlockBug(checkEntry.CS) == false { // CS needs to be checked again in the future
			}
		}
	}
	// print bug info
	str, foundBug := runtime.TmpDumpBlockingInfo()
	if foundBug {
		fmt.Println(str)
	}
}

// When unit test ends, do all delayed bug detect, and wait for the checking process to end
func CheckBugEnd(entry *OracleEntry) {
	if runtime.BoolDelayCheck {
		runtime.SetCurrentGoCheckBug()
		if runtime.BoolDebug {
			println("End of unit test. Check bugs")
		}
		close(chEnforceCheck)
		entry.WgCheckBug.Wait() // let's not use send of channel, to make the code clearer
		// print bug info
		str, foundBug := runtime.TmpDumpBlockingInfo()
		if foundBug {
			fmt.Println(str)
		}
	}
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

func AfterRun(entry *OracleEntry) {
	switch StrTestMod {
	case "TestOnce": // Run all unit tests once, and print a file containing each test's name, # of select visited
		AfterRunTestOnce(entry)
	default: // Normal fuzzing
		AfterRunFuzz(entry)
	}
}

func AfterRunTestOnce(entry *OracleEntry) {
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

func AfterRunFuzz(entry *OracleEntry) {
	PrintNumTimeoutSelect() // Print the number of selects, number of timeout selects and not in input selects

	// if this is the first run, create input file using runtime's global variable
	if BoolFirstRun {
		CreateInput()
	}

	// create output file using runtime's global variable
	CreateRecordFile()

	CheckBugEnd(entry)

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

// Only enables oracle
func LightAfterRun(entry *OracleEntry) {
	CheckBugEnd(entry)
}