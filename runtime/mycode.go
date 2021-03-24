package runtime

func GoID() int64 {
	return getg().goid
}

type mystruct struct {
	mpGoID2Bytes map[int64][]byte
}

var mys mystruct

var ReportedPlace map[string]struct{}

var muMap mutex

func init() {

	mys.mpGoID2Bytes = make(map[int64][]byte)
	ReportedPlace = make(map[string]struct{})
	MapFirstInput = make(map[string]SelectInput)
	MapInput = make(map[string]SelectInput)
	StructRecord = Record{
		MapTupleRecord: make(map[string]uint16),
		MapChanRecord:  make(map[*hchan]*ChanRecord),
	}
}

var FlagSkipMyCode = false

func BeforeBlock() {
	if FlagSkipMyCode {
		return
	}
	gid := GoID()
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	//if Index(string(buf), "mutex.go:75") >= 0 {
	//	buf = buf[:Stack(buf, true)]
	//}
	lock(&muMap)
	mys.mpGoID2Bytes[gid] = buf
	_ = gid
	unlock(&muMap)
}



func AfterBlock() {
	if FlagSkipMyCode {
		return
	}
	gid := GoID()
	lock(&muMap)
	delete(mys.mpGoID2Bytes, gid)
	_ = gid
	unlock(&muMap)
}

func DumpBlockingInfo() (retStr string, foundBug bool) {
	retStr = ""
	foundBug = false
	if FlagSkipMyCode {
		return
	}
	SleepMS(1000)
	lock(&muMap)
	outer:
	for gid, sliceByte := range mys.mpGoID2Bytes {
		if gid != 1 { // No need to print the main goroutine
			str := string(sliceByte)
			switch true {
			case Index(str, "tools/cache/shared_informer.go:628") >= 0:
				continue
			case Index(str, "k8s.io/client-go/tools/cache/reflector.go:373") >= 0:
				continue
			case Index(str, "k8s.io/kubernetes/vendor/k8s.io/klog/v2/klog.go:1169") >= 0:
				continue
			case Index(str, "testing.go") >= 0:
				continue
			case Index(str, "k8s.io/apimachinery/pkg/watch/mux.go:247") >= 0:
				continue
			case Index(str, "k8s.io/apimachinery/pkg/util/wait/wait.go:167") >= 0:
				continue
			case Index(str, "scheduler/internal/cache/debugger/debugger.go:63") >= 0:
				continue
			case Index(str, "testing.go") >= 0:
				continue
			case Index(str, "k8s.io/client-go/tools/cache/shared_informer.go:772") >= 0:
				continue
			case Index(str, "vendor/k8s.io/client-go/tools/record/event.go:301") >= 0:
				continue
			case Index(str, "vendor/k8s.io/client-go/tools/cache/shared_informer.go:742 ") >= 0:
				continue
			case Index(str, "k8s.io/client-go/tools/cache/reflector.go:463 ") >= 0:
				continue
			case Index(str, "/vendor/") >= 0 && Index(str, "Lock(") == -1:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			case Index(str, "=====") >= 0:
				continue
			default:
			}

			stackSingleGo := ParseStackStr(str)
			if len(stackSingleGo.VecFuncName) == 0 {
				retStr += "Warning in DumpBlockingInfo: empty VecFunc*\n"
				continue
			}
			firstFuncName := stackSingleGo.VecFuncName[0]

			if Index(firstFuncName, "runtime.BeforeBlock") > -1 {
				// get the next func
				if len(stackSingleGo.VecFuncFile) < 2 || len(stackSingleGo.VecFuncLine) < 2 { // unexpected problem: no func after BeforeBlock
					retStr += "Warning in DumpBlockingInfo: no func after BeforeBlock\n"
					continue outer
				}
				nextFuncFile := stackSingleGo.VecFuncFile[1]
				nextFuncLine := stackSingleGo.VecFuncLine[1]


				if nextFuncFile != "/usr/local/go/src/sync/mutex.go" {
					// case 1: from channel op
					if _, reported := ReportedPlace[nextFuncFile + nextFuncLine]; reported {
						continue outer
					} else {
						ReportedPlace[nextFuncFile + nextFuncLine] = struct{}{}
					}
				} else {
					// case 2: from Lock op
					if len(stackSingleGo.VecFuncFile) < 3 || len(stackSingleGo.VecFuncLine) < 3 { // unexpected problem: no func after Lock
						retStr += "Warning in DumpBlockingInfo: no func after mutex.Lock\n"
						continue outer
					}
					nextFuncFile := stackSingleGo.VecFuncFile[2]
					nextFuncLine := stackSingleGo.VecFuncLine[2]
					if _, reported := ReportedPlace[nextFuncFile + nextFuncLine]; reported {
						continue outer
					} else {
						ReportedPlace[nextFuncFile + nextFuncLine] = struct{}{}
					}
				}

			} else {
				retStr += "Warning in DumpBlockingInfo: the first func is not BeforeBlock\n"
			}


			retStr += str + "\n"
			foundBug = true

		}
	}
	unlock(&muMap)
	return
}