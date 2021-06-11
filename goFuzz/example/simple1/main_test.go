package main

import (
	"fmt"
	gooracle "goFuzz/gooracle"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()
	gooracle.StoreOpInfo("ChMake", 0)
	ch := make(chan int)

	go func() {
		gooracle.StoreOpInfo("Send", 1)
		ch <- 1
	}()
	switch gooracle.ReadSelect("example/simple1/main_test.go", 16, 2) {
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
