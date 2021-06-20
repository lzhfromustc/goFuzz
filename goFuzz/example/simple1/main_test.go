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
	gooracle.StoreChMakeInfo(ch, 33314)

	go func() {
		gooracle.StoreOpInfo("Send", 33315)
		ch <- 1
	}()
	switch gooracle.ReadSelect("/Users/xsh/code/goFuzz/goFuzz/example/simple1/main_test.go", 17, 2) {
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
