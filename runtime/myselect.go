package runtime

var MapFirstInput map[string]SelectInput // useful only when GenFirstInput is true
var MapInput map[string]SelectInput // useful only when GenFirstInput is false
var GenFirstInput bool = false
var MuFirstInput mutex

var SelectDelayMS int
type SelectInput struct {
	StrFileName string
	StrLineNum string
	IntNumCase int
	IntPrioCase int
}

var BoolChangeSelect bool = true

func StoreSelectInput(intNumCase, intPrioCase int) {
	lock(&MuFirstInput)
	newSelectInput := NewSelectInputFromRuntime(intNumCase, intPrioCase, 3)
	MapFirstInput[newSelectInput.StrFileName + ":" + newSelectInput.StrLineNum] = newSelectInput
	unlock(&MuFirstInput)
}

func NewSelectInputFromRuntime(intNumCase, intPrioCase int, intLayerCallee int) SelectInput {
	// if A contains select, select calls StoreSelectInput, StoreSelectInput calls this function, then intLayerCallee is 3
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncLine) < intLayerCallee {
		return SelectInput{}
	}
	selectInput := SelectInput{
		StrFileName: stackSingleGo.VecFuncFile[intLayerCallee - 1], // where is the select
		StrLineNum:  stackSingleGo.VecFuncLine[intLayerCallee - 1],
		IntNumCase:  intNumCase,
		IntPrioCase: intPrioCase,
	}
	return selectInput
}

func TimePassedSince(begin int64, duration int64) bool {
	now := nanotime()
	if now - begin > duration {
		return true
	} else {
		return false
	}
}

func ChangeSelect() (needChange bool, indexPrioCase int) {
	if BoolChangeSelect == false {
		return
	}
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncFile) < 2 {
		println("less 2")
		return false, -1
	}
	secondFuncStr := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
	selectInput, exist := MapInput[secondFuncStr]
	//print("Second func:", secondFuncStr, "\n")
	if exist {
		return true, selectInput.IntPrioCase
	} else {
		return false, -1
	}
}
