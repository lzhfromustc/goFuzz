package fuzzer

import (
	"fmt"
	"math"
	"testing"
)

func TestRecordTotalCasesHappy(t *testing.T) {
	testID2cases := make(map[string]CaseCoverageTrack)

	selects := []SelectInput{
		{
			StrFileName: "some.go",
			IntLineNum:  1,
			IntPrioCase: 1,
			IntNumCase:  5,
		},

		{
			StrFileName: "other.go",
			IntLineNum:  2,
			IntPrioCase: 1,
			IntNumCase:  3,
		},
	}
	err := recordTotalCases(testID2cases, "TestAbc", selects)

	if err != nil {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].totalCases) != 8 {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].triggeredCases) != 0 {
		t.Fail()
	}

	if _, exist := testID2cases["TestAbc"].totalCases["other.go:2:0"]; !exist {
		t.Fail()
	}
	if _, exist := testID2cases["TestAbc"].totalCases["other.go:2:1"]; !exist {
		t.Fail()
	}

	if _, exist := testID2cases["TestAbc"].totalCases["other.go:2:2"]; !exist {
		t.Fail()
	}
}

func TestRecordTriggeredCaseHappy(t *testing.T) {
	testID2cases := make(map[string]CaseCoverageTrack)

	selects := []SelectInput{
		{
			StrFileName: "some.go",
			IntLineNum:  1,
			IntPrioCase: 1,
			IntNumCase:  5,
		},

		{
			StrFileName: "other.go",
			IntLineNum:  2,
			IntPrioCase: 0,
			IntNumCase:  3,
		},
	}

	err := recordTotalCases(testID2cases, "TestAbc", selects)

	if err != nil {
		t.Fail()
	}

	err = recordTriggeredCase(testID2cases, "TestAbc", selects)

	if err != nil {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].triggeredCases) != 2 {
		t.Fail()
	}

	moreSelects := []SelectInput{
		{
			StrFileName: "some.go",
			IntLineNum:  1,
			IntPrioCase: 1,
			IntNumCase:  5,
		},

		{
			StrFileName: "some.go",
			IntLineNum:  2,
			IntPrioCase: 1,
			IntNumCase:  3,
		},

		{
			StrFileName: "some.go",
			IntLineNum:  2,
			IntPrioCase: 2,
			IntNumCase:  3,
		},
	}

	err = recordTriggeredCase(testID2cases, "TestAbc", moreSelects)

	if err != nil {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].triggeredCases) != 4 {
		t.Fail()
	}

	if _, exist := testID2cases["TestAbc"].triggeredCases["some.go:2:1"]; !exist {
		t.Fail()
	}

	if _, exist := testID2cases["TestAbc"].triggeredCases["some.go:2:2"]; !exist {
		t.Fail()
	}
}

func TestGetCumulativeTriggeredCaseCoverageHappy(t *testing.T) {
	testID2cases := make(map[string]CaseCoverageTrack)

	selects := []SelectInput{
		{
			StrFileName: "some.go",
			IntLineNum:  1,
			IntPrioCase: 1,
			IntNumCase:  5,
		},

		{
			StrFileName: "other.go",
			IntLineNum:  2,
			IntPrioCase: 0,
			IntNumCase:  3,
		},
	}

	err := recordTotalCases(testID2cases, "TestAbc", selects)

	if err != nil {
		t.Fail()
	}

	moreSelects := []SelectInput{
		{
			StrFileName: "some.go",
			IntLineNum:  1,
			IntPrioCase: 1,
			IntNumCase:  5,
		},

		{
			StrFileName: "other.go",
			IntLineNum:  2,
			IntPrioCase: 1,
			IntNumCase:  3,
		},

		{
			StrFileName: "other.go",
			IntLineNum:  2,
			IntPrioCase: 2,
			IntNumCase:  3,
		},
	}

	err = recordTriggeredCase(testID2cases, "TestAbc", moreSelects)
	if err != nil {
		t.Fail()
	}

	cov, err := getCumulativeTriggeredCaseCoverage(testID2cases, "TestAbc")
	if err != nil {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].totalCases) != 8 {
		t.Fail()
	}

	if len(testID2cases["TestAbc"].triggeredCases) != 3 {
		fmt.Printf("expect # of triggeredCases is %d, actual %d\n", 3, len(testID2cases["TestAbc"].triggeredCases))
		t.Fail()
	}

	if !(math.Abs(float64(cov-float32(3)/float32(8))) < 0.0000001) {
		fmt.Printf("expect cov is %0.2f, actual %0.2f\n", float32(3)/float32(8), cov)
		t.Fail()
	}

}
