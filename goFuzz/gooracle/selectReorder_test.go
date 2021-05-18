package gooracle

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestGLDetection(t *testing.T) {

	runtime.BoolRecordPerCh = true
	runtime.RecordSelectChoice = true

	StrTestpath = "/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/gooracle"

	StrOutputFullPath := "/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/gooracle/myoutput.txt"
	out, err := os.OpenFile(StrOutputFullPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open file:", FileNameOfOutput())
		return
	}
	OutputFile = out
	os.Stdout = out

	MapInput = make(map[string]runtime.SelectInfo)
	MapInput["/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/gooracle/selectReorder_test.go:42"] = runtime.SelectInfo{
		StrFileName: "/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/gooracle/selectReorder_test.go",
		StrLineNum:  "42",
		IntNumCase:  0,
		IntPrioCase: 0,
	}
	SelectDelayMS = 1000

	ch := make(chan int)

	go func() {
		ch <- 1
	}()
	switch ReadSelect("/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/gooracle/selectReorder_test.go", 42, 2) {
	case 0:
		select {
		case <-ch:
			fmt.Println("Normal")
		case <-SelectTimeout():
			StoreLastMySwitchChoice(-1)
			select {
			case <-ch:
				fmt.Println("Normal")
			case <-time.After(200 * time.Millisecond):
				fmt.Println("Should be buggy")
			}
		}
	case 1:
		select {
		case <-time.After(200 * time.Millisecond):
			fmt.Println("Should be buggy")
		case <-SelectTimeout():
			StoreLastMySwitchChoice(

				// print bug info
				-1)
			select {
			case <-ch:
				fmt.Println("Normal")
			case <-time.After(200 * time.Millisecond):
				fmt.Println("Should be buggy")
			}
		}
	default:
		StoreLastMySwitchChoice(-1)
		select {
		case <-ch:
			fmt.Println("Normal")
		case <-time.After(200 * time.Millisecond):
			fmt.Println("Should be buggy")
		}
	}

	CreateInput()

	CreateRecordFile()

	str, foundBug := runtime.TmpDumpBlockingInfo()
	if foundBug {
		fmt.Println("-----New Bug:")
		fmt.Println(str)
	}
	CloseOutputFile()

}
