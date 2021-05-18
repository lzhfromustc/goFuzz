package main

import (
	"fmt"
	gooracle "goFuzz/goFuzz/gooracle"
	"testing"
	"time"
)

func TestF1(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()

	ch := make(chan int)

	go func() {
		ch <- 1
	}()
	switch gooracle.ReadSelect("/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/main_test.go", 17, 2) {
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

func TestF2(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()

	ch := make(chan int)
	defer time.Sleep(5 * time.Second)
	defer close(ch)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- 1
	}()
	switch gooracle.ReadSelect("/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/main_test.go", 36, 2) {
	case 0:
		select {
		case <-ch:
			fmt.Print("Line 1")
		case <-gooracle.SelectTimeout():
			gooracle.StoreLastMySwitchChoice(-1)
			select {
			case <-ch:
				fmt.Print("Line 1")
			case <-time.After(1 * time.Second):
				fmt.Println("Buggy")
			}
		}
	case 1:
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Buggy")
		case <-gooracle.SelectTimeout():
			gooracle.StoreLastMySwitchChoice(-1)
			select {
			case <-ch:
				fmt.Print("Line 1")
			case <-time.After(1 * time.Second):
				fmt.Println("Buggy")
			}
		}
	default:
		gooracle.StoreLastMySwitchChoice(-1)
		select {
		case <-ch:
			fmt.Print("Line 1")
		case <-time.After(1 * time.Second):
			fmt.Println("Buggy")
		}
	}
}

func TestF3(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()

	ch := make(chan int)
	close(ch)
	close(ch)
}

func TestF4(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.AfterRun()

	var ch2 chan bool
	close(ch2)
}

func TestF5(t *testing.T) {
	gooracle.BeforeRun()
	defer gooracle.

		// max int64
		AfterRun()

	i := 9223372036854775807
	ch := make(chan int, 2*i)

	ch <- 1

}

func TestTmp(t *testing.T) {
	ch := make(chan int, 1)
	ch2 := make(chan int, 1)
	//ch <- 1
	select {
	case <- time.After( 1 * time.Second) :
		fmt.Println("First case")

	case <- time.After( 3 * time.Second) :
		fmt.Println("Second case")
	case <- time.After( 4 * time.Second) :
		fmt.Println("Third case")
	case ch <- 1:
		fmt.Println("Fifth case")
	case ch2 <- 1:
		fmt.Println("Sixth case")
	case <- time.After( 7 * time.Second) :
		fmt.Println("Fourth case")

	default:
		fmt.Println("Default case")
	}
}
