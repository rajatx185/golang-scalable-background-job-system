package main

import (
	"time"
)

// ✓ CORRECT - Global channel using var
var globalCh = make(chan int)

// ✓ CORRECT - Declare type, initialize in init()
var globalCh2 chan int

func init() {
	globalCh2 = make(chan int)
}

func main() {
	// ✓ CORRECT - Local channel using :=
	localCh := make(chan int) // unbuffered channel

	// 	// Define function
	// sendValue := func() {
	// 	localCh <- 42
	// }

	// // Launch as goroutine
	// go sendValue()

	go func() {
		time.Sleep(1 * time.Second)
		localCh <- 42 // Send value to channel
	}()

	value := <-localCh // Receive value from channel
	println("Received:", value)

	// 	| Time | Main Goroutine         | Worker Goroutine        |
	// | ---- | ---------------------- | ----------------------- |
	// | 1    | Print “receiving”      | not scheduled yet       |
	// | 2    | Block on `<-ch`        | Scheduler starts worker |
	// | 3    |                        | Print “sending”         |
	// | 4    |                        | Block on `ch <- "ping"` |
	// | 5    | Both resume together   | Value transferred       |
	// | 6    | Print “received: ping” | Print “sent”            |
	// “The send happens before the receive completes.”
	// But it does not guarantee any particular ordering of the next lines of code after that.

	// Buffered channel
	bufch := make(chan int, 2) // Buffer size of 1
	go func() {
		bufch <- 1
		println("Sent 1")
		time.Sleep(2 * time.Second) // Simulate some work
		bufch <- 2
		println("Sent 2")
		close(bufch) // Close channel when done
	}()

	println(<-bufch) // prints 1
	println(<-bufch) // prints 2
	v, ok := <- bufch
	if !ok {
		println("Channel closed, no more values")
	} else {
		println(v)
	}
}
