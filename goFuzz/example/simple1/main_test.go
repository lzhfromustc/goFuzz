package main

import (
	"fmt"
	gooracle "gooracle"
	"sync"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()
	ch := make(chan int)
	gooracle.StoreChMakeInfo(ch, 8948)
	go func() {
		gooracle.StoreOpInfo("Send", 8949)
		ch <- 1
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	gooracle.RecordWgUniqueCall(&wg, 8950)
	wg.Done()
	gooracle.RecordWgUniqueCall(&wg, 8951)
	switch gooracle.ReadSelect("/Users/xsh/code/goFuzz/goFuzz/example/simple1/main_test.go", 19, 2) {
	case 0:
		select {
		case <-ch:
			fmt.Println("Normal")
		case <-gooracle.SelectTimeout():
			gooracle.StoreLastMySwitchChoice(-1)
			select {
			case <-ch:
				fmt.Println("Normal")
			case <-time.After(300 * time.Millisecond):
				fmt.Println("Should be buggy")
			}
		}
	case 1:
		select {
		case <-time.After(300 * time.Millisecond):
			fmt.Println("Should be buggy")
		case <-gooracle.SelectTimeout():
			gooracle.StoreLastMySwitchChoice(-1)
			select {
			case <-ch:
				fmt.Println("Normal")
			case <-time.After(300 * time.Millisecond):
				fmt.Println("Should be buggy")
			}
		}
	default:
		gooracle.StoreLastMySwitchChoice(-1)
		select {
		case <-ch:
			fmt.Println("Normal")
		case <-time.After(300 * time.Millisecond):
			fmt.Println("Should be buggy")
		}
	}
}
