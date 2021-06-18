package sync

import (
	"runtime"
	"sync/atomic"
)

type WgRecord struct {
	// ID is the filename + line number (where it creates) and will be determined at instrumentation
	ID     string
	PreLoc uint16
	Wg     *WaitGroup
}

type CondRecord struct {
	ID   string
	PreLoc uint16
	Cond *Cond
}

type MutexRecord struct {
	ID     string
	PreLoc uint16
	M *Mutex
}

type RWMutexRecord struct {
	ID     string
	PreLoc uint16
	M *RWMutex
}

// RecordWgCreate expected to be instrumented at
// WaitGroup creation.
func RecordWgCreate(wg *WaitGroup, id string, opId uint16) {
	wg.Record = WgRecord{
		ID: id,
		PreLoc: opId,
		Wg: wg,
	}
}

// RecordWgOp expected to be instrumented at
// useing of WaitGroup.
func RecordWgOp(wg *WaitGroup, opId uint16) {
	curLoc := opId
	var preLoc, xorLoc uint16

	// TODO: change goBoolRecordPerCh to global to gofuzz instead of runtime
	if runtime.BoolRecordPerCh {
		preLoc = wg.Record.PreLoc
		wg.Record.PreLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&runtime.GlobalLastLoc))
		atomic.StoreUint32(&runtime.GlobalLastLoc, uint32(curLoc>>1))
	}
	xorLoc = runtime.XorUint16(curLoc, preLoc)

	atomic.AddUint32(&runtime.TupleRecord[xorLoc], 1)
}

// RecordCondCreate expected to be instrumented at
// Condition Variable creation.
func RecordCondCreate(cond *Cond, id string, opId uint16) {
	cond.Record = CondRecord{
		Cond: cond,
		ID: id,
		PreLoc:opId,
	}

}

// RecordCondOp expected to be instrumented at
// using of Condition Variable.
func RecordCondOp(cond *Cond, opId uint16) {
	curLoc := opId
	var preLoc, xorLoc uint16

	if runtime.BoolRecordPerCh {
		preLoc = cond.Record.PreLoc
		cond.Record.PreLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&runtime.GlobalLastLoc))
		atomic.StoreUint32(&runtime.GlobalLastLoc, uint32(curLoc>>1))
	}
	xorLoc = runtime.XorUint16(curLoc, preLoc)

	atomic.AddUint32(&runtime.TupleRecord[xorLoc], 1)
}


func RecordMutexCreate(m *Mutex, id string, opId uint16){
	m.Record = MutexRecord{
		M: m,
		ID: id,
		PreLoc:opId,
	}
}


func RecordMutexOp(m *Mutex, opId uint16){
	curLoc := opId
	var preLoc, xorLoc uint16

	if runtime.BoolRecordPerCh {
		preLoc = m.Record.PreLoc
		m.Record.PreLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&runtime.GlobalLastLoc))
		atomic.StoreUint32(&runtime.GlobalLastLoc, uint32(curLoc>>1))
	}
	xorLoc = runtime.XorUint16(curLoc, preLoc)

	atomic.AddUint32(&runtime.TupleRecord[xorLoc], 1)
}

func RecordRWMutexCreate(m *RWMutex, id string, opId uint16){
	m.Record = RWMutexRecord{
		M: m,
		ID: id,
		PreLoc:opId,
	}
}


func RecordRWMutexOp(m *RWMutex, opId uint16){
	curLoc := opId
	var preLoc, xorLoc uint16

	if runtime.BoolRecordPerCh {
		preLoc = m.Record.PreLoc
		m.Record.PreLoc = curLoc >> 1
	} else {
		preLoc = uint16(atomic.LoadUint32(&runtime.GlobalLastLoc))
		atomic.StoreUint32(&runtime.GlobalLastLoc, uint32(curLoc>>1))
	}
	xorLoc = runtime.XorUint16(curLoc, preLoc)

	atomic.AddUint32(&runtime.TupleRecord[xorLoc], 1)
}

