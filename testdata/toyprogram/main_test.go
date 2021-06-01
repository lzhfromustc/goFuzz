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

func TestTmp2(t *testing.T) {

	ch := make(chan int)
	go func() {
		defer func() {
			fmt.Println("Child 1 is exiting")
		}()
		ch <- 1
	}()
	go func() {
		defer func() {
			fmt.Println("Child 2 is exiting")
		}()
	}()

	<-ch
	time.Sleep(time.Second)
	fmt.Println("Finished")
}

func TestTmp(t *testing.T) {
	ch := make(chan int, 1)
	ch2 := make(chan int, 1)
	ch3 := make(chan int, 1)
	ch4 := make(chan int, 1)
	ch5 := make(chan int, 1)
	ch6 := make(chan int, 1)
	go func() {
		for {
			<-ch
		}
	}()
	go func() {
		for {
			<-ch2
		}
	}()
	go func() {
		for {
			ch3 <- 1
		}
	}()
	go func() {
		for {
			ch4 <- 1
		}
	}()
	go func() {
		for {
			ch5 <- 1
		}
	}()
	go func() {
		for {
			ch6 <- 1
		}
	}()

	for {
		select {
		case <-ch3: //6
			fmt.Println("First case")
		case <-ch4: //5
			fmt.Println("Second case")
		case <-ch5: //4
			fmt.Println("Third case")
		case ch <- 1: //0
			fmt.Println("Fourth case")
		default: //-1
			fmt.Println("Default case")
		case ch2 <- 1:  //1
			fmt.Println("Fifth case")
		case <-ch6: //3
			fmt.Println("Sixth case")
		case ch <- 1: //2
			fmt.Println("7th case")
		}
	}
}
