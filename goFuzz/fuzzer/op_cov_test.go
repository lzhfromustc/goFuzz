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

func TestGetCurrOpIDCoverageReportHappy(t *testing.T) {
	chs := map[string]string{}
	chs["1"] = "chsend"
	chs["2"] = "chsend"
	chs["3"] = "chsend"

	records := []string{
		"1",
		"2",
	}
	report := getCurrOpIDCoverageReport(chs, records)

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

func TestGetTriggeredOperationCoverageHappy(t *testing.T) {
	chs := map[string]string{}
	chs["1"] = "chmake"
	chs["2"] = "chsend"
	chs["3"] = "chclose"
	chs["4"] = "chclose"

	records := map[string]bool{
		"1": true,
		"2": true,
		"4": true,
	}
	report := getTriggeredOpIDCoverageReport(chs, records)

	if report.numOfChOp != 2 {
		t.Fail()
	}

	if report.numOfOtherPrimitivesOp != 0 {
		t.Fail()
	}

	if report.numOfChMake != 1 {
		t.Fail()
	}
}

func TestUpdateTriggeredOpID(t *testing.T) {
	record := map[string]bool{}
	ids := []string{
		"1",
		"3",
		"9",
	}
	updateTriggeredOpID(record, ids)
	if _, exist := record["1"]; !exist {
		t.Fail()
	}

	if _, exist := record["3"]; !exist {
		t.Fail()
	}

	if _, exist := record["9"]; !exist {
		t.Fail()
	}
}
