package runtime

import "sync/atomic"

const MaxRecordElem int = 65536 // 2^16

// Settings:
var BoolRecord bool = true
var BoolRecordPerCh bool = true

// TODO: important: extend similar algorithm for mutex, conditional variable, waitgroup, etc

// For the record of each channel:

type ChanRecord struct {
	ChID uint32
	Closed bool
	NotClosed bool
	CapBuf uint16
	PeakBuf uint16
	Ch *hchan
}

var ChRecord [MaxRecordElem]*ChanRecord

// For the record of each channel operation

var GlobalLastLoc uint32
var TupleRecord [MaxRecordElem]uint32

// When a channel is made, create new id, new ChanRecord
func RecordChMake(capBuf int, c *hchan) {
	if BoolRecord == false {
		return
	}

	const size = 64 << 10 // TODO: is 64<<10 too big?
	buf := make([]byte, size)
	buf = buf[:Stack(buf, false)] // TODO: important: is Stack() too heavy? How about replace it will Caller(N)?
	strStack := string(buf)
	stackSingleGo := ParseStackStr(strStack)
	if len(stackSingleGo.VecFuncLine) < 2 {
		return
	}
	strChID := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
	c.id, _ = hashStr(strChID)

	newChRecord := &ChanRecord{
		ChID:      c.id,
		Closed:    false,
		NotClosed: true,
		CapBuf:    uint16(capBuf),
		PeakBuf:   0,
		Ch:		   c,
	}

	ChRecord[c.id] = newChRecord
	c.chanRecord = newChRecord
}

// When a channel operation is executed, update ChanRecord, and update the tuple counter (curLoc XOR prevLoc)
func RecordChOp(c *hchan) {

	// Update ChanRecord
	//print("qcount:",c.qcount, "dataqsiz", c.dataqsiz, "elemsize", c.elemsize, "\n")
	if c.chanRecord.PeakBuf < uint16(c.qcount) { // TODO: only execute this when it is a send operation
		c.chanRecord.PeakBuf = uint16(c.qcount)
		//print("ch:", c.chanRecord.ChID, "\tpeakBuf:", c.chanRecord.PeakBuf, "\n")
	}
	c.chanRecord.Closed = c.closed == 1 // TODO: only execute this when it is a close operation
	if c.chanRecord.Closed {
		c.chanRecord.NotClosed = false
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
	uint32curLoc, _ := hashStr(strOpID) // TODO: important: how to have a uint16 hash of string
	curLoc := uint16(uint32curLoc)
	var preLoc, xorLoc uint16
	if BoolRecordPerCh {
		preLoc = c.preLoc
		c.preLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&GlobalLastLoc))
		atomic.StoreUint32(&GlobalLastLoc, uint32(curLoc >> 1))
	}
	xorLoc = XorUint16(curLoc, preLoc)

	uint32Counter := TupleRecord[xorLoc]
	atomic.AddUint32(&uint32Counter, 1)

}