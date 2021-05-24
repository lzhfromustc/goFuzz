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
	if len(stackSingleGo.VecFuncLine) < 2 { // if the channel is created in runtime, strStack won't contain enough
		// information about where the channel is created. We won't record anything for such channels.
		return
	}
	strChID := stackSingleGo.VecFuncFile[1] + ":" + stackSingleGo.VecFuncLine[1]
	c.id, _ = hashStr(strChID)  // TODO: important: how to have a uint16 hash of string
	c.id = uint32(uint16(c.id))

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

	if c.chanRecord == nil { // As mentioned above, we don't record channels created in runtime
		return
	}

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

	curLoc := getg().uint16ChOpID
	var preLoc, xorLoc uint16
	if BoolRecordPerCh {
		preLoc = c.preLoc
		c.preLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&GlobalLastLoc))
		atomic.StoreUint32(&GlobalLastLoc, uint32(curLoc >> 1))
	}
	xorLoc = XorUint16(curLoc, preLoc)

	atomic.AddUint32(&TupleRecord[xorLoc], 1)
}

func StoreChOpInfo(strOpType string, uint16OpID uint16) {
	getg().strChOpType = strOpType
	getg().uint16ChOpID = uint16OpID
}