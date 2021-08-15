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
var curQueueEntry *list.Element  = nil

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
	allBugID2Fp  map[string]*BugMetrics
	bugID2FpLock sync.RWMutex

	// Metrics
	numOfBugsFound      uint64
	numOfRuns           uint64
	numOfFuzzQueryEntry uint64
	numOfTargets        uint64
	targetStages        map[string]*TargetMetrics
	targetStagesLock    sync.RWMutex
	startAt             time.Time

	// timeout counter: src => how many times timeout when running this src
	// if more than 3, drop it in the queue
	timeoutTargets     map[string]uint32
	timeoutTargetsLock sync.RWMutex
}

// NewFuzzContext returns a new FuzzerContext
func NewFuzzContext() *FuzzContext {
	return &FuzzContext{
		runTaskCh:        make(chan *RunTask),
		fuzzingQueue:     list.New(),
		mainRecord:       EmptyRecord(),
		allRecordHashMap: make(map[string]struct{}),
		allBugID2Fp:      make(map[string]*BugMetrics),
		targetStages:     make(map[string]*TargetMetrics),
		timeoutTargets:   make(map[string]uint32),
		startAt:          time.Now(),
	}
}

func (c *FuzzContext) IterateQueryEntry() (*FuzzQueryEntry, error) {
	c.fqLock.RLock()
	if c.fuzzingQueue.Len() == 0 {
		c.fqLock.RUnlock()
		return nil, nil
	}
	if curQueueEntry == nil {
		curQueueEntry = c.fuzzingQueue.Front()
	} else {
		curQueueEntry = curQueueEntry.Next()
	}
	if curQueueEntry == nil {
		if c.fuzzingQueue.Front() == nil {
			// Should not happen. Dead loop outside.
			return nil, nil
		}
		curQueueEntry = c.fuzzingQueue.Front()
	}
	c.fqLock.RUnlock()

	elm := curQueueEntry.Value

	return elm.(*FuzzQueryEntry), nil
}

func (c *FuzzContext) EnqueueQueryEntry(e *FuzzQueryEntry) error {
	c.fqLock.Lock()
	c.fuzzingQueue.PushBack(e)
	queueLength := c.fuzzingQueue.Len()
	c.fqLock.Unlock()
	log.Printf("enqueued entry: %s, queue length currently (cur entry): %d", e, queueLength)
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
	defer c.bugID2FpLock.Unlock()
	c.allBugID2Fp[bugID] = &BugMetrics{
		FoundAt: time.Now(),
		Stdout:  filepath,
	}

}

func (c *FuzzContext) UpdateTargetStage(target string, stage FuzzStage) {
	c.targetStagesLock.Lock()
	defer c.targetStagesLock.Unlock()

	var track *TargetMetrics
	var exist bool
	if track, exist = c.targetStages[target]; !exist {
		track = &TargetMetrics{
			At: make(map[FuzzStage]time.Time),
		}
		c.targetStages[target] = track
	}

	if _, exist := track.At[stage]; !exist {
		track.At[stage] = time.Now()
	}
}

func (c *FuzzContext) UpdateTargetMaxCaseCov(target string, caseCov float32) {
	c.targetStagesLock.Lock()
	defer c.targetStagesLock.Unlock()

	var track *TargetMetrics
	var exist bool
	if track, exist = c.targetStages[target]; !exist {
		track = &TargetMetrics{
			At: make(map[FuzzStage]time.Time),
		}
		c.targetStages[target] = track
	}

	if caseCov > track.MaxCaseCov {
		track.MaxCaseCov = caseCov
	}
}

func (c *FuzzContext) RecordTargetTimeoutOnce(target string) {
	c.timeoutTargetsLock.Lock()
	defer c.timeoutTargetsLock.Unlock()

	c.timeoutTargets[target] += 1
}
