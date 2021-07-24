package fuzzer

import (
	"testing"
)

func TestGetListOfBugIDFromStdoutContentHappy(t *testing.T) {
	content := `-----New Blocking Bug:
goroutine 3855 [running]:
github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:372 +0x47a
created by github.com/prometheus/prometheus/tsdb/wal.NewSize
	/Users/xsh/code/prometheus/tsdb/wal/wal.go:302 +0x325
alkdfjalkdf
alsdkfjalsd
lkajdfadf
-----New Blocking Bug:
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
	content := `-----New Blocking Bug:
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
	content := `-----New Blocking Bug:
goroutine 11 [running]:
runtime.TmpBeforeBlock()
	/home/luy70/go/src/runtime/myoracle_tmp.go:32 +0x90
google.golang.org/grpc.(*addrConn).resetTransport(0xc0004a3080)
	/home/luy70/goFuzz/src/grpc/clientconn.go:1482 +0xac6
created by google.golang.org/grpc.(*addrConn).connect
	/home/luy70/goFuzz/src/grpc/clientconn.go:1082 +0x12a
-----New Blocking Bug:
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

func TestGetListOfBugIDFromStdoutCausedByPanic(t *testing.T) {
	content := `
panic: send on closed channel

goroutine 7 [running]:
fuzzer-toy/blocking/grpc/1353.(*roundRobin).watchAddrUpdates(0xc00001c810)
	/fuzz/target/blocking/grpc/1353/grpc1353_test.go:84 +0x10f
fuzzer-toy/blocking/grpc/1353.(*roundRobin).Start.func1(0xc00001c810)
	/fuzz/target/blocking/grpc/1353/grpc1353_test.go:52 +0x35
created by fuzzer-toy/blocking/grpc/1353.(*roundRobin).Start
	/fuzz/target/blocking/grpc/1353/grpc1353_test.go:50 +0x91
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/fuzz/target/blocking/grpc/1353/grpc1353_test.go:84") {
		t.Fail()
	}

}

func TestGetListOfBugIDFromStdoutSkipPrimitive(t *testing.T) {
	content := `
-----New Blocking Bug:
goroutine 27 [running]:
sync.(*Mutex).Lock(0xc0000d6340)
	/usr/local/go/src/sync/mutex.go:77 +0x37
go.etcd.io/etcd/mvcc/backend.(*batchTx).safePending(0xc0000d6340, 0x0)
	/fuzz/target/mvcc/backend/batch_tx.go:231 +0x47
go.etcd.io/etcd/mvcc/backend.(*backend).run(0xc00011a090)
	/fuzz/target/mvcc/backend/backend.go:431 +0x265
created by go.etcd.io/etcd/mvcc/backend.newBackend
	/fuzz/target/mvcc/backend/backend.go:186 +0x511
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/fuzz/target/mvcc/backend/batch_tx.go:231") {
		t.Fail()
	}

}
