package runtime

type Record struct {
	MapTupleRecord map[string]uint16
	MapChanRecord  map[*hchan]*ChanRecord
}


type ChanRecord struct {
	ChID string
	Closed bool
	NotClosed bool
	CapBuf uint16
	PeakBuf uint16
	LastOpID string
	Ch *hchan
}

var StructRecord Record
var MuChRecord mutex
//var MuTupleRecord mutex
var LastOpID string

var BoolRecord bool = true
var BoolRecordPerCh bool = true

func RecordChMake(capBuf int, c *hchan) {
	if BoolRecord == false {
		return
	}
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncLine) < 2 {
		return
	}
	strChID := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
	newChRecord := &ChanRecord{
		ChID:      strChID,
		Closed:    false,
		NotClosed: true,
		CapBuf:    uint16(capBuf),
		PeakBuf:   0,
		Ch:		   c,
	}
	//print("Record hchan:", c, "\tID:", strChID, "\n")
	lock(&MuChRecord)
	StructRecord.MapChanRecord[c] = newChRecord
	unlock(&MuChRecord)
}

func RecordChOp(c *hchan) {
	if BoolRecord == false {
		return
	}
	lock(&MuChRecord)

	chRecord, exist := StructRecord.MapChanRecord[c]
	if !exist {
		unlock(&MuChRecord)
		//print("Warning: chRecord not exist when a send op is executed\n")
		//const size = 64 << 10
		//buf := make([]byte, size)
		//buf = buf[:Stack(buf, false)]
		//print(string(buf))
		return
	}
	//print("qcount:",c.qcount, "dataqsiz", c.dataqsiz, "elemsize", c.elemsize, "\n")
	if chRecord.PeakBuf < uint16(c.qcount) {
		chRecord.PeakBuf = uint16(c.qcount)
		//print("ch:", chRecord.ChID, "\tpeakBuf:", chRecord.PeakBuf, "\n")
	}
	chRecord.Closed = c.closed == 1
	if chRecord.Closed {
		chRecord.NotClosed = false
	}



	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)]
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncLine) < 2 {
		return
	}
	strOpID := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
	var xorOpID string
	if BoolRecordPerCh {
		xorOpID = chRecord.LastOpID + "|" + strOpID
		chRecord.LastOpID = strOpID
	} else {
		xorOpID = LastOpID + "|" + strOpID
		LastOpID = strOpID
	}

	//print("xorOpID:", xorOpID, "\n")

	StructRecord.MapTupleRecord[xorOpID] += 1

	unlock(&MuChRecord)
}