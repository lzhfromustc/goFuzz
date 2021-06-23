package fuzzer

import (
	"testing"
)

func TestGetListOfBugIDFromStdoutContentHappy(t *testing.T) {
	content := `-----New Bug:
	goroutine 3855 [running]:
github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:372 +0x47a
	created by github.com/prometheus/prometheus/tsdb/wal.NewSize
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:302 +0x325
alkdfjalkdf
alsdkfjalsd
lkajdfadf
-----New Bug:
	goroutine 3855 [running]:
github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
	/Users/xsh/code/prometheus/tsdb/wal/wal1.go:372 +0x47a
	created by github.com/prometheus/prometheus/tsdb/wal.NewSize
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:302 +0x325
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if len(bugIds) != 2 {
		t.Fail()
	}
	if !contains(bugIds, "/Users/xsh/code/prometheus/tsdb/wal/wal.go:372") {
		t.Fail()
	}

	if !contains(bugIds, "/Users/xsh/code/prometheus/tsdb/wal/wal1.go:372") {
		t.Fail()
	}

}

func TestGetListOfBugIDFromStdoutContentEmpty(t *testing.T) {
	content := `
	goroutine 3855 [running]:
github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:372 +0x47a
	created by github.com/prometheus/prometheus/tsdb/wal.NewSize
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:302 +0x325
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if len(bugIds) != 0 {
		t.Fail()
	}
}

func TestGetListOfBugIDFromStdoutContentBad(t *testing.T) {
	content := `-----New Bug:
	goroutine 3855 [running]:
github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err == nil {
		t.Fail()
	}
	if bugIds != nil {
		t.Fail()
	}
}
