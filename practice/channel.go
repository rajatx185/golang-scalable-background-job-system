package main

import (
	"time"
)

// ✓ CORRECT - Global channel using var
// var globalCh = make(chan int)

// ✓ CORRECT - Declare type, initialize in init()
// var globalCh2 chan int

// func init() {
// 	globalCh2 = make(chan int)
// }

// func main() {
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
	v, ok := <-bufch
	if !ok {
		println("Channel closed, no more values")
	} else {
		println(v)
	}

	// Loop over channel

	loopCh := make(chan int) // Unbuffered channel
	go func() {
		for i := 1; i < 5; i++ {
			loopCh <- i // Send and blocks until consumed
		}
		close(loopCh)
	}()

	for val := range loopCh { // Exits when channel closed
		println("Looped value:", val) // Receive until channel closed
	}

	// Select statement with channels
	globalCh := make(chan int)
	// globalCh2 := make(chan int)

	// Simulate sending messages to channels

	go func() {
		time.Sleep(3 * time.Second)
		globalCh <- 100
		// time.Sleep(1 * time.Second)
		// globalCh2 <- 200
	}()

	// time.Sleep(500 * time.Millisecond) // Wait to ensure goroutine runs

	// select {
	// case msg := <-globalCh:
	// 	println("Received from globalCh:", msg)
	// case msg := <-globalCh2:
	// 	println("Received from globalCh2:", msg)
	// default: // makes it non-blocking it never waits
	// 	println("No messages received")
	// }

	select {
	case res := <-globalCh:
		println(res)
	case <-time.After(2 * time.Second):
		println("timeout")
	}

	// if err := doSomething(); err != nil {
    // return fmt.Errorf("doSomething failed: %w", err)
	// }

}

// func fetchWithTimeout(url string) (string, error) {
// result := make(chan string)

// go func() {
// 	data := fetch(url)  // Slow network call
// 	result <- data
// }()

// select {
// case data := <-result:
// 	return data, nil
// case <-time.After(5 * time.Second):
// 	return "", errors.New("request timeout")
// }
// }
