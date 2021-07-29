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

func TestGetListOfBugIDFromPanicWithPanicStackTrace(t *testing.T) {

	content := `
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x58 pc=0x900610]

goroutine 19 [running]:
testing.tRunner.func1.2(0x9df6c0, 0xe4a060)
	/usr/local/go/src/testing/testing.go:1143 +0x335
testing.tRunner.func1(0xc00034d180)
	/usr/local/go/src/testing/testing.go:1146 +0x4c2
panic(0x9df6c0, 0xe4a060)
	/usr/local/go/src/runtime/panic.go:965 +0x1b9
github.com/docker/docker/client.(*Client).ClientVersion(...)
	/go/src/github.com/docker/docker/client/client.go:197
github.com/docker/docker/client.TestNewClientWithOpsFromEnv(0xc00034d180)
	/go/src/github.com/docker/docker/client/client_test.go:100 +0x690
testing.tRunner(0xc00034d180, 0xaa34b8)
	/usr/local/go/src/testing/testing.go:1193 +0xef
created by testing.(*T).Run
	/usr/local/go/src/testing/testing.go:1238 +0x2b5`

	bugIds, err := GetListOfBugIDFromStdoutContent(content)
	if err != nil {
		t.Fail()
	}
	if bugIds == nil {
		t.Fail()
	}

	if !contains(bugIds, "/go/src/github.com/docker/docker/client/client.go:197") {
		t.Fail()
	}

}
