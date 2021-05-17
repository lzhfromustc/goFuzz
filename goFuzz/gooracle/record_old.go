package gooracle

//
//import (
//	"bufio"
//	"encoding/binary"
//	"os"
//	"strconv"
//)
//
//func init() {
//	MapOpCountV1 = make(map[uint16]uint)
//	MapCh2PrevLoc = make(map[interface{}]*[2]byte)
//	MapCh2Info = make(map[interface{}]*ChFuzzInfo)
//}
//
//func XOR(a, b [2]byte) [2]byte {
//	a[0] ^= b[0]
//	a[1] ^= b[1]
//	return a
//}
//
//
//func init() {
//	initV1()
//}
//
//func initV1() {
//	newUint32 := uint32(0)
//	PrevLocV1 = &newUint32
//
//	MapOpCountV1 = [MapOpLength]*uint32{}
//	for i := 0; i < MapOpLength; i++ {
//		newUint32 := uint32(0)
//		MapOpCountV1[i] = &newUint32
//	}
//
//	MapCh2CloseV1 = [MapChLength]*uint32{}
//	for i := 0; i < MapChLength; i++ {
//		newUint32 := uint32(0)
//		MapCh2CloseV1[i] = &newUint32
//	}
//
//	MapCh2PeakBufV1 = [MapChLength]*uint32{}
//	for i := 0; i < MapChLength; i++ {
//		newUint32 := uint32(0)
//		MapCh2PeakBufV1[i] = &newUint32
//	}
//}
//
//// Version 1:
//
//var MapOpCountV1 [MapOpLength]*uint32
//var PrevLocV1 *uint32
//var MapCh2CloseV1 [MapChLength]*uint32
//var MapCh2PeakBufV1 [MapChLength]*uint32
//
//func RecordOpV1(curLoc [2]byte) {
//	// Compute a [2]byte prevLoc from global PrevLocV1
//	prevLocSlice := make([]byte, 2)
//	prevLocU16 := uint16(atomic.LoadUint32(PrevLocV1))
//	binary.BigEndian.PutUint16(prevLocSlice, prevLocU16)
//	prevLoc := [2]byte{prevLocSlice[0], prevLocSlice[1]}
//
//	// Update the map
//	xor := XOR(curLoc, prevLoc)
//	ptrCount := MapOpCountV1[binary.LittleEndian.Uint16(xor[:])]
//	atomic.AddUint32(ptrCount, uint32(1))
//
//	// Store to global PrevLocV1
//	prevLocSlice[0] = curLoc[0] >> 1
//	prevLocSlice[1] = curLoc[1] >> 1
//	prevLocU32 := uint32(binary.LittleEndian.Uint16(prevLocSlice))
//	atomic.StoreUint32(PrevLocV1, prevLocU32)
//}
//
//func RecordCloseV1(chLoc [2]byte) {
//	ptrClosed := MapCh2CloseV1[binary.LittleEndian.Uint16(chLoc[:])]
//	atomic.StoreUint32(ptrClosed, uint32(1))
//}
//
//func RecordSendV1(chLoc [2]byte) {
//	ptrPeakBuf := MapCh2PeakBufV1[binary.LittleEndian.Uint16(chLoc[:])]
//	atomic.AddUint32(ptrPeakBuf, uint32(1))
//}
//
//func DumpV1() {
//	f, err := os.Open("/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/mylog.txt")
//	if err != nil {
//		fmt.Println("error when print file")
//	}
//	defer f.Close()
//
//	w := bufio.NewWriter(f)
//	for i, _ := range MapCh2CloseV1 {
//		b := MapCh2PeakBufV1[i]
//		c := MapCh2CloseV1[i]
//		w.WriteString(fmt.Sprintf("Ch:%d\tbuffer:%d\tclosed:%d", i, *b, *c))
//	}
//	for i, count := range MapOpCountV1 {
//		w.WriteString(fmt.Sprintf("Op tuple:%d\tcount:%d", i, count))
//	}
//
//	w.Flush()
//}
//func DumpMapCoverageV1(writer *bufio.Writer) {
//	for key, count := range MapOpCountV1 {
//		writer.WriteString(strconv.Itoa(int(key)))
//		writer.WriteString(strconv.Itoa(int(count)))
//		writer.WriteString("\n")
//	}
//}
//
//func DumpMapCoverageV2(writer *bufio.Writer) {
//	for key, count := range MapOpCountV1 {
//		writer.WriteString(strconv.Itoa(int(key)))
//		writer.WriteString(strconv.Itoa(int(count)))
//		writer.WriteString("\n")
//	}
//}
//
//func DumpMapCoverageV3(writer *bufio.Writer) {
//	for ch, chFuzzInfo := range MapCh2Info {
//		for _, b := range chFuzzInfo.ID {
//			writer.WriteByte(b)
//		}
//		writer.WriteString("\t")
//		writer.WriteString(strconv.Itoa(int(chFuzzInfo.PeakBuf)))
//		writer.WriteString("\t")
//		if chFuzzInfo.IsClosed {
//			writer.WriteString("1")
//		} else {
//			writer.WriteString("0")
//		}
//		writer.WriteString("\t")
//		for key, count := range chFuzzInfo.MapOpCountV1 {
//			writer.WriteString(strconv.Itoa(int(key)))
//			writer.WriteString(strconv.Itoa(int(count)))
//		}
//		_ = ch
//	}
//}
//
//var MapOpCountV1 map[uint16]uint
//
//var PrevLocSlice [2]byte
//
//
//func RecordV1(curLoc [2]byte, ch interface{}, op uint8) {
//	xor := XOR(curLoc, PrevLocSlice)
//	MapOpCountV1[binary.LittleEndian.Uint16(xor[:])]++
//	PrevLocSlice[0] = curLoc[0] >> 1
//	PrevLocSlice[1] = curLoc[1] >> 1
//}
//
//var MapCh2PrevLoc map[interface{}]*[2]byte
//
//func SetMap(Id int, m map[uint16]uint){
//
//}
//
//func MergeMap(vecMapCoverage []map[uint16]uint) {
//
//}
//
//func RecordV2(curLoc [2]byte, ch interface{}, op uint8) {
//	prevLoc, ok := MapCh2PrevLoc[ch]
//	if !ok {
//		prevLoc = &[2]byte{0, 0}
//	}
//	xor := XOR(curLoc, *prevLoc)
//	MapOpCountV1[binary.LittleEndian.Uint16(xor[:])]++
//	(*prevLoc)[0] = curLoc[0] >> 1
//	(*prevLoc)[1] = curLoc[1] >> 1
//}
//
//type ChFuzzInfo struct {
//	ID [2]byte
//	Buf uint
//	PeakBuf uint
//	IsClosed bool
//	PrevLocSlice *[2]byte
//	MapOpCountV1 map[uint16]uint
//}
//
//var MapCh2Info map[interface{}]*ChFuzzInfo
//
//const (
//	ChSend uint8 = 0
//	ChRecv uint8 = 1
//	ChClose uint8 = 2
//	ChNotSelected uint8 = 3
//)
//
//func RecordV3(curLoc [2]byte, ch interface{}, op uint8) {
//	info := MapCh2Info[ch]
//
//	if info == nil {
//		MapCh2Info[ch] = &ChFuzzInfo{
//			ID:          [2]byte{0, 0},
//			Buf:         0,
//			PeakBuf:     0,
//			IsClosed:    false,
//			PrevLocSlice:     &[2]byte{0, 0},
//			MapOpCountV1: make(map[uint16]uint),
//		}
//		info = MapCh2Info[ch]
//	}
//
//	if info.PrevLocSlice == nil {
//		info.PrevLocSlice = &[2]byte{0, 0}
//	}
//	prevLoc := info.PrevLocSlice
//	xor := XOR(curLoc, *prevLoc)
//	if info.MapOpCountV1 == nil {
//		info.MapOpCountV1 = make(map[uint16]uint)
//	}
//	info.MapOpCountV1[binary.LittleEndian.Uint16(xor[:])]++
//	(*prevLoc)[0] = curLoc[0] >> 1
//	(*prevLoc)[1] = curLoc[1] >> 1
//
//	switch op {
//	case ChSend, ChRecv:
//		chCast, ok := ch.(chan interface{})
//		if ok {
//			curBuf := uint(len(chCast))
//			if info.PeakBuf < curBuf {
//				info.PeakBuf = curBuf
//			}
//		}
//	case ChClose:
//		info.IsClosed = true
//	}
//}
//
//func ReadNextNInt(in *os.File, n int) ([]int, error) {
//	result := make([]int, n)
//	b, err := ReadNextNByte(in, 4 * n)
//	for i := 0; i < n; i++ {
//		b_each := b[i * 4 : i * 4 + 4]
//		result[i] = int(binary.LittleEndian.Uint32(b_each))
//	}
//
//	return result, err
//}
//
//func ReadNextInt64(in *os.File) (int64, error) {
//	n := 1
//	result := make([]int64, n)
//	b, err := ReadNextNByte(in, 4 * n)
//	for i := 0; i < n; i++ {
//		b_each := b[i * 4 : i * 4 + 4]
//		result[i] = int64(binary.LittleEndian.Uint32(b_each))
//	}
//
//	return result[0], err
//}
//
//func read_bytes_into_uint32(in *os.File) (uint32, error) {
//	b, err := ReadNextNByte(in, 4)
//	return binary.LittleEndian.Uint32(b), err
//}
//
//func ReadNextNByte(in *os.File, size int) ([]byte, error) {
//	b := make([]byte, size)
//	_, err := in.Read(b)
//	return b, err
//}