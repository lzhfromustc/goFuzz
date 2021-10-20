package fuzzer

import (
	"log"
	"math"
	"strconv"
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
	var curTupleCount = 0
	var preTupleCount = 0
	var curTupleNum = 0
	var preTupleNum = 0

	var tupleCountScore = 0
	var tupleNumScore = 0

	for _, count := range runResult.RetRecord.MapTupleRecord {
		curTupleCount += count
		curTupleNum += 1
		log.Printf("Log: Get tuples count: " + strconv.Itoa(count))
	}

	log.Printf("In current " + id + ", prevID:" + prevID + " curTupleCount: " + strconv.Itoa(curTupleCount) + "curTupleNum: " + strconv.Itoa(curTupleNum))

	log.Printf("In computeScore, getting mainRecord size: " + strconv.Itoa(len(mainRecord)))

	curMainRecord := mainRecord[prevID]

	// If the previous main record is not nil. We can calculate the current record based on
	// tuple differences. Otherwise, use score = 0
	if curMainRecord != nil {
		for _, count := range curMainRecord.MapTupleRecord {
			countLog := math.Log2(float64(count))
			if int(countLog) != -9223372036854775808 {
				preTupleCount += int(countLog)
				preTupleNum += 1
			}
		}

		tupleCountScore = curTupleCount - preTupleCount
		if tupleCountScore < 0 {
			tupleCountScore = - tupleCountScore
		}
		tupleNumScore = curTupleNum - preTupleNum
		if tupleNumScore < 0 {
			tupleNumScore = - tupleNumScore
		}
	} else {
		log.Printf("MainRecord is NULL. ")
	}

	log.Printf("In current " + id + ", prevID:" + prevID + " curTupleCount: " + strconv.Itoa(curTupleCount) + "curTupleNum: " + strconv.Itoa(curTupleNum) + " preTupleCount: " + strconv.Itoa(preTupleCount) + " preTupleNum: " + strconv.Itoa(preTupleNum))

	// Write curRecord, use it to save the current Record for the next run.
	mainRecord[id] = curRecord

	score = tupleCountScore + tupleNumScore

	//var notClosedScore = 0
	//var bufferScore = 0
	//var chScore = 0
	//
	//for _, chRecord := range curRecord.MapChanRecord {
	//
	//	// ScoreNewClosed/ScoreNotClosed: score if this is the first time for a closed/notclosed status of existing channel
	//	if chRecord.NotClosed == true {
	//		score += ScoreNotClosed
	//		notClosedScore += ScoreNotClosed
	//	}
	//
	//	// ScoreBuf: ScoreBuffer * BufferPercentage
	//	if chRecord.PeakBuf > 0 && chRecord.CapBuf != 0 {
	//		bufferPer := chRecord.PeakBuf / chRecord.CapBuf
	//		score += ScoreBuf * bufferPer
	//		bufferScore += ScoreBuf * bufferPer
	//	}
	//
	//	// ScoreCh: score for each detected channel
	//	score += ScoreCh
	//	chScore += ScoreCh
	//}

	log.Printf("Score_Log: For stdout case: %s, ScoreTupleCountLog2: %d", runResult.StdoutFilepath, tupleCountScore)
	log.Printf("Score_Log: For stdout case: %s, ScoreTupleNum: %d", runResult.StdoutFilepath, tupleNumScore)
	//log.Printf("Score_Log: For stdout case: %s, ScoreNotClosed: %d", runResult.StdoutFilepath, notClosedScore)
	//log.Printf("Score_Log: For stdout case: %s, ScoreBuf: %d", runResult.StdoutFilepath, bufferScore)
	//log.Printf("Score_Log: For stdout case: %s, ScoreCh: %d", runResult.StdoutFilepath, chScore)

	return score
}
