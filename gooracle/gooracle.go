package gooracle

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

//// Part 1.1: data struct for each channel

// ChanInfo is 1-to-1 with every channel. It tracks a list of goroutines that hold the reference to the channel
type ChanInfo struct {
	Chan            interface{} // Stores the channel. Can be used as ID of channel
	intBuffer       int // The buffer capability of channel. 0 if channel is unbuffered
	mapRefGoroutine sync.Map // Stores all goroutines that still hold reference to this channel
	strDebug        string
	intFlagFoundBug int32 // Use atomic int32 operations to mark if a bug is reported
}

var MapChan2ChanInfo sync.Map // Help us retrieve ChanInfo with a given channel
var NotTracingChan interface{} // This is a fake channel. We use it in the place of channels that we are not interested in
var NotTracingChanInfo = &ChanInfo{}
var DefaultCaseChan interface{}
var DefaultCaseChanInfo = &ChanInfo{}

// Initialize a new ChanInfo with a given channel
func NewChanInfo(ch interface{}) *ChanInfo {
	newChInfo := &ChanInfo{
		Chan:            ch,
		intBuffer:       reflect.ValueOf(ch).Cap(),
		mapRefGoroutine: sync.Map{},
		strDebug:        "", // TODO: add information like filename and line number here
		intFlagFoundBug: 0,
	}

	MapChan2ChanInfo.Store(ch, newChInfo)

	//Log.mu.Lock()
	//defer Log.mu.Unlock()
	//
	//_, strFileName, intLine, ok := runtime.Caller(1)
	//if !ok {
	//	fmt.Println("Fatal error in NewGoroutine: can't find the caller")
	//	return newChInfo
	//}
	//strPosition := strFileName + ":" + strconv.Itoa(intLine)
	////Log.MapChPtr2Site[ch] = strPosition
	//newChInfo.strDebug = strPosition

	return newChInfo
}

// FindChanInfo can retrieve a initialized ChanInfo for a given channel
func FindChanInfo(ch interface{}) *ChanInfo {
	if ch == NotTracingChan {
		return NotTracingChanInfo
	}

	if chInfo, ok := MapChan2ChanInfo.Load(ch); ok {
		if concrete, ok := chInfo.(*ChanInfo); ok {
			return concrete
		}
	}

	fmt.Println("Warning in FindChanInfo: no ChanInfo found for ch")
	return nil
}

// After the creation of a new channel, or at the head of a goroutine that holds a reference to a channel,
// or whenever a goroutine obtains a reference to a channel, call this function
// AddRefGoroutine links a channel with a goroutine, meaning the goroutine holds the reference to the channel
func AddRefGoroutine(chInfo *ChanInfo, goInfo *GoInfo) {
	chInfo.AddGoroutine(goInfo)
	goInfo.AddChan(chInfo)
}

// This means the goroutine mapped with goInfo holds the reference to chInfo.Chan
func (chInfo *ChanInfo) AddGoroutine(goInfo *GoInfo) {
	chInfo.mapRefGoroutine.Store(goInfo, true)
}

// At the end of the lifetime of a channel, call this function.
// At the end of a goroutine, GoInfo.RemoveRef() will be called, and it will automatically call this function for every
// channel it hold a reference to.
// RemoveGoroutine means this goroutine will no longer access the channel
func RemoveRefGoroutine(chInfo *ChanInfo, goInfo *GoInfo) {
	chInfo.mapRefGoroutine.Delete(goInfo)
	goInfo.mapChanInfo.Delete(chInfo)
}
//func (chInfo *ChanInfo) RemoveGoroutine(goInfo *GoInfo) {
//	RemoveRefGoroutine()
//	chInfo.mapRefGoroutine.Delete(goInfo)
//}

// A blocking bug is detected, if all goroutines that hold the reference to a channel are blocked at an operation of the channel
// TODO: Improve this to detect circular wait
func (chInfo *ChanInfo) CheckBlockBug() {
	if atomic.LoadInt32(&chInfo.intFlagFoundBug) != 0 {
		return
	}

	if chInfo.intBuffer == 0 {
		countRefGo := 0 // Number of goroutines that hold the reference to the channel
		countBlockAtThisChanGo := 0 // Number of goroutines that are blocked at an operation of this channel
		f := func(key interface{}, value interface{}) bool {
			goInfo, _ := key.(*GoInfo)

			boolIsBlock, _ := goInfo.IsBlockAtGivenChan(chInfo)
			if boolIsBlock {
				countBlockAtThisChanGo++
			}
			countRefGo++
			return true // continue Range
		}
		chInfo.mapRefGoroutine.Range(f)

		if countRefGo == countBlockAtThisChanGo {
			if countRefGo == 0 {
				// debug code
				countRefGo2 := 0 // Number of goroutines that hold the reference to the channel
				countBlockAtThisChanGo2 := 0 // Number of goroutines that are blocked at an operation of this channel
				f := func(key interface{}, value interface{}) bool {
					goInfo, _ := key.(*GoInfo)

					boolIsBlock, _ := goInfo.IsBlockAtGivenChan(chInfo)
					if boolIsBlock {
						countBlockAtThisChanGo2++
					}
					countRefGo2++
					return true // continue Range
				}
				chInfo.mapRefGoroutine.Range(f)
				fmt.Print()

				return
			}
			ReportBug(chInfo)
			atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
		}

	} else { // Buffered channel
		if reflect.ValueOf(chInfo.Chan).Len() == chInfo.intBuffer { // Buffer is full
			// Check if all ref goroutines are blocked at send
			boolAllBlockAtSend := true
			countRefGo := 0
			countBlockAtThisChanGo := 0
			f := func(key interface{}, value interface{}) bool {
				goInfo, _ := key.(*GoInfo)

				boolIsBlock, strOp := goInfo.IsBlockAtGivenChan(chInfo)
				if boolIsBlock {
					countBlockAtThisChanGo++
				}
				if strOp != Send {
					boolAllBlockAtSend = false
				}
				countRefGo++
				return true // continue Range
			}
			chInfo.mapRefGoroutine.Range(f)

			if countRefGo == countBlockAtThisChanGo && boolAllBlockAtSend {
				ReportBug(chInfo)
				atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
			}

		} else if reflect.ValueOf(chInfo.Chan).Len() == 0 { // Buffer is empty
			// Check if all ref goroutines are blocked at receive
			boolAllBlockAtRecv := true
			countRefGo := 0
			countBlockAtThisChanGo := 0
			f := func(key interface{}, value interface{}) bool {
				goInfo, _ := key.(*GoInfo)

				boolIsBlock, strOp := goInfo.IsBlockAtGivenChan(chInfo)
				if boolIsBlock {
					countBlockAtThisChanGo++
				}
				if strOp != Recv {
					boolAllBlockAtRecv = false
				}
				countRefGo++
				return true // continue Range
			}
			chInfo.mapRefGoroutine.Range(f)

			if countRefGo == countBlockAtThisChanGo && boolAllBlockAtRecv {
				ReportBug(chInfo)
				atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
			}

		} else { // Buffer is not full or empty. Then it is not possible to block
			// do nothing
		}
	}
}

var PrintBug bool = true// This doesn't influnece whether we detect the bug

func ReportBug(chInfo *ChanInfo) {
	countRefGo := 0
	f := func(key interface{}, value interface{}) bool {
		countRefGo++
		return true // continue Range
	}
	chInfo.mapRefGoroutine.Range(f)

	if PrintBug {
		fmt.Println("A blocking bug is found for channel: ", chInfo.strDebug)
		fmt.Printf("\tThere are %d goroutines blocked at this channel\n", countRefGo)
		debug.PrintStack()
		panic(1)
	}
}





//// Part 1.2: data struct for each goroutine

// GoInfo is 1-to-1 with each goroutine.
// Go language doesn't allow us to acquire the ID of a goroutine, because they want goroutines to be anonymous.
// Normally, Go programmers use runtime.Stack() to print all IDs of all goroutines, but this function is very inefficient
//, since it calls stopTheWorld()
// Currently we use a global atomic int64 to differentiate each goroutine, and a variable currentGo to represent each goroutine
// This is not a good practice because the goroutine need to pass currentGo to its every callee
type GoInfo struct {
	ID          int64
	//BlockAtOp   *OpInfo // When the goroutine is about to be blocked at an operation, store the operation. Nil when normally running
	BlockMap map[*ChanInfo]string // Nil when normally running. When blocked at an operation of ChanInfo, store
	// one ChanInfo and the operation. When blocked at select, store multiple ChanInfo and
	// operation. Default in select is also also stored in map, which is DefaultCaseChanInfo
	mapChanInfo sync.Map // Stores all channels that this goroutine still hold reference to


	Logger GoLogger
}

type GoLogger struct {
	StrCreationSite string // Where the goroutine is created. Use filename+":"+strconv.Itoa(linenumber)
	VecOp []LogOp // A list of executed operations with order
}


type LogOp interface {
	Log() string
}

type LogSendRecvClose struct {
	Ch string // Used to differentiate channels created at the same line in source code
	StrOp string
}

// TODO: modify Log()
func (l LogSendRecvClose) Log() string {
	return "" + "_" +l.StrOp
}

type LogSelect struct {
	BoolHasDefault bool
	VecCh []string // The order of this slice is the same as the order of select cases. Default is DefaultCaseChanInfo
	VecOp []string
	ChChosen interface{}
}

// TODO: modify Log()
func (l LogSelect) Log() string {
	result := ""
	//for _, chInfo := range l.VecCh {
	//	result += l.MapCh2Str[chInfo][0] + "_" + l.MapCh2Str[chInfo][1]
	//}
	//result += "|" + l.MapCh2Str[l.ChChosen][0] + "_" + l.MapCh2Str[l.ChChosen][1]

	return result
}

type LogSleep struct {
	StrSleepSite string // Where is the sleep. Use filename+":"+strconv.Itoa(linenumber)
	SleepTime int64
}

// TODO: modify Log()
func (l LogSleep) Log() string {
	//return l.StrSleepSite + strconv.Itoa(int(l.SleepTime))
	return ""
}

var MapID2GoInfo sync.Map

var intGoID int64

func GoID() int64 {
	id := runtime.GoID()
	return id
}

// Assign a new GoID and initialize a GoInfo
func NewGoroutine() *GoInfo {
	id := GoID()
	newGoInfo := &GoInfo{
		ID:          id,
		//BlockAtOp:   nil,
		BlockMap:    nil,
		mapChanInfo: sync.Map{},
	}

	//Log.mu.Lock()
	//defer Log.mu.Unlock()
	//_, strFileName, intLine, ok := runtime.Caller(1)
	//if !ok {
	//	fmt.Println("Fatal error in NewGoroutine: can't find the caller")
	//	return newGoInfo
	//}
	//strPosition := strFileName + ":" + strconv.Itoa(intLine)
	//Log.MapGoID2Site[id] = strPosition
	//newGoInfo.Logger.StrCreationSite = strPosition

	return newGoInfo
}

func CurrentGoInfo() *GoInfo {
	return FindGoInfo(GoID())
}

func FindGoInfo(id int64) *GoInfo {
	goInfo, ok := MapID2GoInfo.Load(id)
	if ok {
		if concrete, ok := goInfo.(*GoInfo); ok {
			return concrete
		} else {
			fmt.Println("Warning in FindGoInfo: the GoInfo loaded is of incorrect type. Id:", id)
		}
	} else {
		fmt.Println("Warning in FindGoInfo: no GoInfo found for the given id:", id)
	}

	return nil
}

// This means the goroutine mapped with goInfo holds the reference to chInfo.Chan
func (goInfo *GoInfo) AddChan(chInfo *ChanInfo) {
	goInfo.mapChanInfo.Store(chInfo, true)
}

// RemoveRef should be called at the end of every goroutine. It will remove goInfo from the reference list of every
// channel it holds the reference to
func (goInfo *GoInfo) RemoveAllRef() {
	f := func(key interface{}, value interface{}) bool {
		chInfo, _ := key.(*ChanInfo)
		RemoveRefGoroutine(chInfo, goInfo)

		chInfo.CheckBlockBug()

		goInfo.mapChanInfo.Delete(key)
		return true // Continue Range()
	}
	goInfo.mapChanInfo.Range(f)
}


// SetBlockAt should be called before each channel operation, meaning the current goroutine is about to execute that operation
// Note that we check bug in this function, because it's possible for the goroutine to be blocked forever if it execute that operation
// For example, a channel with no buffer is held by a parent and a child.
//              The parent has already exited, but the child is now about to send to that channel.
//              Then now is our only chance to detect this bug, so we call CheckBlockBug()
func (goInfo *GoInfo) SetBlockAt(ch interface{}, strOp string) {
	goInfo.BlockMap = make(map[*ChanInfo]string)
	goInfo.BlockMap[FindChanInfo(ch)] = strOp
}

// WithdrawBlock should be called after each channel operation, meaning the current goroutine finished execution that operation
// If the operation is select, remember to call this function right after each case of the select
func (goInfo *GoInfo) WithdrawBlock() {
	goInfo.BlockMap = nil
}

func (goInfo *GoInfo) IsBlock() (boolIsBlock bool, strOp string) {
	boolIsBlock, strOp = false, ""
	boolIsSelect, boolIsBlockingSelect := false, true

	mapBlock := goInfo.BlockMap
	if mapBlock == nil {
		return
	} else {
		boolIsBlock = true
	}

	// Now we compute strOp

	if len(mapBlock) > 1 {
		boolIsSelect = true
	} else if len(mapBlock) == 0 {
		fmt.Println("Fatal in GoInfo.IsBlock(): mapBlock is not nil but lenth is 0")
	}

	if boolIsSelect {
		for chanInfo, _ := range mapBlock {
			if chanInfo == DefaultCaseChanInfo {
				boolIsBlockingSelect = false
				break
			}
		}

		if boolIsBlockingSelect {
			strOp = BSelect
		} else {
			strOp = NBSelect
		}
	} else {
		for _, op := range mapBlock { // This loop will be executed only one time, since mapBlock's len() is 1
			strOp = op
		}
	}

	return
}

// This function checks if the goroutine mapped with goInfo is currently blocking at an operation of chInfo.Chan
// If so, returns true and the string of channel operation
func (goInfo *GoInfo) IsBlockAtGivenChan(chInfo *ChanInfo) (boolIsBlockAtGiven bool, strOp string) {
	boolIsBlockAtGiven, strOp = false, ""

	mapBlock := goInfo.BlockMap
	if mapBlock == nil {
		return
	}

	for chanInfo, op := range mapBlock {
		if chanInfo == chInfo {
			boolIsBlockAtGiven = true
			strOp = op
			break
		}
	}

	return
}



// Log file
type Logger struct {
	RunID int
	MapGoID2Site map[int64]string
	//MapChPtr2Site map[interface{}]string
	VecGoLogger []GoLogger

	mu sync.Mutex
}

var Log Logger

func InitLogger(IDLast int) {
	Log = Logger{
		RunID:         IDLast + 1,
		MapGoID2Site:  make(map[int64]string),
		//MapChPtr2Site: make(map[interface{}]string),
		VecGoLogger:     nil,
		mu:				sync.Mutex{},
	}
}

func PrintLogger() {
	jsonLogger, err := json.Marshal(Log)
	if err != nil {
		fmt.Println("Fatal in PrintLogger: ", err)
		return
	}
	fmt.Println(jsonLogger)
}

func (g *GoInfo) PrintLog() {
	Log.mu.Lock()
	defer Log.mu.Unlock()

	Log.VecGoLogger = append(Log.VecGoLogger, g.Logger)
}

func (g *GoInfo) RecordSimpleChanOp(chInfo *ChanInfo, strOp string) {
	newOpLog := LogSendRecvClose{
		Ch:    chInfo.strDebug,
		StrOp: strOp,
	}
	g.Logger.VecOp = append(g.Logger.VecOp, newOpLog)
}

func (g *GoInfo) RecordSleep(duration time.Duration) {
	_, strFileName, intLine, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Fatal error in RecordSleep: can't find the caller")
		return
	}
	strPosition := strFileName + ":" + strconv.Itoa(intLine)
	newOpLog := LogSleep{
		StrSleepSite: strPosition,
		SleepTime:    int64(duration),
	}
	g.Logger.VecOp = append(g.Logger.VecOp, newOpLog)
}

func (g *GoInfo) RecordSelect(strOp string) LogSelect {
	boolHasDefault := true
	if strOp == BSelect {
		boolHasDefault = false
	} else {
		boolHasDefault = true
	}
	newOpLog := LogSelect{
		BoolHasDefault: boolHasDefault,
		VecCh:          nil,
		VecOp:          nil,
		ChChosen:       nil,
	}

	return newOpLog
}

func (g *GoInfo) RecordSelectCase(l LogSelect, chInfo *ChanInfo, strOp string) {
	l.VecCh = append(l.VecCh, chInfo.strDebug)
	l.VecOp = append(l.VecOp, strOp)
}

func (g *GoInfo) RecordSelectChosen(l LogSelect, intIndex int) {
	l.ChChosen = l.VecCh[intIndex]
	g.Logger.VecOp = append(g.Logger.VecOp, l)
}




////// Part 1.3: data struct for each operation
//
//// OpInfo is 1-to-1 mapped with an operation. It is also 1-to-1 mapped with the goroutine that will execute it
//// However, if boolIsSelect is true, then it can be mapped to multiple channels
//type OpInfo struct {
//	ChanInfo     *ChanInfo  // nil if boolIsSelect is true
//	WhichOp      string
//	GoInfo       *GoInfo
//
//	// fields below are for select operation
//	boolIsSelect bool
//	mapSelectOp  map[*OpInfo]struct{} // not nil if boolIsSelect is true. This is a list of operations in the cases of select
//}

const (
	Send = "Send"
	Recv = "Recv"
	Close = "Close"
	BSelect = "BlockingSelect"
	NBSelect = "NonBlockingSelect"
)
//var NotTracingChanOp *OpInfo = &OpInfo{} // bonded with NotTracingChan. We use it in the place of operations that belong
//// to channels we are not interested in
//
//// Initalize a OpInfo for a send/receive/close operation
//func NewSimpleOpInfo(ch interface{}, goInfo *GoInfo, strOp string) *OpInfo {
//	newOpInfo := &OpInfo{
//		ChanInfo:     FindChanInfo(ch),
//		WhichOp:      strOp,
//		GoInfo:       goInfo,
//		boolIsSelect: false,
//		mapSelectOp:  nil,
//	}
//
//	return newOpInfo
//}
//
//// Initalize a OpInfo for a blocking or nonblocking select operation
//func NewSelectInfo(goInfo *GoInfo, strOp string) *OpInfo {
//	newOpInfo := &OpInfo{
//		ChanInfo:     nil,
//		WhichOp:      strOp,
//		GoInfo:       goInfo,
//		boolIsSelect: true,
//		mapSelectOp:  make(map[*OpInfo]struct{}),
//	}
//
//	return newOpInfo
//}
//
//// After NewSelectInfo, call AddCase for each of the case in select
//func (selectInfo *OpInfo) AddCase(ch interface{}, strOp string) {
//	if ch == NotTracingChan { // This case is about a channel we are not interested in
//		selectInfo.mapSelectOp[NotTracingChanOp] = struct{}{}
//		return
//	}
//
//	newOpInfo := NewSimpleOpInfo(ch, selectInfo.GoInfo, strOp)
//	selectInfo.mapSelectOp[newOpInfo] = struct{}{}
//}
//


//// Part 2.1: Original Benchmark for Goroutine Leak

// A benchmark that creates numCh channels, and numGoroutine child goroutines for each channel. numSyncGoroutine of the goroutines belong
// to channels we are interested in. Each child goroutine sleeps for timeSleep and then send to a channel.
// For each child goroutine, the parent goroutine will wait for it to send the message, or choose timeout case after timeOut.
func BenchGL(numCh int, numGoroutine int, numSyncGoroutine int, timeSleep, timeOut time.Duration) {

	vecCh := []chan bool{}
	for j := 0; j < numCh; j++ {
		newCh := make(chan bool)

		vecCh = append(vecCh, newCh)
	}

	vecNonSyncCh := []chan int{}
	for j := 0; j < numCh; j++ {
		newCh := make(chan int)
		vecNonSyncCh = append(vecNonSyncCh, newCh)
	}

	for j := 0; j < numCh; j++ {
		for i := 0; i < numGoroutine; i++ {

			if i < numSyncGoroutine {
				go func(j int) {
					time.Sleep(timeSleep)
					vecCh[j] <- true
				}(j)

			} else {
				go func(j int) {
					time.Sleep(timeSleep)
					vecNonSyncCh[j] <- 1
				}(j)
			}

		}
	}



	for j := 0; j < numCh; j++ {
		for i := 0; i < numGoroutine; i++ {

			select {
			case <-vecCh[j]:
			case <-vecNonSyncCh[j]:
			case <-time.After(timeOut):
			}
		}
	}
}


// Code to control time
func SelectWaitTime() time.Duration {
	return 30 * time.Second
}

// Code to read log and indicate which case has priority
func PriorityCaseIndex() int {
	runtime.Caller(1)
	return 1
}



//// Part 2.2: Benchmark monitored by Oracle described in Part 1

// Same as BenchGL, but we insert the Oracle code here to detect bug
func BenchGLDetect(numCh int, numGoroutine int, numSyncGoroutine int, timeSleep, timeOut time.Duration) {
	currentGo := NewGoroutine()		// initialize a GoInfo
	//defer func() {
	//	currentGo.PrintLog()
	//}()
	defer func() {
		currentGo.RemoveAllRef()		// In this benchmark, we know that at the end of this function, the parent goroutine should lose
		// the reference to all the channels. We need further implemenation to decide where
		// to call RemoveRef() for more general cases
	}()

	vecCh := []chan bool{}
	for j := 0; j < numCh; j++ {
		newCh := make(chan bool)
		newChInfo := NewChanInfo(newCh)			// After the creation of a channel we interested in, initialize a ChanInfo
		AddRefGoroutine(newChInfo, currentGo)

		vecCh = append(vecCh, newCh)
	}

	vecNonSyncCh := []chan int{}
	for j := 0; j < numCh; j++ {
		newCh := make(chan int)					// This is a channel we are not interested in, don't create ChanInfo
		vecNonSyncCh = append(vecNonSyncCh, newCh)
	}

	for j := 0; j < numCh; j++ {
		for i := 0; i < numGoroutine; i++ {

			if i < numSyncGoroutine {			// The first numSyncGoroutine goroutines are using channels we are interested in
				go func(j int) {
					currentGo := NewGoroutine()		// Note that currentGo here is a local variable in this child goroutine
					AddRefGoroutine(FindChanInfo(vecCh[j]), currentGo)		// This function should be called for every
					// channel this child goroutine holds reference to. In this benchmark, only vecCh[j]
					//defer func() {
					//	currentGo.PrintLog()
					//}()
					defer func() {
						currentGo.RemoveAllRef()		// When this child goroutine ends, it will not longer hold the reference to any channels
						// CheckBlockBug() is called inside
					}()

					time.Sleep(timeSleep)

					// The original "vecCh[j] <- true" is now accompanied with 3 API calls
					currentGo.SetBlockAt(vecCh[j], Send)			// CheckBlockBug() is called inside
					vecCh[j] <- true
					currentGo.WithdrawBlock()
				}(j)		// remember to pass j, to avoid data race

			} else {							// The other goroutines are using channels we are not interested in
				go func(j int) {
					time.Sleep(timeSleep)
					vecNonSyncCh[j] <- 1
				}(j)
			}

		}
	}



	for j := 0; j < numCh; j++ {
		for i := 0; i < numGoroutine; i++ {

			currentGo.SetBlockAt(vecCh[j], Recv)
			currentGo.SetBlockAt(NotTracingChan, Recv)		// Use NotTracingChan to replace vecNonSyncCh[j], because we are not interested in it
			select {
			case <-vecCh[j]:
				currentGo.WithdrawBlock()
			case <-vecNonSyncCh[j]:
				currentGo.WithdrawBlock()
			case <-time.After(timeOut):
				currentGo.WithdrawBlock()
			}
		}
	}
}

//// Part 2.3: Original Benchmark for Channel In Critical Section

// This benchmark creates numChan channels. For each channel, it creates 1 receiving goroutine, numSendGoroutine send
// goroutines, and (numGoroutine - numSendGoroutine - 1) normal goroutines. Every goroutine is protected by a MuCh which
// mimics the behavior of Mutex. Every send goroutine sleeps for timeSleep before the send operation. The parent
// goroutine waits for all receive goroutines to receive all messages from send goroutines
func BenchmarkCICS(numChan int, numGoroutine int, numSendGoroutine int, intBuffer int, timeSleep time.Duration) {
	var vecCh []chan int
	for i := 0; i < numChan; i++ {
		newCh := make(chan int, intBuffer)
		vecCh = append(vecCh, newCh)
	}

	var vecMuCh []chan bool
	for i := 0; i < numChan; i++ {
		newMuCh := make(chan bool, 1)
		vecMuCh = append(vecMuCh, newMuCh)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < numChan; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < numSendGoroutine; j++ {
				vecMuCh[i] <- true

				time.Sleep(timeSleep)

				<-vecMuCh[i]

				<-vecCh[i]
			}

			wg.Done()
		}()

		for j := 0; j < numGoroutine - 1; j++ {
			if j < numSendGoroutine {
				go func() {
					vecMuCh[i] <- true

					vecCh[i] <- 1

					<-vecMuCh[i]
				}()
			} else {
				go func() {
					vecMuCh[i] <- true

					<-vecMuCh[i]
				}()
			}
		}
	}

	wg.Wait()
}





//// Part 2.5: main function with parameters
//		Run with -help to see the description of each parameter
//		Example:
//		(1)	If you want to see all bugs are reported, you can run `$GoOracle -numChan=2 -sleep=5000 -timeout=10000 -wait=10000 -print`
//		(2)	If you want to see the overhead, you can use overhead.sh

func main() {
	ptrNumChan := flag.Int("numChan", 2, "Number of channels created and monitored in benchmark")
	ptrNumGo := flag.Int("numGo", 10, "Number of child goroutines created for each channel")
	ptrNumMonitorGo := flag.Int("numSyncGo", 5, "Number of child goroutines that we monitor created for each channel")
	ptrMSSleep := flag.Int("sleep", 5000, "Time that monitored child goroutine sleeps before send")
	ptrMSTimeOut := flag.Int("timeout", 10, "Time that parent goroutine waits for child goroutines in each loop." +
		" Bugs are more likely to be triggered if timeout is larger than sleep")
	ptrMSWait := flag.Int("wait", 10000, "Time that parent goroutine waits after the lifetime of channels. " +
		"This is necessary if you want to see all the bug reports, because the parent goroutine may choose the timeout case for many times and exit. " +
		"Then the OS will kill all child goroutines, leaving some bugs not found.")
	ptrPrintBug := flag.Bool("print", false, "Whether a bug is printed out. Default value is false." +
		" Don't turn this on if we want to measure overhead")
	ptrOrigin := flag.Bool("origin", false, "Whether to run the original benchmark. Default value is false.")

	flag.Parse()

	numChan := *ptrNumChan
	numGoroutine := *ptrNumGo
	numSyncGoroutine := *ptrNumMonitorGo
	timeSleep := time.Duration(*ptrMSSleep) * time.Millisecond
	timeOut := time.Duration(*ptrMSTimeOut) * time.Millisecond
	timeWait := time.Duration(*ptrMSWait) * time.Millisecond
	PrintBug = *ptrPrintBug

	//var get_g func() interface{}
	//err := forceexport.GetFunc(&get_g, "runtime.getg")
	//if err != nil {
	//	fmt.Println("Error in forceexport:", err)
	//}
	//g := get_g()
	//gid := reflect.ValueOf(g).FieldByName("goid")
	//fmt.Println("ID:",gid.Int())

	for i := 0; i < 5; i++ {
		InitLogger(i - 1)
		if *ptrOrigin {
			BenchGL(numChan, numGoroutine, numSyncGoroutine, timeSleep, timeOut)
		} else {
			BenchGLDetect(numChan, numGoroutine, numSyncGoroutine, timeSleep, timeOut)
		}
		PrintLogger()
	}


	time.Sleep(timeWait)
}
