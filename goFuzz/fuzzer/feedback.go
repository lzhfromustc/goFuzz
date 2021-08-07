package fuzzer

import (
	"log"
	"math"
)

const (
	ScoreTupleCountLog2 = 1
	ScoreCh = 10
	ScoreNewClosed, ScoreNewNotClosed = 10, 10
	ScoreBuf = 10
)

// ComputeScore Rule:
// Dependent with previous cases:
//		ScoreNewClosed/ScoreNewNotClosed: score if this is the FIRST TIME for a closed/notclosed status of existing channel = 10
// Independent with previous cases:
//		ScoreTupleCountLog2: For every detected tuple, scores = log2(tuple_num).
//		ScoreChNum: For every detected channel, we add score = 10
//		ScorePeakBuffer: The score for each peak buffer would be: score = 10 * (PeakBuffer / BufferSize).
func ComputeScore(mainRecord, curRecord *Record) int {
	score := 0
	for _, count := range curRecord.MapTupleRecord {
		countLog := math.Log2(float64(count))
		score += int(countLog) * ScoreTupleCountLog2
		log.Printf("Score_Log: ScoreTupleCountLog2: %d", int(countLog))
	}

	for chID, chRecord := range curRecord.MapChanRecord {
		mainChRecord, exist := mainRecord.MapChanRecord[chID]

		if exist {
			// ScoreNewClosed/ScoreNewNotClosed: score if this is the first time for a closed/notclosed status of existing channel
			if mainChRecord.Closed == false && chRecord.Closed == true {
				score += ScoreNewClosed
				log.Printf("Score_Log: ScoreNewClosed: %d", 1)
			}
			if mainChRecord.NotClosed == false && chRecord.NotClosed == true {
				score += ScoreNewNotClosed
				log.Printf("Score_Log: ScoreNewNotClosed: %d", 1)
			}
		}

		// ScoreBuf: ScoreBuffer * BufferPercentage
		if chRecord.PeakBuf > 0 && chRecord.CapBuf != 0 {
			bufferPer := chRecord.PeakBuf / chRecord.CapBuf
			score += ScoreBuf * bufferPer
			log.Printf("Score_Log: ScoreBuf: %d", ScoreBuf * bufferPer)
		}

		// ScoreCh: score for each detected channel
		score += ScoreCh
		log.Printf("Score_Log: ScoreCh: %d", 1)
	}

	return score
}
