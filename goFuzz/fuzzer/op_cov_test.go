package fuzzer

import (
	"math"
	"testing"
)

const TOLERANCE = 0.000001

func TestParseOpCovFileContentHappy(t *testing.T) {
	res, err := parseOperationCoverageFileContent(`

1:chmake
2:chsend

`)
	if err != nil {
		t.Fail()
	}

	if _, exist := res["1"]; !exist {
		t.Fail()
	}

	if _, exist := res["2"]; !exist {
		t.Fail()
	}

}

func TestGetOperationCoverageHappy(t *testing.T) {
	chs := map[string]string{}
	chs["1"] = "chsend"
	chs["2"] = "chsend"
	chs["3"] = "chsend"

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
	chs := map[string]string{}

	records := []string{
		"1",
	}
	cov := GetOperationCoverage(chs, records)
	if cov != 0 {
		t.Fail()
	}

}
