package main

import (
	"fmt"
	"gotest/gooracle"
	"testing"
	"time"
)

func TestF1(t *testing.T) {
	gooracle.PreRun()
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
	time.Sleep(100 * time.Millisecond)
	ch := make(chan int)

	go func() {
		ch <- 1
	}()

	<-ch
}