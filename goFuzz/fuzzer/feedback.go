package fuzzer

import (
	"log"
)

const (
	ScoreTupleCountLog2 = 1
	ScoreCh                        = 10
	ScoreNewClosed, ScoreNotClosed = 10, 10
	ScoreBuf                       = 10
)

// ComputeScore Rule:
// Dependent with previous cases:
//		ScoreNewClosed/ScoreNotClosed: score if this is the FIRST TIME for a closed/notclosed status of existing channel = 10
// Independent with previous cases:
//		ScoreTupleCountLog2: For every detected tuple, scores = log2(tuple_num).
//		ScoreChNum: For every detected channel, we add score = 10
//		ScorePeakBuffer: The score for each peak buffer would be: score = 10 * (PeakBuffer / BufferSize).
func ComputeScore(mainRecord map[string]*Record, curRecord *Record, runResult *RunResult, id string, prevID string) int {
	score := 0
	var tupleCountScore = 0

	var tupleCount = 0
	var tupleNum = 0
	var notclosedNum = 0
	var peakBuffer = 0
	var capBuffer = 0
	var channelNum = 0


	if curRecord == nil {
		return 0
	}

	for _, count := range curRecord.MapTupleRecord {
		score += int(count) * ScoreTupleCountLog2
		tupleCountScore += int(count) * ScoreTupleCountLog2
		tupleCount += int(count)
		tupleNum += 1
	}

	var notClosedScore = 0
	var bufferScore = 0
	var chScore = 0

	for _, chRecord := range curRecord.MapChanRecord {

		// ScoreNewClosed/ScoreNotClosed: score if this is the first time for a closed/notclosed status of existing channel
		if chRecord.NotClosed == true {
			score += ScoreNotClosed
			notClosedScore += ScoreNotClosed
			notclosedNum += 1
		}

		// ScoreBuf: ScoreBuffer * BufferPercentage
		if chRecord.PeakBuf > 0 && chRecord.CapBuf != 0 {
			bufferPer := float64(chRecord.PeakBuf) / float64(chRecord.CapBuf)
			score += int(float64(ScoreBuf) * bufferPer)
			bufferScore += int(float64(ScoreBuf) * bufferPer)
			peakBuffer += chRecord.PeakBuf
			capBuffer += chRecord.CapBuf
		}

		// ScoreCh: score for each detected channel
		score += ScoreCh
		chScore += ScoreCh
		channelNum += 1
	}

	log.Printf("Score_Log: For stdout case: %s, ScoreTupleCountLog2: %d", runResult.StdoutFilepath, tupleCountScore)
	log.Printf("Score_Log: For stdout case: %s, ScoreNotClosed: %d", runResult.StdoutFilepath, notClosedScore)
	log.Printf("Score_Log: For stdout case: %s, ScoreBuf: %d", runResult.StdoutFilepath, bufferScore)
	log.Printf("Score_Log: For stdout case: %s, ScoreCh: %d", runResult.StdoutFilepath, chScore)

	log.Printf("SScore_LOG: For stdout case: %s, Tuple Count: %d, Tuple Num: %d, Not closed: %d, Buffer Size: %d, Buffer Max Size: %d, Channel Num: %d",
		runResult.StdoutFilepath, tupleCount, tupleNum, notclosedNum, peakBuffer, capBuffer, channelNum )

	return score
}
