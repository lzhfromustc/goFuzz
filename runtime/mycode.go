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
}

var FlagSkipMyCode = false

func TimerTest() {
	for i := 0; i < 5; i ++ {
		print("Current time:", nanotime() / 1000 / 1000, "\n")
		SleepMS(500)
	}
}

func BeforeBlock() {
	if FlagSkipMyCode {
		return
	}
	gid := GoID()
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	if Index(string(buf), "mutex.go:75") >= 0 {
		buf = buf[:Stack(buf, true)]
	}
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

func DumpBlockingInfo() {
	if FlagSkipMyCode {
		return
	}
	lock(&muMap)
	outer:
	for gid, sliceByte := range mys.mpGoID2Bytes {
		if gid != 1 { // No need to print the main goroutine
			str := string(sliceByte)
			//switch true {
			//case Index(str, "tools/cache/shared_informer.go:628") >= 0:
			//	continue
			//case Index(str, "k8s.io/client-go/tools/cache/reflector.go") >= 0:
			//	continue
			//case Index(str, "k8s.io/kubernetes/vendor/k8s.io/klog/v2/klog.go:1169") >= 0:
			//	continue
			//case Index(str, "testing.go") >= 0:
			//	continue
			//case Index(str, "k8s.io/apimachinery/pkg/watch/mux.go:247") >= 0:
			//	continue
			//case Index(str, "k8s.io/apimachinery/pkg/util/wait/wait.go:167") >= 0:
			//	continue
			//case Index(str, "scheduler/internal/cache/debugger/debugger.go:63") >= 0:
			//	continue
			//case Index(str, "testing.go") >= 0:
			//	continue
			//default:
			//}

			stackSingleGo := ParseStackStr(str)
			if len(stackSingleGo.VecFuncName) == 0 {
				print("Warning in DumpBlockingInfo: empty VecFunc*")
				continue
			}
			firstFuncName := stackSingleGo.VecFuncName[0]

			if Index(firstFuncName, "runtime.BeforeBlock") > -1 {
				// get the next func
				if len(stackSingleGo.VecFuncFile) < 2 || len(stackSingleGo.VecFuncLine) < 2 { // unexpected problem: no func after BeforeBlock
					print("Warning in DumpBlockingInfo: no func after BeforeBlock")
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
						print("Warning in DumpBlockingInfo: no func after mutex.Lock")
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
				print("Warning in DumpBlockingInfo: the first func is not BeforeBlock")
			}


			print(str,"\n")

		}
	}
	unlock(&muMap)
}

