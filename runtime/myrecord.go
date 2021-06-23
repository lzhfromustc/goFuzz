package runtime

import "sync/atomic"

const MaxRecordElem int = 65536 // 2^16

// Settings:
var BoolRecord bool = true
var BoolRecordPerCh bool = true
var BoolRecordSDK bool = true

// TODO: important: extend similar algorithm for mutex, conditional variable, waitgroup, etc

// For the record of each channel:

type ChanRecord struct {
	StrCreation string // Example: "/data/ziheng/shared/gotest/stubs/toy/src/toy/main_test.go:34"
	Closed      bool
	NotClosed   bool
	CapBuf      uint16
	PeakBuf     uint16
	Ch          *hchan
}

var ChRecord [MaxRecordElem]*ChanRecord

// For the record of each channel operation

var GlobalLastLoc uint32
var TupleRecord [MaxRecordElem]uint32
var ChCount uint16

// When a channel is made, create new id, new ChanRecord
func RecordChMake(capBuf int, c *hchan) {

	if BoolRecordSDK == false {
		if c.chInfo.BoolMakeInSDK == false {
			return
		}
	}

	c.id = ChCount
	ChCount++

	newChRecord := &ChanRecord{
		StrCreation: c.chInfo.StrDebug,
		Closed:      false,
		NotClosed:   true,
		CapBuf:      uint16(capBuf),
		PeakBuf:     0,
		Ch:          c,
	}

	ChRecord[c.id] = newChRecord
	c.chanRecord = newChRecord
}

// When a channel operation is executed, update ChanRecord, and update the tuple counter (curLoc XOR prevLoc)
func RecordChOp(c *hchan) {

	// As mentioned above, we don't record channels created in runtime
	if c.chanRecord == nil {
		return
	}
	if BoolRecordSDK {
		if c.chInfo.BoolMakeInSDK == false {
			return
		}
	}

	// Update ChanRecord
	//print("qcount:",c.qcount, "dataqsiz", c.dataqsiz, "elemsize", c.elemsize, "\n")
	if c.chanRecord.PeakBuf < uint16(c.qcount) { // TODO: only execute this when it is a send operation
		c.chanRecord.PeakBuf = uint16(c.qcount)
		//print("ch:", c.chanRecord.StrCreation, "\tpeakBuf:", c.chanRecord.PeakBuf, "\n")
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

func CurrentGoAddMutex(ch interface{}) {
	lock(&MuMapChToChanInfo)
	chInfo, exist := MapChToChanInfo[ch]
	unlock(&MuMapChToChanInfo)
	if !exist {
		return
	}
	AddRefGoroutine(chInfo, CurrentGoInfo())
}


func CurrentGoAddCond(ch interface{}) {
	lock(&MuMapChToChanInfo)
	chInfo, exist := MapChToChanInfo[ch]
	unlock(&MuMapChToChanInfo)
	if !exist {
		return
	}
	AddRefGoroutine(chInfo, CurrentGoInfo())
}


func CurrentGoAddWaitgroup(ch interface{}) {
	lock(&MuMapChToChanInfo)
	chInfo, exist := MapChToChanInfo[ch]
	unlock(&MuMapChToChanInfo)
	if !exist {
		return
	}
	AddRefGoroutine(chInfo, CurrentGoInfo())
}