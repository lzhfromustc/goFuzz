package gooracle

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

//Note: different from channel, the instrumentation of traditional primitives may fail.
// Because channel must be created by make(chan Type), but traditional primitives can be created by just declare the variable, which may happen when declaring a global variable

var BoolRecordTrad bool = os.Getenv("GF_SCORE_TRAD") == "1"

func RecordLockCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		if runtime.BoolPrintDebugInfo {
			println("For the Lock operation, we don't record its operation")

		}
		return
	} else {
		if runtime.BoolPrintDebugInfo {
			println("For the Lock operation, we recorded its operation")
		}
	}
	switch concrete := ident.(type) {
	case *sync.Mutex:
		RecordMutexOp(concrete, opID)
	case **sync.Mutex:
		RecordMutexOp(*concrete, opID)
	case *sync.RWMutex:
		RecordRWMutexOp(concrete, opID)
	case **sync.RWMutex:
		RecordRWMutexOp(*concrete, opID)
	}
}

func RecordUnlockCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		return
	}
	switch concrete := ident.(type) {
	case *sync.Mutex:
		RecordMutexOp(concrete, opID)
	case **sync.Mutex:
		RecordMutexOp(*concrete, opID)
	case *sync.RWMutex:
		RecordRWMutexOp(concrete, opID)
	case **sync.RWMutex:
		RecordRWMutexOp(*concrete, opID)
	}
}

func RecordRWMutexUniqueCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		return
	}
	switch concrete := ident.(type) {
	case *sync.RWMutex:
		RecordRWMutexOp(concrete, opID)
	case **sync.RWMutex:
		RecordRWMutexOp(*concrete, opID)
	}
}


func RecordWaitCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		return
	}
	switch concrete := ident.(type) {
	case *sync.WaitGroup:
		RecordWgOp(concrete, opID)
	case **sync.WaitGroup:
		RecordWgOp(*concrete, opID)
	case *sync.Cond:
		RecordCondOp(concrete, opID)
	case **sync.Cond:
		RecordCondOp(*concrete, opID)
	}
}

func RecordWgUniqueCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		return
	}
	switch concrete := ident.(type) {
	case *sync.WaitGroup:
		RecordWgOp(concrete, opID)
	case **sync.WaitGroup:
		RecordWgOp(*concrete, opID)
	}
}


func RecordCondUniqueCall(ident interface{}, opID uint16) {
	if !BoolRecordTrad {
		return
	}
	switch concrete := ident.(type) {
	case *sync.Cond:
		RecordCondOp(concrete, opID)
	case **sync.Cond:
		RecordCondOp(*concrete, opID)
	}
}


// RecordWgCreate expected to be instrumented at
// WaitGroup creation.
func RecordWgCreate(wg *sync.WaitGroup, id string, opId uint16) {
	wg.Record = &sync.WgRecord{
		ID: id,
		PreLoc: opId,
		Wg: wg,
	}
	recordOp(opId)
}

// RecordWgOp expected to be instrumented at
// useing of WaitGroup.
func RecordWgOp(wg *sync.WaitGroup, opId uint16) {
	if wg.Record == nil { // This waitgroup's creation is not instrumented
		return
	}
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
	recordOp(opId)
}

// RecordCondCreate expected to be instrumented at
// Condition Variable creation.
func RecordCondCreate(cond *sync.Cond, id string, opId uint16) {
	cond.Record = &sync.CondRecord{
		Cond: cond,
		ID: id,
		PreLoc:opId,
	}
	recordOp(opId)
}

// RecordCondOp expected to be instrumented at
// using of Condition Variable.
func RecordCondOp(cond *sync.Cond, opId uint16) {
	if cond.Record == nil { // This cond's creation is not instrumented
		return
	}
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
	recordOp(opId)
}


func RecordMutexCreate(m *sync.Mutex, id string, opId uint16){
	m.Record = &sync.MutexRecord{
		M: m,
		ID: id,
		PreLoc:opId,
	}
	recordOp(opId)
}


func RecordMutexOp(m *sync.Mutex, opId uint16){
	if m.Record == nil { // This mutex's creation is not instrumented
		return
	}
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
	recordOp(opId)
}

func RecordRWMutexCreate(m *sync.RWMutex, id string, opId uint16){
	m.Record = &sync.RWMutexRecord{
		M: m,
		ID: id,
		PreLoc:opId,
	}
	recordOp(opId)
}


func RecordRWMutexOp(m *sync.RWMutex, opId uint16){
	if m.Record == nil { // This rwmutex's creation is not instrumented
		return
	}
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
	recordOp(opId)
}

