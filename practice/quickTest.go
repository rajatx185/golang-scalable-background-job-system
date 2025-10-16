package main

import (
	// "time"
	"sync"
    "fmt"
)

// func bufferedChannelQuestion() {
//     ch := make(chan int, 1) // Buffered!
//     var data int

//     go func() { // Not immediately executed/scheduled
//         data = 42
//         ch <- 1  // Doesn't block
//     }()

//     // Is reading 'data' here safe? Why or why not?
//     // println(data) // This is NOT safe

//     <-ch
//     println(data) // This IS safe
// }

var (
    mu sync.Mutex
    data int
)

// func mutexRule() {
//     // Rule: For sync.Mutex/RWMutex:
//     // - Call n to Unlock() happens-before call n+1 to Lock()
//     var wg sync.WaitGroup
//     wg.Add(1)
//     // Thread 1
//     go func() {
//         defer wg.Done()
//         mu.Lock()
//         data = 1    // Write 1
//         mu.Unlock() // Unlock call n happens-before...
//     }()
    
//     wg.Wait() // Wait for thread 1 to finish
//     // Thread 2
//     mu.Lock()       // ...Lock call n+1
//     println(data)   // Guaranteed to see value written by thread 1
//     mu.Unlock()
// }

func onceRule() {
    var once sync.Once
    var initialized bool
    
    // Rule: once.Do(f) call happens-before any once.Do(f) returns
    
    initialize := func() {
        initialized = true
        // Complex initialization...
    }
    
    // Multiple goroutines
    for i := 0; i < 10; i++ {
        go func() {
            fmt.Printf("Goroutine %d\n", i)
            once.Do(initialize) // Returns only after f completes
            println(initialized) // Always true
        }()
    }
}

// func main() {
//     // bufferedChannelQuestion()
//     // mutexRule()
//     onceRule()
// }