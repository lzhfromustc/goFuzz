package fuzzer

import (
	"log"
	"math"
)

const (
	ScoreTupleCountLog2Increase = 1
	ScoreNewCh = 10
	ScoreNewClosed, ScoreNewNotClosed = 5, 5
	ScorePeakBufLog2Increase = 1
	ScoreBufFull = 10
)

// Rule:
// ScoreTupleCountLog2Increase: score for each +1 of log2(tuple_count) = 1
// ScoreNewCh: score for each new channel = 10
// ScoreNewClosed/ScoreNewNotClosed: score if this is the first time for a closed/notclosed status of existing channel = 5
// ScorePeakBufLog2Increase: score for each +1 of log2(peakBuf) = 1
// ScoreBufFull: score if this is the first time for buffer to be full: peakBuf equals to capBuf and not zero = 10
func ComputeScore(mainRecord, curRecord *Record) int {
	score := 0
	for tuple, count := range curRecord.MapTupleRecord {
		mainCount, exist := mainRecord.MapTupleRecord[tuple]
		if exist {
			// ScoreTupleCountLog2Increase: score for each +1 of log2(tuple_count)
			if mainCount < count { // the best record for this tuple
				mainCountLog := math.Log2(float64(mainCount))
				countLog := math.Log2(float64(count))
				score += (int(countLog) - int(mainCountLog)) * ScoreTupleCountLog2Increase
				log.Printf("Score_Log: ScoreTupleCountLog2Increase: %d", (int(countLog) - int(mainCountLog)))
			}
		} else {
			// ScoreTupleCountLog2Increase: score for each +1 of log2(tuple_count)
			//score += ScoreTupleCountLog2Increase
		}
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


			if mainChRecord.PeakBuf < chRecord.PeakBuf { // the best record for this channel
				// ScorePeakBufLog2Increase: score for each +1 of log2(peakBuf)
				//mainCountLog := math.Log2(float64(mainChRecord.PeakBuf))
				//countLog := math.Log2(float64(chRecord.PeakBuf))
				//score += (int(countLog) - int(mainCountLog)) * ScorePeakBufLog2Increase
				//log.Printf("Score_Log: ScorePeakBufLog2Increase: %d", int(countLog) - int(mainCountLog))

				// ScoreBufFull: score if this is the first time for buffer to fill out more than half;
				if chRecord.PeakBuf >= (chRecord.CapBuf / 2) && chRecord.CapBuf != 0 && chRecord.PeakBuf != chRecord.CapBuf{
					score += ScoreBufFull / 2
					log.Printf("Score_Log: ScoreBufHalf: %d", 1)
				}

				// ScoreBufFull: score if this is the first time for buffer to be full: peakBuf equals to capBuf and not zero
				if chRecord.PeakBuf == chRecord.CapBuf && chRecord.CapBuf != 0 {
					score += ScoreBufFull
					log.Printf("Score_Log: ScoreBufFull: %d", 1)
				}
			}

		} else {
			// ScoreNewCh: score for each new channel
			score += ScoreNewCh
			log.Printf("Score_Log: ScoreNewCh: %d", 1)
		}
	}

	return score
}
