package fuzzer

import (
	"encoding/json"
	"log"
	"os"
	"time"
	// "github.com/edsrzf/mmap-go" TODO: use mmap to increase performance
)

type FuzzerMetrics struct {
	// Map from bug ID to stdout file
	Bugs                map[string]*BugMetrics
	NumOfBugsFound      uint64
	NumOfRuns           uint64
	NumOfFuzzQueryEntry uint64
	StartAt             time.Time
	// Seconds
	Duration uint64
}

type BugMetrics struct {
	FoundAt time.Time
	Stdout  string
}

func StreamMetrics(filePath string, intervalSec time.Duration) {
	go func() {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("failed to open metrics file: %v", err)
		}
		log.Printf("metrics file: %s", filePath)
		ticker := time.NewTicker(intervalSec * time.Second)

		for {
			<-ticker.C
			m := GetFuzzerMetrics(fuzzerContext)
			b, err := json.MarshalIndent(m, "", " ")
			if err != nil {
				log.Printf("failed to serialize metrics: %v", err)
				continue
			}
			if err := f.Truncate(0); err != nil {
				log.Printf("failed to truncate file: %v", err)
				continue
			}
			if _, err := f.Seek(0, 0); err != nil {
				log.Printf("failed to seek file: %v", err)
				continue
			}
			n, err := f.Write(b)
			if err != nil {
				log.Printf("failed to write to file: %v", err)
				continue
			}
			if n != len(b) {
				log.Printf("failed to write all metrics to file, epected %d, actial: %d", len(b), n)
				continue
			}
		}

	}()

}

func GetFuzzerMetrics(fuzzCtx *FuzzContext) *FuzzerMetrics {
	return &FuzzerMetrics{
		Bugs:                fuzzCtx.allBugID2Fp,
		NumOfFuzzQueryEntry: fuzzCtx.numOfFuzzQueryEntry,
		NumOfBugsFound:      fuzzCtx.numOfBugsFound,
		NumOfRuns:           fuzzCtx.numOfRuns,
		StartAt:             fuzzCtx.startAt,
		Duration:            uint64(time.Now().Sub(fuzzCtx.startAt).Seconds()),
	}
}
