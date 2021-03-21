package main

import (
	"fmt"
	"testing"
)

func TestFuzz(t *testing.T) {

}

func testcase1() {
	// ch1 ID
	ch1 := make(chan int)
	ch2 := make(chan bool)

	go func() {

		ch1 <- 1
		// op record(ID)
		// send record(chID)
	}()

	go func() {
		ch2 <- true
	}()


	select {
	case <-ch1:
		// op record(ID)
		// first run: case record
		fmt.Println("Select 1 case 1")
	case <-ch2:
		fmt.Println("Select 1 case 2")
	}


}