package fuzzer

import "testing"

func TestParseGoCmdListOutput(t *testing.T) {
	output := `goFuzz/cmd/fuzz
goFuzz/cmd/instrument
goFuzz/config
goFuzz/example/simple1
goFuzz/fuzzer
goFuzz/gooracle
`
	packages, err := parseGoCmdListOutput(output)
	if err != nil {
		t.Fail()
	}
	if len(packages) != 6 {
		t.Fail()
	}

	if !contains(packages, "goFuzz/cmd/fuzz") {
		t.Fail()
	}
	if !contains(packages, "goFuzz/cmd/instrument") {
		t.Fail()
	}
	if !contains(packages, "goFuzz/config") {
		t.Fail()
	}
	if !contains(packages, "goFuzz/example/simple1") {
		t.Fail()
	}
	if !contains(packages, "goFuzz/fuzzer") {
		t.Fail()
	}
	if !contains(packages, "goFuzz/gooracle") {
		t.Fail()
	}
}

func TestParseGoCmdTestListOutput(t *testing.T) {
	output := `TestParseInputFileHappy
TestParseInputFileShouldFail
TestSelectInputHappy
TestSelectInputShouldFail
?   	k8s.io/kubernetes/cluster/images/etcd-version-monitor	[no test files]
ok      goFuzz/fuzzer   0.082s
`
	tests, err := parseGoCmdTestListOutput(output)
	if err != nil {
		t.Fail()
	}
	if len(tests) != 4 {
		t.Fail()
	}

	if !contains(tests, "TestParseInputFileHappy") {
		t.Fail()
	}

	if !contains(tests, "TestParseInputFileShouldFail") {
		t.Fail()
	}

	if !contains(tests, "TestSelectInputHappy") {
		t.Fail()
	}

	if !contains(tests, "TestSelectInputShouldFail") {
		t.Fail()
	}

}
