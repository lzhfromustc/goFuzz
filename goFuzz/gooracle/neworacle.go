package gooracle

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var MapGoID2Bytes map[int64]*[]byte

var MuMap sync.Mutex

func init() {
	x := uint32(0)
	CountS = &x
	MuMap = sync.Mutex{}
	MapGoID2Bytes = make(map[int64]*[]byte)
}

func BeforeBlock() {
	gid := runtime.GoID()
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	MuMap.Lock()
	MapGoID2Bytes[gid] = &buf
	MuMap.Unlock()
}

func AfterBlock() {
	gid := runtime.GoID()
	MuMap.Lock()
	delete(MapGoID2Bytes, gid)
	MuMap.Unlock()
}

func DumpBlockingInfo() {
	time.Sleep(8 * time.Second)
	MuMap.Lock()
	fmt.Println("\tPrinting blocking goroutines:")
	for gid, ptrBytes := range MapGoID2Bytes {
		fmt.Println("Blocking goroutine:", gid)
		fmt.Printf("%s\n",*ptrBytes)
	}
	MuMap.Unlock()
}

var CountS *uint32