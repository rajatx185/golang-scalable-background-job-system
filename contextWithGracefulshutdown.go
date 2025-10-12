package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // Cleanup if main exits early
    
    // Setup signal handling
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    
    // Signal handler goroutine
    go func() {
        sig := <-sigCh  // Wait for signal
        fmt.Printf("\nReceived signal: %v\n", sig)
        fmt.Println("Initiating graceful shutdown...")
        cancel()  // Trigger shutdown
    }()
    
    // Start multiple workers
    fmt.Println("Starting workers... (Press Ctrl+C to stop)")
    
    for i := 1; i <= 3; i++ {
        go worker(ctx, i)
    }
    
    // Wait for shutdown signal
    <-ctx.Done()
    
    // Give workers time to cleanup
    fmt.Println("Waiting for workers to finish...")
    time.Sleep(2 * time.Second)
    
    fmt.Println("All workers stopped. Goodbye!")
}

func worker(ctx context.Context, id int) {
    fmt.Printf("Worker %d: Started\n", id)
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            // Shutdown signal received
            fmt.Printf("Worker %d: Shutting down...\n", id)
            
            // Cleanup work here
            fmt.Printf("Worker %d: Closing connections...\n", id)
            time.Sleep(500 * time.Millisecond)
            
            fmt.Printf("Worker %d: Cleanup complete\n", id)
            return
            
        case <-ticker.C:
            // Normal work
            fmt.Printf("Worker %d: Processing...\n", id)
        }
    }
}