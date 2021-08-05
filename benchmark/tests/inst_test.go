package tests

import (
	"fmt"
	"testing"
	"time"
)

func TestOneSelect(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
	}()
	select {
	case <-ch:
		fmt.Println("Normal")
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Should be buggy")
	}
}
