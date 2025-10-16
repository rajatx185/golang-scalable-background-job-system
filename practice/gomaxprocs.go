package main

import (
    "fmt"
    "net/http"
    "runtime"
    "time"
)

func demonstrateThreadCreation() {
    runtime.GOMAXPROCS(4) // 4 P's (logical CPUs)
    
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
    fmt.Printf("Initial threads: %d\n", runtime.NumGoroutine())

    // Create goroutines that do BLOCKING I/O
    for i := 0; i < 100; i++ {
        go func(id int) {
            // Blocking system call (file I/O, network, etc.)
            resp, _ := http.Get("https://example.com")
            if resp != nil {
                resp.Body.Close()
            }
            
            // While THIS goroutine waits for network response...
            // Its M (thread) is BLOCKED
            // Go creates a NEW thread to keep other goroutines running!
        }(i)
    }
    
    time.Sleep(1 * time.Second)
    
    // You'll see: Much more than 4 threads created!
    fmt.Printf("After I/O operations, threads in use: Many more than GOMAXPROCS\n")
}

// func main() {
// 	demonstrateThreadCreation()
// }