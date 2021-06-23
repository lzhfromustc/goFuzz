package fuzzer

import (
	"container/list"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// fuzzerContext is a global context during whole fuzzing process.
var fuzzerContext *FuzzContext = NewFuzzContext()

type FuzzStage string

const (
	// InitStage simply run the empty without any mutation
	InitStage FuzzStage = "init"

	// DeterStage is to create input by tweak select choice one by one
	DeterStage FuzzStage = "deter"

	// CalibStage choose an input from queue to run (prepare for rand)
	CalibStage FuzzStage = "calib"

	// RandStage randomly mutate select choice
	RandStage FuzzStage = "rand"
)

// FuzzContext record all necessary information for help fuzzer to prioritize input and record process.
type FuzzContext struct {
	runTaskCh        chan *RunTask // task for worker to run
	fuzzingQueue     *list.List
	fqLock           sync.RWMutex // lock for fuzzingQueue
	mainRecord       *Record
	allRecordHashMap map[string]struct{}

	// A map from bug ID to stdout file contains that bug
	allBugID2Fp  map[string]string
	bugID2FpLock sync.RWMutex

	// Metrics
	numOfBugsFound      uint64
	numOfRuns           uint64
	numOfFuzzQueryEntry uint64
	startAt             time.Time
}

// NewFuzzContext returns a new FuzzerContext
func NewFuzzContext() *FuzzContext {
	return &FuzzContext{
		runTaskCh:        make(chan *RunTask),
		fuzzingQueue:     list.New(),
		mainRecord:       EmptyRecord(),
		allRecordHashMap: make(map[string]struct{}),
		allBugID2Fp:      make(map[string]string),
		startAt:          time.Now(),
	}
}

func (c *FuzzContext) DequeueQueryEntry() (*FuzzQueryEntry, error) {
	c.fqLock.RLock()
	if c.fuzzingQueue.Len() == 0 {
		c.fqLock.RUnlock()
		return nil, nil
	}
	elm := c.fuzzingQueue.Front()
	c.fqLock.RUnlock()

	c.fqLock.Lock()
	entry := c.fuzzingQueue.Remove(elm)
	c.fqLock.Unlock()
	return entry.(*FuzzQueryEntry), nil

}
func (c *FuzzContext) EnqueueQueryEntry(e *FuzzQueryEntry) error {
	c.fqLock.Lock()
	c.fuzzingQueue.PushBack(e)
	c.fqLock.Unlock()
	log.Printf("enqueued entry: %+v", *e)
	return nil
}

func (c *FuzzContext) IncNumOfRun() {
	atomic.AddUint64(&c.numOfRuns, 1)
}

func (c *FuzzContext) IncNumOfBugsFound(num uint64) {
	atomic.AddUint64(&c.numOfBugsFound, num)
}

func (c *FuzzContext) NewFuzzQueryEntryIndex() uint64 {
	return atomic.AddUint64(&c.numOfFuzzQueryEntry, 1)
}

func (c *FuzzContext) HasBugID(id string) bool {
	c.bugID2FpLock.RLock()
	_, exists := c.allBugID2Fp[id]
	c.bugID2FpLock.RUnlock()
	return exists
}

func (c *FuzzContext) AddBugID(bugID string, filepath string) {
	c.bugID2FpLock.Lock()
	c.allBugID2Fp[bugID] = filepath
	c.bugID2FpLock.Unlock()

}
