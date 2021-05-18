package gooracle

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBenchGLDetect(t *testing.T) {
	BenchGLDetect(2,2,2,3 * time.Second, 0)
}

func TestGLDetect(t *testing.T) {
	cg := NewGoroutine()
	defer cg.RemoveAllRef()

	ch := make(chan int)
	cI := NewChanInfo(ch)
	AddRefGoroutine(cI, cg)

	go func() {
		cg := NewGoroutine()
		defer cg.RemoveAllRef()
		AddRefGoroutine(FindChanInfo(ch), cg)
		defer FindChanInfo(ch).CheckBlockBug()

		time.Sleep(2 * time.Second)
		cg.SetBlockAt(ch, Send)
		FindChanInfo(ch).CheckBlockBug()
		ch <- 1
		cg.WithdrawBlock()
	}()

	cg.SetBlockAt(ch, Recv)
	FindChanInfo(ch).CheckBlockBug()
	select {
	case <-ch:
		cg.WithdrawBlock()
		fmt.Println("Case 1")
	case <- time.After(0):
		cg.WithdrawBlock()
		fmt.Println("Case 2")
	}




	time.Sleep(5 * time.Second)

	RemoveRefGoroutine(cI, cg)
	FindChanInfo(ch).CheckBlockBug()

}

func TestGL(t *testing.T) {
	defer runtime.TmpDumpBlockingInfo()
	ch := make(chan int)
	go func() {
		time.Sleep(time.Second)
		ch <- 1
	}()
	select {
	case <-ch:
	default:
	}
	time.Sleep(time.Second * 2)
}

func TestCICS(t *testing.T) {
	defer runtime.TmpDumpBlockingInfo()
	ch := make(chan int)
	mu := sync.Mutex{}
	go func() {
		mu.Lock()
		ch <- 1
		mu.Unlock()
	}()
	go func() {
		mu.Lock()
		<-ch
		mu.Unlock()
	}()
	time.Sleep(time.Second * 2)
}

func Test1(t *testing.T) {
	defer runtime.TmpDumpBlockingInfo()
	MapInput["\t/data/ziheng/shared/gotest/gotest/src/gotest/gooracle/gooracle_test.go:109"] = runtime.SelectInfo{
		StrFileName: "\t/data/ziheng/shared/gotest/gotest/src/gotest/gooracle/gooracle_test.go",
		StrLineNum:  "109",
		IntNumCase:  2,
		IntPrioCase: 1,
	}
	SelectDelayMS = 1000

	ch := make(chan int)
	ch2 := make(chan int)
	go func() {
		ch <- 1
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("After sleep")
		ch2 <- 1
	}()

	select {
	case <-ch:
		fmt.Println("Normal")
	case <-ch2:
		fmt.Println("Buggy")
	}
}

