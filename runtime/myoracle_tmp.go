package runtime

// A temporary oracle for fuzzer

type mystruct struct {
	mpGoID2Bytes map[int64][]byte
}

var mys mystruct

var ReportedPlace map[string]struct{}

var muMap mutex

func init() {

	mys.mpGoID2Bytes = make(map[int64][]byte)
	ReportedPlace = make(map[string]struct{})
	MapSelectInfo = make(map[string]SelectInfo)
	//MapInput = make(map[string]SelectInput)
}

var BoolDebug = true

func TmpBeforeBlock() {
	if !BoolDebug {
		return
	}
	gid := GoID()
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	lock(&muMap)
	mys.mpGoID2Bytes[gid] = buf
	unlock(&muMap)
}



func TmpAfterBlock() {
	if !BoolDebug {
		return
	}
	gid := GoID()
	lock(&muMap)
	delete(mys.mpGoID2Bytes, gid)
	unlock(&muMap)
}

func TmpDumpBlockingInfo() (retStr string, foundBug bool) {
	retStr = ""
	foundBug = false
	if !BoolDebug {
		return
	}
	SleepMS(500)
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
			case Index(str, "k8s.io/klog:1169") >= 0:
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
				retStr += "Warning in TmpDumpBlockingInfo: empty VecFunc*\n"
				continue
			}
			firstFuncName := stackSingleGo.VecFuncName[0]

			if Index(firstFuncName, "runtime.TmpBeforeBlock") > -1 {
				// get the next func
				if len(stackSingleGo.VecFuncFile) < 2 || len(stackSingleGo.VecFuncLine) < 2 { // unexpected problem: no func after TmpBeforeBlock
					retStr += "Warning in TmpDumpBlockingInfo: no func after TmpBeforeBlock\n"
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
					if indexSDK := Index(nextFuncFile, "/usr/local/go/src"); indexSDK > -1 { // In SDK, don't report
						continue outer
					}
				} else {
					// case 2: from Lock op
					if len(stackSingleGo.VecFuncFile) < 3 || len(stackSingleGo.VecFuncLine) < 3 { // unexpected problem: no func after Lock
						retStr += "Warning in TmpDumpBlockingInfo: no func after mutex.Lock\n"
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
				retStr += "Warning in TmpDumpBlockingInfo: the first func is not TmpBeforeBlock\n"
			}
			// delete the BeforeBlock function
			for {
				indexTBB := Index(str, "runtime.TmpBeforeBlock()\n\t/usr/local/go/src/runtime/myoracle_tmp.go:")
				if indexTBB == -1 {
					break
				} else {
					str = str[:indexTBB] + str[indexTBB + 77:]
				}
			}

			retStr += "-----New Bug:\n" + str + "\n"
			print(retStr)
			foundBug = true

		}
	}
	unlock(&muMap)
	return
}