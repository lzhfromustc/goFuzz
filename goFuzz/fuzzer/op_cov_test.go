package fuzzer

import (
	"math"
	"testing"
)

const TOLERANCE = 0.000001

func TestParseChStatsFileContentHappy(t *testing.T) {
	chs, err := parseOperationCoverageFileContent(`

abc:1
alsdkf:2

`)
	if err != nil {
		t.Fail()
	}

	if !contains(chs, "abc:1") {
		t.Fail()
	}

	if !contains(chs, "alsdkf:2") {
		t.Fail()
	}

}

func TestGetOperationCoverageHappy(t *testing.T) {
	chs := []string{
		"1",
		"2",
		"3",
	}

	records := []string{
		"1",
		"2",
	}
	cov := GetOperationCoverage(chs, records)

	if diff := math.Abs(float64(cov) - float64(2)/float64(3)); diff > TOLERANCE {
		t.Fail()
	}
}

func TestGetChannelCoverageEmpty(t *testing.T) {
	chs := []string{}

	records := []string{
		"1",
	}
	cov := GetOperationCoverage(chs, records)
	if cov != 0 {
		t.Fail()
	}

}
