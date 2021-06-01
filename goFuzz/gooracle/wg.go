package gooracle

import (
	"sync"
)

// WgInfo is 1-to-1 with every WaitGroup.
type WgInfo struct {
	WgCounter       uint32
	MapRefGoroutine sync.Map
}

func AddRefGoroutineAndWg(w *WgInfo, goInfo *GoInfo) {
	w.MapRefGoroutine.Store(goInfo, true)
	goInfo.mapWgInfo.Store(w, true)
}

func RemoveRefGoroutineAndWg(w *WgInfo, goInfo *GoInfo) {
	w.MapRefGoroutine.Delete(goInfo)
	goInfo.mapChanInfo.Delete(w)
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
	w.MapRefGoroutine.Range(func(key, value interface{}) bool {
		goInfo, _ := key.(*GoInfo)
		isBlock, _ := goInfo.IsBlock()

		if isBlock {
			numOfBlockedGoroutines += 1
		}
		return true
	})

	if numOfBlockedGoroutines == 0 {
		w.IamBug()
	}

}
