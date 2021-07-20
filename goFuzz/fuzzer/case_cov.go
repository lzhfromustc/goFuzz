package fuzzer

import (
	"fmt"
)

var (
	testID2cases map[string]CaseCoverageTrack = make(map[string]CaseCoverageTrack)
)

type CaseCoverageTrack struct {
	totalCases     map[string]bool
	triggeredCases map[string]bool
}

func RecordTotalCases(testID2cases map[string]CaseCoverageTrack, testID string, selects []SelectInput) error {
	if _, exist := testID2cases[testID]; exist {
		return fmt.Errorf("duplicated record cases for %s", testID)
	}

	track := CaseCoverageTrack{
		totalCases:     make(map[string]bool),
		triggeredCases: make(map[string]bool),
	}

	for _, s := range selects {
		base := fmt.Sprintf("%s:%d", s.StrFileName, s.IntLineNum)
		for i := 0; i < s.IntNumCase; i++ {
			selectID := fmt.Sprintf("%s:%d", base, i)
			track.totalCases[selectID] = true
		}
	}

	testID2cases[testID] = track
	return nil
}

func RecordTriggeredCase(testID2cases map[string]CaseCoverageTrack, testID string, selects []SelectInput) error {

	track, exist := testID2cases[testID]
	if !exist {
		return fmt.Errorf("cannot find case track for %s", testID)
	}
	for _, s := range selects {
		base := fmt.Sprintf("%s:%d", s.StrFileName, s.IntLineNum)
		selectID := fmt.Sprintf("%s:%d", base, s.IntPrioCase)
		track.triggeredCases[selectID] = true
	}

	return nil
}

func GetCumulativeTriggeredCaseCoverage(testID2cases map[string]CaseCoverageTrack, testID string) (float32, error) {
	track, exist := testID2cases[testID]
	if !exist {
		return 0, fmt.Errorf("cannot find case track for %s", testID)
	}
	total := len(track.totalCases)
	if total == 0 {
		return 0, nil
	}

	count := 0
	for caseID := range track.triggeredCases {
		if _, exist := track.totalCases[caseID]; exist {
			count += 1
		}
	}

	return float32(count) / float32(total), nil
}
