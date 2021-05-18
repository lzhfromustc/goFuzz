package fuzzer

import (
	"bufio"
	"fmt"
	"goFuzz/goFuzz/config"
	"os"
	"strconv"
	"strings"
)

type Record struct {
	MapTupleRecord map[int]int
	MapChanRecord  map[string]ChanRecord
}


type ChanRecord struct {
	ChID string
	Closed bool
	NotClosed bool
	CapBuf int
	PeakBuf int
}

const (
	RecordFileName = "myrecord.txt"
	RecordSplitter = "-----"
)

func FileNameOfRecord() string {
	return config.StrTestPath + "/" + RecordFileName
}

func EmptyRecord() *Record {
	return &Record{}
}

func ParseRecordFile() (retRecord *Record) {
	retRecord = EmptyRecord()
	// The input being parsed shouldn't be empty
	file, err := os.Open(FileNameOfRecord())
	if err != nil {
		fmt.Println("Failed to open record file:", FileNameOfRecord())
		return
	}
	defer file.Close()

	var text []string

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	if len(text) == 0 {
		fmt.Println("Record is empty:", FileNameOfInput())
		return
	}

	newRecord := &Record{
		MapTupleRecord: make(map[int]int),
		MapChanRecord:  make(map[string]ChanRecord),
	}

	indexSplitter := -1
	for i, eachline := range text {
		if eachline == RecordSplitter {
			indexSplitter = i
			break
		}
		if eachline == "" {
			continue
		}
		vecStr := strings.Split(eachline, ":")
		if len(vecStr) != 2 {
			fmt.Println("One line of tuple in record has incorrect format:", eachline, "\tLine:", i)
			return
		}
		var tuple, count int
		if tuple, err = strconv.Atoi(vecStr[0]); err != nil {
			fmt.Println("One line of tuple in record has incorrect format:", vecStr, "\tLine:", i)
			return
		}
		if count, err = strconv.Atoi(vecStr[1]); err != nil {
			fmt.Println("One line of tuple in record has incorrect format:", vecStr, "\tLine:", i)
			return
		}
		newRecord.MapTupleRecord[tuple] = count
	}

	if indexSplitter == -1 {
		fmt.Println("Doesn't find RecordSplitter in record. Full text:", text)
		return
	}

	for i := indexSplitter + 1; i < len(text); i++ {
		eachline := text[i]
		if eachline == "" {
			continue
		}
		//chIDString:closedBit:notClosedBit:capBuf:peakBuf
		vecStr := strings.Split(eachline, ":")
		if len(vecStr) != 5 {
			fmt.Println("One line of channel in record has incorrect format:", eachline, "\tLine:", i)
			return
		}
		chRecord := ChanRecord{}
		chRecord.ChID = vecStr[0]
		if vecStr[1] == "0" {
			chRecord.Closed = false
		} else {
			chRecord.Closed = true
		}
		if vecStr[2] == "0" {
			chRecord.NotClosed = false
		} else {
			chRecord.NotClosed = true
		}
		if chRecord.CapBuf, err = strconv.Atoi(vecStr[3]); err != nil {
			fmt.Println("One line of channel in record has incorrect format:", eachline, "\tLine:", i)
			return
		}
		if chRecord.PeakBuf, err = strconv.Atoi(vecStr[4]); err != nil {
			fmt.Println("One line of channel in record has incorrect format:", eachline, "\tLine:", i)
			return
		}
		newRecord.MapChanRecord[chRecord.ChID] = chRecord
	}
	retRecord = newRecord
	return
}

func UpdateMainRecord(mainRecord, curRecord Record) Record {

	// Update a tuple if it doesn't exist or it exists but a better count is observed
	for tuple, count := range curRecord.MapTupleRecord {
		mainCount, exist := mainRecord.MapTupleRecord[tuple]
		if exist {
			if mainCount < count {
				mainRecord.MapTupleRecord[tuple] = count
			}
		} else {
			mainRecord.MapTupleRecord[tuple] = count
		}
	}

	for chID, chRecord := range curRecord.MapChanRecord {
		mainChRecord, exist := mainRecord.MapChanRecord[chID]
		if exist {
			// Update an existing channel's status
			if mainChRecord.Closed == false && chRecord.Closed == true {
				mainChRecord.Closed = true
			}
			if mainChRecord.NotClosed == false && chRecord.NotClosed == true {
				mainChRecord.NotClosed = true
			}
			if mainChRecord.PeakBuf < chRecord.PeakBuf {
				mainChRecord.PeakBuf = chRecord.PeakBuf
			}
			mainChRecord.CapBuf = chRecord.CapBuf

			mainRecord.MapChanRecord[chID] = mainChRecord
		} else {
			// Update a new chan if it doesn't exist
			mainRecord.MapChanRecord[chID] = chRecord
		}
	}

	return mainRecord
}