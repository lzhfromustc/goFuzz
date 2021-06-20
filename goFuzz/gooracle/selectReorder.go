package gooracle

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var uint32ReadSelectCount uint32
var uint32OriginalSelectCount uint32
var uint32SelectNotInInput uint32

func SelectTimeout() <- chan time.Time {
	// if this channel wins, remember to call "runtime.StoreLastMySwitchChoice(-1)", which means we will use the original select
	return time.After(time.Duration(SelectDelayMS) * time.Millisecond)
}

func StoreLastMySwitchChoice(choice int) {
	if choice == -1 {
		atomic.AddUint32(&uint32OriginalSelectCount, 1)
	}
	runtime.StoreLastMySwitchChoice(choice)
}

func ReadSelect(strFileName string, lineOriSelect int, intOriSelectNumCase int) int {
	atomic.AddUint32(&uint32ReadSelectCount, 1)
	strLine := strconv.Itoa(lineOriSelect)
	runtime.StoreLastMySwitchLineNum(strLine) // store to runtime.g.lastMySwitchLineNum
	runtime.StoreLastMySwitchSelectNumCase(intOriSelectNumCase)
	input, exist := MapInput[strFileName + ":" +strLine]
	if exist {
		runtime.StoreLastMySwitchChoice(input.IntPrioCase)
		return input.IntPrioCase
	} else {
		atomic.AddUint32(&uint32SelectNotInInput, 1)
		runtime.StoreLastMySwitchChoice(-1)
		return -1 // let switch choose the default case
	}
}

// Should be called at the end of a program
func PrintNumTimeoutSelect() {
	intToTalSelect := int(uint32ReadSelectCount) // no need to use atomic
	// The number of ReadSelect is equal to the number of original select executed
	intNotInInputSelect := int(uint32SelectNotInInput)
	// The number of selects that are not in input (probably not instrumented) and we just execute the original select
	intTimeOutSelect := int(uint32OriginalSelectCount - uint32SelectNotInInput)
	// uint32OriginalSelectCount is the number of selects that we execute the original code. There are possibilities:
	// (1) the select is not in input (uint32SelectNotInInput); (2) the select is in input and we tried to force
	// it choose one case, but it timeouts and goes back to the original select

	fmt.Printf("====Total number of selects:%d\tTimeout selects:%d\tNot in input selects:%d\n", intToTalSelect, intTimeOutSelect, intNotInInputSelect)
}
