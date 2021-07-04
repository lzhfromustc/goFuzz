package fuzzer

import (
	"math"
	"testing"
)

const TOLERANCE = 0.000001

func TestParseChStatsFileContentHappy(t *testing.T) {
	chs, err := parseChStatsFileContent(`

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

func TestGetChannelCoverageHappy(t *testing.T) {
	chs := []string{
		"abc.go:1",
		"abc.go:2",
		"abc.go:3",
	}

	records := []ChanRecord{
		{
			ChID: "abc.go:1",
		},
		{
			ChID: "abc.go:2",
		},
	}
	cov := GetChannelCoverage(chs, records)

	if diff := math.Abs(float64(cov) - float64(2)/float64(3)); diff > TOLERANCE {
		t.Fail()
	}
}

func TestGetChannelCoverageEmpty(t *testing.T) {
	chs := []string{}

	records := []ChanRecord{
		{
			ChID: "abc.go:1",
		},
	}
	cov := GetChannelCoverage(chs, records)
	if cov != 0 {
		t.Fail()
	}

}
