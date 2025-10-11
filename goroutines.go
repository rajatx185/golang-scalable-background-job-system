package main

import (
	"fmt"
	"sync"
	"time"
	// "runtime"
)

// func worker(id int, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	fmt.Printf("Worker %d starting\n", id)
// 	time.Sleep(time.Second)
// 	fmt.Printf("Worker %d done\n", id)
// 	// fmt.Printf("Alive goroutines %v\n", runtime.NumGoroutine())
// }

// func main() {
// 	var wg sync.WaitGroup
// 	for i := 1; i <= 10; i++ {
// 		wg.Add(1)         // <-- Add BEFORE launching goroutine
// 		go worker(i, &wg) // // pass pointer to WaitGroup
// 	}

// 	fmt.Println("All workers launched")
// 	time.Sleep(2 * time.Second)
// 	wg.Wait() // wait for all goroutines to finish
// 	fmt.Println("Main Exiting")
// }
