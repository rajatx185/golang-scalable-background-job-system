package main

import (
    "fmt"
    "sync"
    "time"
)

// ResilientOnce - wrapper around sync.Once with panic recovery
type ResilientOnce struct {
    once      sync.Once
    initErr   error
    mu        sync.RWMutex
    succeeded bool
}

func (r *ResilientOnce) Do(f func() error) error {
    r.once.Do(func() {
        defer func() {
            if rec := recover(); rec != nil {
                r.mu.Lock()
                r.initErr = fmt.Errorf("panic during initialization: %v", rec)
                r.succeeded = false
                r.mu.Unlock()
            }
        }()

        err := f()
        r.mu.Lock()
        r.initErr = err
        r.succeeded = (err == nil)
        r.mu.Unlock()
    })

    r.mu.RLock()
    defer r.mu.RUnlock()
    
    if !r.succeeded {
        return fmt.Errorf("initialization failed: %w", r.initErr)
    }
    return nil
}

// Alternative: Retry Pattern
type RetryOnce struct {
    mu          sync.Mutex
    initialized bool
    maxRetries  int
}

func (r *RetryOnce) Do(f func() error) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.initialized {
        return nil
    }

    var lastErr error
    for attempt := 0; attempt < r.maxRetries; attempt++ {
        err := func() (err error) {
            defer func() {
                if rec := recover(); rec != nil {
                    err = fmt.Errorf("panic: %v", rec)
                }
            }()
            return f()
        }()

        if err == nil {
            r.initialized = true
            return nil
        }

        lastErr = err
        fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
        time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1))
    }

    return fmt.Errorf("all %d attempts failed: %w", r.maxRetries, lastErr)
}

// Demo
func main() {
    fmt.Println("=== Testing ResilientOnce ===")
    testResilientOnce()
    
    fmt.Println("\n=== Testing RetryOnce ===")
    testRetryOnce()
}

func testResilientOnce() {
    var resOnce ResilientOnce
    var wg sync.WaitGroup
    attempt := 0

    initialize := func() error {
        attempt++
        fmt.Printf("Initialize attempt %d\n", attempt)
        if attempt == 1 {
            panic("first attempt fails!")
        }
        return nil
    }

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            err := resOnce.Do(initialize)
            if err != nil {
                fmt.Printf("Goroutine %d: ERROR - %v\n", id, err)
            } else {
                fmt.Printf("Goroutine %d: SUCCESS\n", id)
            }
        }(i)
        time.Sleep(50 * time.Millisecond)
    }

    wg.Wait()
}

func testRetryOnce() {
    retryOnce := RetryOnce{maxRetries: 3}
    var wg sync.WaitGroup
    attempt := 0

    initialize := func() error {
        attempt++
        fmt.Printf("Initialize attempt %d\n", attempt)
        if attempt < 2 {
            return fmt.Errorf("attempt %d failed", attempt)
        }
        fmt.Println("✅ Initialization succeeded!")
        return nil
    }

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            err := retryOnce.Do(initialize)
            if err != nil {
                fmt.Printf("Goroutine %d: ERROR - %v\n", id, err)
            } else {
                fmt.Printf("Goroutine %d: SUCCESS\n", id)
            }
        }(i)
        time.Sleep(50 * time.Millisecond)
    }

    wg.Wait()
}