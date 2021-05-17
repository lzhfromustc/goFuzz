package main

import (
	"fmt"
	"goFuzz/goFuzz/gooracle"
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

	select {
	case <-ch:
		fmt.Println("Normal")
	case <-time.After(300 * time.Millisecond):
		fmt.Println("Should be buggy")
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

	select {
	case <-ch:
		fmt.Print("Line 1")
	case <-time.After(1 * time.Second):
		fmt.Println("Buggy")
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
	defer gooracle.AfterRun()

	i := 9223372036854775807 // max int64
	ch := make(chan int, 2 * i)

	ch <- 1

}