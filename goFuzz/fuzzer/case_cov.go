package fuzzer

import (
	"fmt"
	"sync"
)

var (
	testID2cases     map[string]CaseCoverageTrack = make(map[string]CaseCoverageTrack)
	testID2casesLock sync.RWMutex
)

type CaseCoverageTrack struct {
	// total number of case combination
	totalCaseComb uint64
	// hash of a list of inputs(case combination) triggered
	triggeredCaseCombHashs map[string]uint32
	totalCases             map[string]bool
	triggeredCases         map[string]bool
}

func RecordTotalCases(testID string, selects []SelectInput) error {
	testID2casesLock.Lock()
	defer testID2casesLock.Unlock()
	err := recordTotalCases(testID2cases, testID, selects)
	return err
}

func recordTotalCases(testID2cases map[string]CaseCoverageTrack, testID string, selects []SelectInput) error {
	if _, exist := testID2cases[testID]; exist {
		return fmt.Errorf("duplicated record cases for %s", testID)
	}

	track := CaseCoverageTrack{
		totalCases:             make(map[string]bool),
		triggeredCaseCombHashs: make(map[string]uint32),
		triggeredCases:         make(map[string]bool),
		totalCaseComb:          1,
	}

	for _, s := range selects {
		base := fmt.Sprintf("%s:%d", s.StrFileName, s.IntLineNum)
		for i := 0; i < s.IntNumCase; i++ {
			selectID := fmt.Sprintf("%s:%d", base, i)
			track.totalCases[selectID] = true
		}
		track.totalCaseComb = uint64(s.IntNumCase) * track.totalCaseComb
	}

	testID2cases[testID] = track
	return nil
}

func RecordTriggeredCase(testID string, selects []SelectInput) error {
	testID2casesLock.Lock()
	defer testID2casesLock.Unlock()
	err := recordTriggeredCase(testID2cases, testID, selects)
	return err
}

func recordTriggeredCase(testID2cases map[string]CaseCoverageTrack, testID string, selects []SelectInput) error {

	track, exist := testID2cases[testID]
	if !exist {
		return fmt.Errorf("cannot find case track for %s", testID)
	}
	for _, s := range selects {
		base := fmt.Sprintf("%s:%d", s.StrFileName, s.IntLineNum)
		selectID := fmt.Sprintf("%s:%d", base, s.IntPrioCase)
		track.triggeredCases[selectID] = true
	}

	hash := GetHashOfSelects(selects)
	track.triggeredCaseCombHashs[hash] += 1

	return nil
}

func GetCountsOfSelects(testID string, selects []SelectInput) uint32 {
	testID2casesLock.Lock()
	defer testID2casesLock.Unlock()
	track, exist := testID2cases[testID]
	if !exist {
		return 0
	}
	hash := GetHashOfSelects(selects)

	count, exist := track.triggeredCaseCombHashs[hash]
	if !exist {
		return 0
	}
	return count
}

func GetCumulativeTriggeredCaseCoverage(testID string) (float32, float32, error) {
	testID2casesLock.RLock()
	defer testID2casesLock.RUnlock()
	cov, combCov, err := getCumulativeTriggeredCaseCoverage(testID2cases, testID)
	return cov, combCov, err
}

// getCumulativeTriggeredCaseCoverage returns the (case coverage, case comb coverage)
func getCumulativeTriggeredCaseCoverage(testID2cases map[string]CaseCoverageTrack, testID string) (float32, float32, error) {
	track, exist := testID2cases[testID]
	if !exist {
		return 0, 0, fmt.Errorf("cannot find case track for %s", testID)
	}
	total := len(track.totalCases)
	if total == 0 {
		return 0, 0, nil
	}

	count := 0
	for caseID := range track.triggeredCases {
		if _, exist := track.totalCases[caseID]; exist {
			count += 1
		}
	}

	combCount := len(track.triggeredCaseCombHashs)

	return float32(count) / float32(total), float32(combCount) / float32(track.totalCaseComb), nil
}
