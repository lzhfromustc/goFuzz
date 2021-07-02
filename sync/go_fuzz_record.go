package sync

//Note: different from channel, the instrumentation of traditional primitives may fail.
// Because channel must be created by make(chan Type), but traditional primitives can be created by just declare the variable, which may happen when declaring a global variable

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

// The usage of structs above is in gooracle/syncRecord.go