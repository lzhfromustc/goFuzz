package runtime

import (
	"sync/atomic"
)

func init() {
	MapChToChanInfo = make(map[interface{}]*ChanInfo)
}

//// Part 1.1: data struct for each channel

// ChanInfo is 1-to-1 with every channel. It tracks a list of goroutines that hold the reference to the channel
type ChanInfo struct {
	Chan            *hchan          // Stores the channel. Can be used as ID of channel
	IntBuffer       int             // The buffer capability of channel. 0 if channel is unbuffered
	MapRefGoroutine map[*GoInfo]struct{} // Stores all goroutines that still hold reference to this channel
	StrDebug        string
	IntFlagFoundBug int32 // Use atomic int32 operations to mark if a bug is reported
}

var MapChToChanInfo map[interface{}]*ChanInfo
var MuMapChToChanInfo mutex
//var DefaultCaseChanInfo = &ChanInfo{}

// Initialize a new ChanInfo with a given channel
func NewChanInfo(ch *hchan) *ChanInfo {
	newChInfo := &ChanInfo{
		Chan:            ch,
		IntBuffer:       int(ch.dataqsiz),
		MapRefGoroutine: make(map[*GoInfo]struct{}),
		StrDebug:        "",
		IntFlagFoundBug: 0,
	}
	AddRefGoroutine(newChInfo, CurrentGoInfo())

	return newChInfo
}

// FindChanInfo can retrieve a initialized ChanInfo for a given channel
func FindChanInfo(ch interface{}) *ChanInfo {
	lock(&MuMapChToChanInfo)
	chInfo := MapChToChanInfo[ch]
	unlock(&MuMapChToChanInfo)
	return chInfo
}

func LinkChToLastChanInfo(ch interface{}) {
	lock(&MuMapChToChanInfo)
	MapChToChanInfo[ch] = LoadLastChanInfo()
	unlock(&MuMapChToChanInfo)
}

// After the creation of a new channel, or at the head of a goroutine that holds a reference to a channel,
// or whenever a goroutine obtains a reference to a channel, call this function
// AddRefGoroutine links a channel with a goroutine, meaning the goroutine holds the reference to the channel
func AddRefGoroutine(chInfo *ChanInfo, goInfo *GoInfo) {
	chInfo.AddGoroutine(goInfo)
	goInfo.AddChan(chInfo)
}

func RemoveRefGoroutine(chInfo *ChanInfo, goInfo *GoInfo) {
	chInfo.RemoveGoroutine(goInfo)
	goInfo.RemoveChan(chInfo)
}

// This means the goroutine mapped with goInfo holds the reference to chInfo.Chan
func (chInfo *ChanInfo) AddGoroutine(goInfo *GoInfo) {
	chInfo.MapRefGoroutine[goInfo] = struct{}{}
}

func (chInfo *ChanInfo) RemoveGoroutine(goInfo *GoInfo) {
	delete(chInfo.MapRefGoroutine, goInfo)
}

// A blocking bug is detected, if all goroutines that hold the reference to a channel are blocked at an operation of the channel
// Should be called with chInfo.Chan.lock is held
func (chInfo *ChanInfo) CheckBlockBug(CS []*ChanInfo) {
	mapCS := make(map[*ChanInfo]struct{})
	mapGS := make(map[*GoInfo]struct{}) // all goroutines that hold reference to ch0
	if CS == nil || len(CS) == 0 { // called from separate send or receive
		mapCS[chInfo] = struct{}{}
		if atomic.LoadInt32(&chInfo.IntFlagFoundBug) != 0 {
			return
		}
	} else { // called from select
		for _, chI := range CS {
			mapCS[chI] = struct{}{}
			if atomic.LoadInt32(&chI.IntFlagFoundBug) != 0 {
				return
			}
		}
	}


	for chI, _ := range mapCS {
		for goInfo, _ := range chI.MapRefGoroutine {
			mapGS[goInfo] = struct{}{}
		}
	}

	loopGS:
	for goInfo, _ := range mapGS {
		if goInfo.BlockMap == nil { // The goroutine is executing non-blocking operations
			return
		}

		for chI, _ := range goInfo.BlockMap { // if op is select, op.Channels() returns multiple channels
			if _, exist := mapCS[chI]; !exist {
				mapCS[chI] = struct{}{} // update CS
				for goInfo, _ := range chI.MapRefGoroutine { // update GS
					mapGS[goInfo] = struct{}{}
				}
				goto loopGS // since mapGS is updated, we should run this loop once again
			}
		}
	}

	ReportBug(mapCS)
}

//func (chInfo *ChanInfo) CheckBlockBug() {
//	if atomic.LoadInt32(&chInfo.intFlagFoundBug) != 0 {
//		return
//	}
//
//	if chInfo.intBuffer == 0 {
//		countRefGo := 0 // Number of goroutines that hold the reference to the channel
//		countBlockAtThisChanGo := 0 // Number of goroutines that are blocked at an operation of this channel
//		f := func(key interface{}, value interface{}) bool {
//			goInfo, _ := key.(*GoInfo)
//
//			boolIsBlock, _ := goInfo.IsBlockAtGivenChan(chInfo)
//			if boolIsBlock {
//				countBlockAtThisChanGo++
//			}
//			countRefGo++
//			return true // continue Range
//		}
//		chInfo.mapRefGoroutine.Range(f)
//
//		if countRefGo == countBlockAtThisChanGo {
//			if countRefGo == 0 {
//				// debug code
//				countRefGo2 := 0 // Number of goroutines that hold the reference to the channel
//				countBlockAtThisChanGo2 := 0 // Number of goroutines that are blocked at an operation of this channel
//				f := func(key interface{}, value interface{}) bool {
//					goInfo, _ := key.(*GoInfo)
//
//					boolIsBlock, _ := goInfo.IsBlockAtGivenChan(chInfo)
//					if boolIsBlock {
//						countBlockAtThisChanGo2++
//					}
//					countRefGo2++
//					return true // continue Range
//				}
//				chInfo.mapRefGoroutine.Range(f)
//				fmt.Print()
//
//				return
//			}
//			ReportBug(chInfo)
//			atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
//		}
//
//	} else { // Buffered channel
//		if reflect.ValueOf(chInfo.Chan).Len() == chInfo.intBuffer { // Buffer is full
//			// Check if all ref goroutines are blocked at send
//			boolAllBlockAtSend := true
//			countRefGo := 0
//			countBlockAtThisChanGo := 0
//			f := func(key interface{}, value interface{}) bool {
//				goInfo, _ := key.(*GoInfo)
//
//				boolIsBlock, strOp := goInfo.IsBlockAtGivenChan(chInfo)
//				if boolIsBlock {
//					countBlockAtThisChanGo++
//				}
//				if strOp != Send {
//					boolAllBlockAtSend = false
//				}
//				countRefGo++
//				return true // continue Range
//			}
//			chInfo.mapRefGoroutine.Range(f)
//
//			if countRefGo == countBlockAtThisChanGo && boolAllBlockAtSend {
//				ReportBug(chInfo)
//				atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
//			}
//
//		} else if reflect.ValueOf(chInfo.Chan).Len() == 0 { // Buffer is empty
//			// Check if all ref goroutines are blocked at receive
//			boolAllBlockAtRecv := true
//			countRefGo := 0
//			countBlockAtThisChanGo := 0
//			f := func(key interface{}, value interface{}) bool {
//				goInfo, _ := key.(*GoInfo)
//
//				boolIsBlock, strOp := goInfo.IsBlockAtGivenChan(chInfo)
//				if boolIsBlock {
//					countBlockAtThisChanGo++
//				}
//				if strOp != Recv {
//					boolAllBlockAtRecv = false
//				}
//				countRefGo++
//				return true // continue Range
//			}
//			chInfo.mapRefGoroutine.Range(f)
//
//			if countRefGo == countBlockAtThisChanGo && boolAllBlockAtRecv {
//				ReportBug(chInfo)
//				atomic.AddInt32(&chInfo.intFlagFoundBug, 1)
//			}
//
//		} else { // Buffer is not full or empty. Then it is not possible to block
//			// do nothing
//		}
//	}
//}

func ReportBug(mapCS map[*ChanInfo]struct{}) {
	for chInfo, _ := range mapCS {
		atomic.AddInt32(&chInfo.IntFlagFoundBug, 1)
	}
	print("======A blocking bug is found!======\n")
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	print(string(buf), "\n")
}

// Part 1.2 Data structure for waitgroup

// WgInfo is 1-to-1 with every WaitGroup.
type WgInfo struct {
	WgCounter       uint32
	MapRefGoroutine map[*GoInfo]struct{}
}

func NewWgInfo() *WgInfo {
	return &WgInfo{MapRefGoroutine: make(map[*GoInfo]struct{})}
}

func AddRefGoroutineAndWg(w *WgInfo, goInfo *GoInfo) {
	w.MapRefGoroutine[goInfo] = struct{}{}
	goInfo.AddWg(w)
}

func RemoveRefGoroutineAndWg(w *WgInfo, goInfo *GoInfo) {
	delete(w.MapRefGoroutine, goInfo)
	goInfo.RemoveWg(w)
}

func (w *WgInfo) IamBug() {

}

func (w *WgInfo) CheckBlockBug() {
	// Wait will not be blocked if counter is 0
	if w.WgCounter == 0 {
		return
	}

	numOfBlockedGoroutines := 0

	// Counting running Goroutine
	for goInfo, _ := range w.MapRefGoroutine {
		isBlock, _ := goInfo.IsBlock()
		if isBlock {
			numOfBlockedGoroutines += 1
		}
	}

	if numOfBlockedGoroutines == 0 {
		w.IamBug()
	}

}

//// Part 2.1: data struct for each goroutine

// GoInfo is 1-to-1 with each goroutine.
// Go language doesn't allow us to acquire the ID of a goroutine, because they want goroutines to be anonymous.
// Normally, Go programmers use runtime.Stack() to print all IDs of all goroutines, but this function is very inefficient
//, since it calls stopTheWorld()
// Currently we use a global atomic int64 to differentiate each goroutine, and a variable currentGo to represent each goroutine
// This is not a good practice because the goroutine need to pass currentGo to its every callee
type GoInfo struct {
	G *g
	BlockMap map[*ChanInfo]string // Nil when normally running. When blocked at an operation of ChanInfo, store
	// one ChanInfo and the operation. When blocked at select, store multiple ChanInfo and
	// operation. Default in select is also also stored in map, which is DefaultCaseChanInfo
	MapChanInfo map[*ChanInfo]struct{} // Stores all channels that this goroutine still hold reference to
	MapWgInfo map[*WgInfo]struct{}
}

const (
	Send = "Send"
	Recv = "Recv"
	Close = "Close"
	BSelect = "BlockingSelect"
	NBSelect = "NonBlockingSelect"

	MuLock = "MuLock"
	MuUnlock = "MuUnlock"

	WgWait = "WgWait"

	CdWait = "CdWait"
	CdSignal = "CdSignal"
	CdBroadcast = "CdBroadcast"
)

// Initialize a GoInfo
func NewGoInfo(goroutine *g) *GoInfo {
	newGoInfo := &GoInfo{
		G:          goroutine,
		BlockMap:    nil,
		MapChanInfo: make(map[*ChanInfo]struct{}),
	}
	return newGoInfo
}

func CurrentGoInfo() *GoInfo {
	return getg().goInfo
}

func StoreLastChanInfo(chInfo *ChanInfo) {
	getg().lastChanInfo = chInfo
}

func LoadLastChanInfo() *ChanInfo {
	return getg().lastChanInfo
}

// This means the goroutine mapped with goInfo holds the reference to chInfo.Chan
func (goInfo *GoInfo) AddChan(chInfo *ChanInfo) {
	if goInfo.MapChanInfo == nil {
		goInfo.MapChanInfo = make(map[*ChanInfo]struct{})
	}
	goInfo.MapChanInfo[chInfo] = struct{}{}
}

func (goInfo *GoInfo) RemoveChan(chInfo *ChanInfo) {
	if goInfo.MapChanInfo != nil {
		delete(goInfo.MapChanInfo, chInfo)
	}
}

func (goInfo *GoInfo) AddWg(wgInfo *WgInfo) {
	if goInfo.MapWgInfo == nil {
		goInfo.MapWgInfo = make(map[*WgInfo]struct{})
	}
	goInfo.MapWgInfo[wgInfo] = struct{}{}
}

func (goInfo *GoInfo) RemoveWg(wgInfo *WgInfo) {
	if goInfo.MapWgInfo != nil {
		delete(goInfo.MapWgInfo, wgInfo)
	}
}

func CurrentGoAddCh(ch interface{}) {
	lock(&MuMapChToChanInfo)
	chInfo, exist := MapChToChanInfo[ch]
	unlock(&MuMapChToChanInfo)
	if !exist {
		return
	}
	AddRefGoroutine(chInfo, CurrentGoInfo())
}

// RemoveRef should be called at the end of every goroutine. It will remove goInfo from the reference list of every
// channel it holds the reference to
func (goInfo *GoInfo) RemoveAllRef() {

	if goInfo.MapChanInfo == nil {
		return
	}
	for chInfo, _ := range goInfo.MapChanInfo {
		RemoveRefGoroutine(chInfo, goInfo)

		lock(&chInfo.Chan.lock)
		chInfo.CheckBlockBug(nil)
		unlock(&chInfo.Chan.lock)
	}
}

// SetBlockAt should be called before each channel operation, meaning the current goroutine is about to execute that operation
// Note that we check bug in this function, because it's possible for the goroutine to be blocked forever if it execute that operation
// For example, a channel with no buffer is held by a parent and a child.
//              The parent has already exited, but the child is now about to send to that channel.
//              Then now is our only chance to detect this bug, so we call CheckBlockBug()
func (goInfo *GoInfo) SetBlockAt(hch *hchan, strOp string) {
	if goInfo.BlockMap == nil {
		goInfo.BlockMap = make(map[*ChanInfo]string)
	}
	goInfo.BlockMap[hch.chInfo] = strOp
}

// WithdrawBlock should be called after each channel operation, meaning the current goroutine finished execution that operation
// If the operation is select, remember to call this function right after each case of the select
func (goInfo *GoInfo) WithdrawBlock() {
	goInfo.BlockMap = nil
}

func (goInfo *GoInfo) IsBlock() (boolIsBlock bool, strOp string) {
	boolIsBlock, strOp = false, ""
	boolIsSelect := false

	if goInfo.BlockMap == nil {
		return
	} else {
		boolIsBlock = true
	}

	// Now we compute strOp

	if len(goInfo.BlockMap) > 1 {
		boolIsSelect = true
	} else if len(goInfo.BlockMap) == 0 {
		print("Fatal in GoInfo.IsBlock(): goInfo.BlockMap is not nil but lenth is 0\n")
	}

	if boolIsSelect {
		strOp = BSelect
	} else {
		for _, op := range goInfo.BlockMap { // This loop will be executed only one time, since goInfo.BlockMap's len() is 1
			strOp = op
		}
	}

	return
}

// This function checks if the goroutine mapped with goInfo is currently blocking at an operation of chInfo.Chan
// If so, returns true and the string of channel operation
func (goInfo *GoInfo) IsBlockAtGivenChan(chInfo *ChanInfo) (boolIsBlockAtGiven bool, strOp string) {
	boolIsBlockAtGiven, strOp = false, ""

	if goInfo.BlockMap == nil {
		return
	}

	for chanInfo, op := range goInfo.BlockMap {
		if chanInfo == chInfo {
			boolIsBlockAtGiven = true
			strOp = op
			break
		}
	}

	return
}

