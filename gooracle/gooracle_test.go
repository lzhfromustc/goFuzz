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
	defer runtime.DumpBlockingInfo()
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
	defer runtime.DumpBlockingInfo()
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
	runtime.TimerTest()
}

