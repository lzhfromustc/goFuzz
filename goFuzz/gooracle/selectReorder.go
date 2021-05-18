package gooracle

import (
	"runtime"
	"strconv"
	"time"
)

func SelectTimeout() <- chan time.Time {
	// if this channel wins, remember to call "runtime.StoreLastMySwitchChoice(-1)", which means we will use the original select
	return time.After(time.Duration(SelectDelayMS) * time.Millisecond)
}

func StoreLastMySwitchChoice(choice int) {
	runtime.StoreLastMySwitchChoice(choice)
}

func ReadSelect(strFileName string, lineOriSelect int, intOriSelectNumCase int) int {
	strLine := strconv.Itoa(lineOriSelect)
	runtime.StoreLastMySwitchLineNum(strLine) // store to runtime.g.lastMySwitchLineNum
	runtime.StoreLastMySwitchSelectNumCase(intOriSelectNumCase)
	input, exist := MapInput[strFileName + ":" +strLine]
	if exist {
		runtime.StoreLastMySwitchChoice(input.IntPrioCase)
		return input.IntPrioCase
	} else {
		runtime.StoreLastMySwitchChoice(-1)
		return -1 // let switch choose the default case
	}
}
