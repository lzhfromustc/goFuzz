package gooracle

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"runtime"
	"strconv"
)

func FileNameOfOutput() string {
	return os.Getenv("OutputFullPath")
}

var OutputFile *os.File

func OpenOutputFile() {
	out, err := os.OpenFile(FileNameOfOutput(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open file:", FileNameOfOutput())
		return
	}
	OutputFile = out
	os.Stdout = out
}

func CloseOutputFile() {
	OutputFile.Close()
}

func FileNameOfErr() string {
	return StrTestpath + "/" + ErrFileName
}

var ErrFile *os.File

func CreateErrFile() {
	out, err := os.Create(FileNameOfErr())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfErr())
		return
	}
	ErrFile = out
	os.Stderr = out
}

func CloseErrFile() {
	ErrFile.Close()
}

func FileNameOfRecord() string {
	return StrTestpath + "/" + RecordFileName
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

	str := ""
	// tuple record
	for xorLoc, count := range runtime.TupleRecord {
		if count == 0 {
			continue // no need to record tuple that doesn't show up at all
		}
		h := fnv.New32a()
		h.Write([]byte(strconv.Itoa(xorLoc)))
		str += strconv.Itoa(xorLoc) + ":" + strconv.Itoa(int(count)) + "\n"
	}
	str += RecordSplitter + "\n"

	// channel record
	for _, chRecord := range runtime.ChRecord {
		if chRecord == nil {
			continue
		}
		//chIDString:closedBit:notClosedBit:capBuf:peakBuf
		strChID := chRecord.StrCreation
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
