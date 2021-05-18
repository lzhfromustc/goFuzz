package runtime


var MapSelectInfo map[string]SelectInfo // useful only when RecordSelectChoice is true
//var MapInput map[string]SelectInfo // useful only when RecordSelectChoice is false
var RecordSelectChoice bool = false
var MuFirstInput mutex

type SelectInfo struct {
	StrFileName string
	StrLineNum string
	IntNumCase int
	IntPrioCase int
}

func StoreSelectInput(intNumCase, intPrioCase int) {
	lock(&MuFirstInput)
	newSelectInput := NewSelectInputFromRuntime(intNumCase, intPrioCase, 3)
	MapSelectInfo[newSelectInput.StrFileName + ":" + newSelectInput.StrLineNum] = newSelectInput
	unlock(&MuFirstInput)
}

func NewSelectInputFromRuntime(intNumCase, intPrioCase int, intLayerCallee int) SelectInfo {
	// if A contains select, select calls StoreSelectInput, StoreSelectInput calls this function, then intLayerCallee is 3
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncLine) < intLayerCallee {
		return SelectInfo{}
	}
	selectInput := SelectInfo{
		StrFileName: stackSingleGo.VecFuncFile[intLayerCallee - 1], // where is the select
		StrLineNum:  LastMySwitchLineNum(),
		IntNumCase:  LastMySwitchOriSelectNumCase(),
		IntPrioCase: -1,
	}

	if LastMySwitchChoice() == -1 {
		// Executing the original select
		selectInput.IntPrioCase = intPrioCase
	} else {
		// Executing our select, so the chosen case is same to switch's choice
		selectInput.IntPrioCase = LastMySwitchChoice()
	}

	return selectInput
}

func LastMySwitchLineNum() string {
	if getg().lastMySwitchLineNum != "" {
		return getg().lastMySwitchLineNum
	} else {
		return "0"
	}
}

func LastMySwitchOriSelectNumCase() int {
	return getg().lastMySwitchOriSelectNumCase
}

func LastMySwitchChoice() int {
	return getg().lastMySwitchChoice
}

func StoreLastMySwitchSelectNumCase(numCase int) {
	getg().lastMySwitchOriSelectNumCase = numCase
}

func StoreLastMySwitchChoice(choice int) {
	getg().lastMySwitchChoice = choice
}

func StoreLastMySwitchLineNum(strLine string) {
	getg().lastMySwitchLineNum = strLine // no need for synchronization.
}
//
//func TimePassedSince(begin int64, duration int64) bool {
//	now := nanotime()
//	if now - begin > duration {
//		return true
//	} else {
//		return false
//	}
//}
//
//func ChangeSelect() (needChange bool, indexPrioCase int) {
//	if BoolChangeSelect == false {
//		return
//	}
//	const size = 64 << 10
//	buf := make([]byte, size)
//	buf = buf[:Stack(buf, false)]
//	strStack := string(buf)
//	stackSingleGo := ParseStackStr(strStack)
//	if len(stackSingleGo.VecFuncFile) < 2 {
//		println("less 2")
//		return false, -1
//	}
//	secondFuncStr := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
//	selectInput, exist := MapInput[secondFuncStr]
//	//print("Second func:", secondFuncStr, "\n")
//	if exist {
//		return true, selectInput.IntPrioCase
//	} else {
//		return false, -1
//	}
//}
