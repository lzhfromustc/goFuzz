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

func TestGetListOfBugIDFromStdoutContentSkipGoOracle(t *testing.T) {
	content := `-----New Bug:
goroutine 11 [running]:
runtime.TmpBeforeBlock()
        /home/luy70/go/src/runtime/myoracle_tmp.go:32 +0x90
google.golang.org/grpc.(*addrConn).resetTransport(0xc0004a3080)
        /home/luy70/goFuzz/src/grpc/clientconn.go:1482 +0xac6
created by google.golang.org/grpc.(*addrConn).connect
        /home/luy70/goFuzz/src/grpc/clientconn.go:1082 +0x12a
-----New Bug:
goroutine 10 [running]:
runtime.TmpBeforeBlock()
        /home/luy70/go/src/runtime/myoracle_tmp.go:32 +0x90
google.golang.org/grpc.(*ccBalancerWrapper).watcher(0xc0001ad270)
        /home/luy70/goFuzz/src/grpc/balancer_conn_wrappers.go:152 +0x815
created by google.golang.org/grpc.newCCBalancerWrapper
        /home/luy70/goFuzz/src/grpc/balancer_conn_wrappers.go:63 +0x1de
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/home/luy70/goFuzz/src/grpc/balancer_conn_wrappers.go:152") {
		t.Fail()
	}

	if !contains(bugIds, "/home/luy70/goFuzz/src/grpc/clientconn.go:1482") {
		t.Fail()
	}
}