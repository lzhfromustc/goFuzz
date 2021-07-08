package fuzzer

import (
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
	report := GetOperationCoverageReport(chs, records)

	if report.numOfChOp != 2 {
		t.Fail()
	}

	if report.numOfOtherPrimitivesOp != 0 {
		t.Fail()
	}

	if report.numOfChMake != 0 {
		t.Fail()
	}
}
