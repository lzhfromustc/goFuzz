package fuzzer

import (
	"testing"
)

func TestGetListOfBugIDFromStdoutContentHappyNonBlocking(t *testing.T) {
	content := `=== RUN   TestGrpc1687
-----New NonBlocking Bug:
---Stack:
goroutine 9 [running]:
runtime.ReportNonBlockingBug(...)
	/usr/local/go/src/runtime/myoracle.go:538
command-line-arguments.(*serverHandlerTransport).do(0xc00005a560, 0x55efd8)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:58 +0x2e5
command-line-arguments.(*serverHandlerTransport).Write(0xc00005a560)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:75 +0x37
command-line-arguments.TestGrpc1687.func1(0xc00005a570)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:177 +0x45
created by command-line-arguments.testHandlerTransportHandleStreams.func1
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:169 +0x3b

-----End Bug
panic: send on closed channel

goroutine 9 [running]:
command-line-arguments.(*serverHandlerTransport).do(0xc00005a560, 0x55efd8)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:58 +0x2e5
command-line-arguments.(*serverHandlerTransport).Write(0xc00005a560)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:75 +0x37
command-line-arguments.TestGrpc1687.func1(0xc00005a570)
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:177 +0x45
created by command-line-arguments.testHandlerTransportHandleStreams.func1
	/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:169 +0x3b

Process finished with exit code 1`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if len(bugIds) != 1 {
		t.Fail()
	}
	if !contains(bugIds, "/data/ziheng/shared/gotest/empirical/gobench/gobench/goker/nonblocking/grpc/1687/grpc1687_test.go:58") {
		t.Fail()
	}
}

func TestGetListOfBugIDFromStdoutContentHappyBlocking(t *testing.T) {
	content := `=== RUN   TestBalancerUnderBlackholeNoKeepAliveDelete
Check bugs after 2 minutes
Failed to create file: 
-----New Blocking Bug:
---Blocking location:
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/v3_server.go:838
---Primitive location:
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/server.go:806
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/server.go:809
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/server.go:1215
---Primitive pointer:
0xc000442540
0xc000442580
0xc001d8bf40
-----End Bug
-----New Blocking Bug:
---Blocking location:
/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229
---Primitive location:
/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:135
---Primitive pointer:
0xc001affc40
-----End Bug
{"level":"warn","ts":"2021-07-26T23:25:16.731-0400","caller":"clientv3/retry_interceptor.go:62","msg":"retrying of unary invoker failed","target":"endpoint://client-c437f924-48d8-4d97-a6e9-ed535f5f4462/localhost:37946441868244577890","attempt":0,"error":"rpc error: code = DeadlineExceeded desc = context deadline exceeded"}
    black_hole_test.go:277: #1: current error expected error
-----Withdraw prim:0xc000442540
{"level":"warn","ts":"2021-07-26T23:25:16.731-0400","caller":"clientv3/retry_interceptor.go:62","msg":"retrying of unary invoker failed","target":"endpoint://client-c437f924-48d8-4d97-a6e9-ed535f5f4462/localhost:37946441868244577890","attempt":0,"error":"rpc error: code = DeadlineExceeded desc = context deadline exceeded"}
    black_hole_test.go:277: #1: current error expected error
-----New Blocking Bug:
---Blocking location:
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/v3_server.go:123456
---Primitive location:
/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/server.go:806
---Primitive pointer:
0xc000442540
-----End Bug
--- PASS: TestBalancerUnderBlackholeNoKeepAliveDelete (10.26s)
PASS`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if len(bugIds) != 2 {
		t.Fail()
	}
	if !contains(bugIds, "/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229") {
		t.Fail()
	}
	if !contains(bugIds, "/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/etcdserver/v3_server.go:123456") {
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

// Deprecated: our bug format is changed
func TestGetListOfBugIDFromStdoutContentBad(t *testing.T) {
	t.SkipNow()
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

// Deprecated: our bug format now doesn't contain the oracle
func TestGetListOfBugIDFromStdoutContentSkipGoOracle(t *testing.T) {
	t.SkipNow()
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

// Deprecated: our bug format now doesn't contain lines like mutex.go:77
func TestGetListOfBugIDFromStdoutSkipPrimitive(t *testing.T) {
	t.SkipNow()
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

func TestGetListOfBugIDFromStdoutSkipTimeout(t *testing.T) {
	content := `
-----New Blocking Bug:
---Blocking location:
/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229
---Primitive location:
/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:135
---Primitive pointer:
0xc001affc40
-----End Bug

panic: test timed out after 1m0s

goroutine 468 [running]:
testing.(*M).startAlarm.func1()
	/usr/local/go/src/testing/testing.go:1700 +0xe5
created by time.goFunc
	/usr/local/go/src/time/sleep.go:180 +0x45
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/data/ziheng/shared/gotest/stubs/etcd/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229") {
		t.Fail()
	}

}

// Deprecated: our bug format now doesn't contain offset
func TestGetListOfBugIDFromStdoutNoOffset(t *testing.T) {
	t.SkipNow()
	content := `
-----New Blocking Bug:
goroutine 388 [running]:
github.com/soheilhy/cmux.muxListener.Accept(...)
	/go/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229
net/http.(*Server).Serve(0xc0003467e0, 0x11bf5b0, 0xc00043c7e0, 0x0, 0x0)
	/usr/local/go/src/net/http/server.go:2981 +0x285
net/http/httptest.(*Server).goServe.func1(0xc0002f06e0)
	/usr/local/go/src/net/http/httptest/server.go:308 +0x6e
created by net/http/httptest.(*Server).goServe
	/usr/local/go/src/net/http/httptest/server.go:306 +0x5c
	`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/go/pkg/mod/github.com/soheilhy/cmux@v0.1.4/cmux.go:229") {
		t.Fail()
	}

}
