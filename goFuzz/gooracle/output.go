package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//func FileNameOfOutput() string {
//	return os.Getenv("OutputFullPath")
//}
//
//var OutputFile *os.File
//
//func OpenOutputFile() {
//	out, err := os.OpenFile(FileNameOfOutput(),
//		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		fmt.Println("Failed to open file:", FileNameOfOutput())
//		return
//	}
//	OutputFile = out
//	os.Stdout = out
//}
//
//func CloseOutputFile() {
//	OutputFile.Close()
//}
//
//func FileNameOfErr() string {
//	return StrTestpath + "/" + ErrFileName
//}
//
//var ErrFile *os.File
//
//func CreateErrFile() {
//	out, err := os.Create(FileNameOfErr())
//	if err != nil {
//		fmt.Println("Failed to create file:", FileNameOfErr())
//		return
//	}
//	ErrFile = out
//	os.Stderr = out
//}
//
//func CloseErrFile() {
//	ErrFile.Close()
//}

func FileNameOfRecord() string {
	return os.Getenv("GF_RECORD_FILE")
}

func CreateRecordFile() {
	out, err := os.Create(FileNameOfRecord())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfRecord())
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	var sb strings.Builder
	// tuple record
	for xorLoc, count := range runtime.TupleRecord {
		if count == 0 {
			continue // no need to record tuple that doesn't show up at all
		}
		// h := fnv.New32a()
		// h.Write([]byte(strconv.Itoa(xorLoc)))
		sb.WriteString(strconv.Itoa(xorLoc))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(int(count)))
		sb.WriteString("\n")
	}
	sb.WriteString(RecordSplitter)
	sb.WriteString("\n")

	// channel record
	for _, chRecord := range runtime.ChRecord {
		if chRecord == nil {
			continue
		}
		//chIDString:closedBit:notClosedBit:capBuf:peakBuf
		strChID := chRecord.StrCreation
		sb.WriteString(strChID)
		sb.WriteString(":")
		if chRecord.Closed {
			sb.WriteString("1:")
		} else {
			sb.WriteString("0:")
		}
		if chRecord.NotClosed {
			sb.WriteString("1:")
		} else {
			sb.WriteString("0:")
		}
		sb.WriteString(strconv.Itoa(int(chRecord.CapBuf)))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(int(chRecord.PeakBuf)))
		sb.WriteString("\n")
	}

	w.WriteString(sb.String())
}

func StrPointer(v interface{}) string {
	return fmt.Sprintf("%p", v)
}
