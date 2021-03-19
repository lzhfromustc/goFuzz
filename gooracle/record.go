package gooracle

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
)

const MapOpLength int = 65536 // 2^16
const MapChLength int = 65536 // 2^16


func XOR(a, b [2]byte) [2]byte {
	a[0] ^= b[0]
	a[1] ^= b[1]
	return a
}

func init() {
	initV1()
}

func initV1() {
	newUint32 := uint32(0)
	PrevLocV1 = &newUint32

	MapOpCountV1 = [MapOpLength]*uint32{}
	for i := 0; i < MapOpLength; i++ {
		newUint32 := uint32(0)
		MapOpCountV1[i] = &newUint32
	}

	MapCh2CloseV1 = [MapChLength]*uint32{}
	for i := 0; i < MapChLength; i++ {
		newUint32 := uint32(0)
		MapCh2CloseV1[i] = &newUint32
	}

	MapCh2PeakBufV1 = [MapChLength]*uint32{}
	for i := 0; i < MapChLength; i++ {
		newUint32 := uint32(0)
		MapCh2PeakBufV1[i] = &newUint32
	}
}

// Version 1:

var MapOpCountV1 [MapOpLength]*uint32
var PrevLocV1 *uint32
var MapCh2CloseV1 [MapChLength]*uint32
var MapCh2PeakBufV1 [MapChLength]*uint32

func RecordOpV1(curLoc [2]byte) {
	// Compute a [2]byte prevLoc from global PrevLocV1
	prevLocSlice := make([]byte, 2)
	prevLocU16 := uint16(atomic.LoadUint32(PrevLocV1))
	binary.BigEndian.PutUint16(prevLocSlice, prevLocU16)
	prevLoc := [2]byte{prevLocSlice[0], prevLocSlice[1]}

	// Update the map
	xor := XOR(curLoc, prevLoc)
	ptrCount := MapOpCountV1[binary.LittleEndian.Uint16(xor[:])]
	atomic.AddUint32(ptrCount, uint32(1))

	// Store to global PrevLocV1
	prevLocSlice[0] = curLoc[0] >> 1
	prevLocSlice[1] = curLoc[1] >> 1
	prevLocU32 := uint32(binary.LittleEndian.Uint16(prevLocSlice))
	atomic.StoreUint32(PrevLocV1, prevLocU32)
}

func RecordCloseV1(chLoc [2]byte) {
	ptrClosed := MapCh2CloseV1[binary.LittleEndian.Uint16(chLoc[:])]
	atomic.StoreUint32(ptrClosed, uint32(1))
}

func RecordSendV1(chLoc [2]byte) {
	ptrPeakBuf := MapCh2PeakBufV1[binary.LittleEndian.Uint16(chLoc[:])]
	atomic.AddUint32(ptrPeakBuf, uint32(1))
}

func DumpV1() {
	f, err := os.Open("/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/mylog.txt")
	if err != nil {
		fmt.Println("error when print file")
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for i, _ := range MapCh2CloseV1 {
		b := MapCh2PeakBufV1[i]
		c := MapCh2CloseV1[i]
		w.WriteString(fmt.Sprintf("Ch:%d\tbuffer:%d\tclosed:%d", i, *b, *c))
	}
	for i, count := range MapOpCountV1 {
		w.WriteString(fmt.Sprintf("Op tuple:%d\tcount:%d", i, count))
	}

	w.Flush()
}