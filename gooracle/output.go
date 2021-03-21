package gooracle

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func FileNameOfRecord() string {
	return StrTestpath + "/" + RecordFileName
}

func CreateRecordFile() {
	out, err := os.Create(FileNameOfRecord())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfInput())
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	str := ""
	// tuple record
	for xorStr, count := range runtime.StructRecord.MapTupleRecord {
		h := fnv.New32a()
		h.Write([]byte(xorStr))
		tupleUint16 := uint16(h.Sum32())
		str += strconv.Itoa(int(tupleUint16)) + ":" + strconv.Itoa(int(count)) + "\n"
	}
	str += RecordSplitter + "\n"

	// channel record
	for _, chRecord := range runtime.StructRecord.MapChanRecord {
		// chIDString:closedBit:notClosedBit:capBuf:peakBuf
		strChID := chRecord.ChID
		if indexEnter := strings.Index(strChID, "\n"); indexEnter > -1 { // sometimes chID contains "\n"
			strChID = strChID[:indexEnter]
		}
		str += strChID + ":"
		if chRecord.Closed {
			str += "1" + ":"
		} else {
			str += "0" + ":"
		}
		if chRecord.NotClosed {
			str += "1" + ":"
		} else {
			str += "0" + ":"
		}
		str += strconv.Itoa(int(chRecord.CapBuf)) + ":"
		str += strconv.Itoa(int(chRecord.PeakBuf)) + "\n"
	}

	w.WriteString(str)
}
