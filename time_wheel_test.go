package timewheel

import (
	"time"
	"fmt"
	"testing"
)

func TestNewTimeWheel(t *testing.T) {

	c := make(chan bool)

	tw := NewTimeWheel(1 *  time.Second, 10)
	fmt.Println(time.Now())
	tw.Add(2 * time.Second, func() {
		fmt.Println("hello world2")
		fmt.Println(time.Now())
	})
	tw.Add(2 * time.Second, func() {
		fmt.Println("hello world3")
		fmt.Println(time.Now())
	})
	tw.Add(11 * time.Second, func() {
		fmt.Println("hello world11")
		fmt.Println(time.Now())
		c <- true
	})

	<-c
}

