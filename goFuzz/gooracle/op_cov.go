package gooracle

import (
	"bytes"
	"os"
	"strconv"
	"sync"
)

// opRecords is a list of touched operation ID
var opRecords []uint16
var opMutex sync.Mutex

func recordOp(opID uint16) {
	opMutex.Lock()
	opRecords = append(opRecords, opID)
	opMutex.Unlock()
}


func dumpOpRecordsToFile(filepath string, opRecords []uint16) error {
	outputF, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer outputF.Close()

	var buffer bytes.Buffer
	for _, id := range opRecords {
		buffer.WriteString(strconv.Itoa(int(id)))
		buffer.WriteByte('\n')
	}

	if _, err = outputF.Write(buffer.Bytes()); err != nil {
		return err
	}

	return nil
}


