package main

import (
	"fmt"
	gooracle "gooracle"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()
	ch := make(chan int)
	gooracle.StoreChMakeInfo(ch, 28421)

	go func() {
		gooracle.StoreOpInfo("Send", 28422)
		ch <- 1
	}()
	switch gooracle.ReadSelect("/data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/example/simple1/main_test.go", 15, 2) {
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
