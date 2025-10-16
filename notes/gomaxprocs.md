# GOMAXPROCS and Go Scheduler Deep Dive

## 1. Concurrency vs Parallelism

### **Concurrency**
- Multiple tasks making progress by switching between them
- About **dealing with lots of things at once**
- Can run on a single core via time-slicing
- Example: One chef handling multiple dishes by switching between them

### **Parallelism**
- Multiple tasks executing simultaneously
- About **doing lots of things at once**
- Requires multiple cores
- Example: Multiple chefs each cooking different dishes

```go
// Concurrency example - tasks can interleave
func concurrent() {
    go task1() // Goroutine 1
    go task2() // Goroutine 2
    // Both may run on same core, switching back and forth
}

// Parallelism - tasks run simultaneously on different cores
func parallel() {
    runtime.GOMAXPROCS(4) // Use 4 cores
    go cpuIntensiveTask1() // Runs on core 1
    go cpuIntensiveTask2() // Runs on core 2
    go cpuIntensiveTask3() // Runs on core 3
    go cpuIntensiveTask4() // Runs on core 4
}
```

## 2. Go Scheduler: GMP Model

### **Components**

- **G (Goroutine)**: Lightweight thread (~2KB stack)
- **M (Machine)**: OS thread
- **P (Processor)**: Scheduling context (logical CPU)

```
┌─────────────────────────────────────┐
│  GOMAXPROCS = Number of Ps          │
│  (Default: runtime.NumCPU())        │
└─────────────────────────────────────┘
         │
         ▼
    ┌────────┐  ┌────────┐  ┌────────┐
    │   P1   │  │   P2   │  │   P3   │
    │ Local  │  │ Local  │  │ Local  │
    │ Queue  │  │ Queue  │  │ Queue  │
    └────┬───┘  └────┬───┘  └────┬───┘
         │           │           │
         ▼           ▼           ▼
    ┌────────┐  ┌────────┐  ┌────────┐
    │   M1   │  │   M2   │  │   M3   │
    │(Thread)│  │(Thread)│  │(Thread)│
    └────────┘  └────────┘  └────────┘
```

### **How It Works**

1. Each P has a local run queue of goroutines
2. Each M (OS thread) must have a P to execute goroutines
3. Work stealing: Idle P steals from other P's queues
4. Global run queue for overflow

## 3. GOMAXPROCS Deep Dive

````go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func demonstrateGOMAXPROCS() {
    cores := runtime.NumCPU()
    fmt.Printf("Available CPU cores: %d\n", cores)
    fmt.Printf("Current GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

    // Test with different GOMAXPROCS values
    testParallelism(1, "Single Core")
    testParallelism(cores/2, "Half Cores")
    testParallelism(cores, "All Cores")
    testParallelism(cores*2, "2x Cores")
}

func testParallelism(maxprocs int, label string) {
    runtime.GOMAXPROCS(maxprocs)
    
    start := time.Now()
    var wg sync.WaitGroup
    
    // CPU-intensive work
    for i := 0; i < 8; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            sum := 0
            for j := 0; j < 1e8; j++ {
                sum += j
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("%s (GOMAXPROCS=%d): %v\n", label, maxprocs, time.Since(start))
}
````

## 4. Tuning for Multi-Core Throughput

### **CPU-Bound Workloads**

````go
package main

import (
    "runtime"
    "sync"
)

func optimizeCPUBound() {
    // Set to number of physical cores (not logical)
    // For hyper-threading: NumCPU()/2
    physicalCores := runtime.NumCPU() / 2
    runtime.GOMAXPROCS(physicalCores)
    
    numWorkers := runtime.GOMAXPROCS(0)
    jobs := make(chan int, 100)
    var wg sync.WaitGroup
    
    // Worker pool matching GOMAXPROCS
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                // CPU-intensive work
                processJob(job)
            }
        }()
    }
    
    // Send jobs
    for i := 0; i < 1000; i++ {
        jobs <- i
    }
    close(jobs)
    wg.Wait()
}

func processJob(id int) {
    // Simulate CPU work
    sum := 0
    for i := 0; i < 1e7; i++ {
        sum += i
    }
}
````

### **I/O-Bound Workloads**

````go
func optimizeIOBound() {
    // For I/O: More goroutines than cores is fine
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Can have 100s or 1000s of goroutines
    numWorkers := 100 // Much more than GOMAXPROCS
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // I/O operations (network, disk)
            // Scheduler will context-switch during blocking
        }()
    }
    wg.Wait()
}
````

## 5. Practical Tuning Guidelines

### **Default Setting (Recommended)**
```go
// Let Go decide - usually optimal
runtime.GOMAXPROCS(runtime.NumCPU())
```

### **CPU-Intensive Tasks**
```go
// Use physical cores only
physicalCores := runtime.NumCPU() / 2 // If hyper-threading
runtime.GOMAXPROCS(physicalCores)
```

### **Mixed Workloads**
```go
// Start with defaults, benchmark, and adjust
runtime.GOMAXPROCS(runtime.NumCPU())
```

### **Container/Cloud Environments**
```go
// Respect CPU quotas
import "github.com/uber-go/automaxprocs"

func init() {
    // Automatically sets GOMAXPROCS based on container limits
    automaxprocs.Set()
}
```

## 6. Monitoring & Benchmarking

````go
package main

import (
    "runtime"
    "testing"
)

func BenchmarkWithDifferentGOMAXPROCS(b *testing.B) {
    for _, procs := range []int{1, 2, 4, 8} {
        b.Run(fmt.Sprintf("GOMAXPROCS=%d", procs), func(b *testing.B) {
            runtime.GOMAXPROCS(procs)
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                // Your workload
            }
        })
    }
}

func monitorScheduler() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
    fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
}
````

## Key Takeaways

1. **GOMAXPROCS** = number of OS threads that can execute Go code simultaneously
2. **Default** (NumCPU) is optimal for most cases
3. **CPU-bound**: Consider using physical cores only
4. **I/O-bound**: Keep defaults, spawn many goroutines
5. **Always benchmark** before tuning
6. In containers, use tools like `automaxprocs` to respect CPU quotas

# Why Goroutines When One Thread = One Goroutine at a Time?

Great question! This is a critical distinction. Let me break down why goroutines are superior to threads.

## 1. The Key Insight

**Yes, one M (thread) executes one G (goroutine) at a time.**

But the magic is in:
- **How fast they switch**
- **How cheap they are**
- **How the scheduler manages them**

## 2. Goroutines vs Threads: The Real Differences

### **Memory Cost**

```go
// THREADS
// Each OS thread: ~1-2 MB stack
1000 threads = 1-2 GB memory minimum
10,000 threads = System crash 💥

// GOROUTINES
// Each goroutine: ~2 KB initial stack (grows dynamically)
1000 goroutines = ~2 MB memory
1,000,000 goroutines = ~2 GB memory ✅
```

````go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func demonstrateScalability() {
    // Try creating 100,000 goroutines
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < 100000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            time.Sleep(1 * time.Second)
        }(i)
    }
    
    fmt.Printf("Created 100,000 goroutines in %v\n", time.Since(start))
    fmt.Printf("Memory used: ~%d MB\n", runtime.NumGoroutine()*2/1024)
    
    wg.Wait()
    // Try this with OS threads - system will die! 💀
}
````

### **Context Switching Cost**

```
OS THREAD CONTEXT SWITCH (Kernel Space):
┌─────────────────────────────────────┐
│ 1. Save registers                   │
│ 2. Save stack pointer               │
│ 3. Switch to kernel mode            │ ~1-2 microseconds
│ 4. Update kernel scheduler          │
│ 5. Load new thread state            │
│ 6. Switch back to user mode         │
└─────────────────────────────────────┘

GOROUTINE CONTEXT SWITCH (User Space):
┌─────────────────────────────────────┐
│ 1. Save 3 registers (PC, SP, DX)    │ ~200 nanoseconds
│ 2. Switch goroutine pointer         │ (10x faster!)
└─────────────────────────────────────┘
```

## 3. The Real Power: Cooperative vs Preemptive Scheduling

### **OS Threads (Preemptive)**

```go
// OS decides when to switch - expensive!
Thread 1: Running... [INTERRUPT] [SAVE STATE] [KERNEL CALL]
Thread 2: Now running... [INTERRUPT] [SAVE STATE] [KERNEL CALL]
// Involves kernel, lots of overhead
```

### **Goroutines (Cooperative + Smart Preemption)**

````go
package main

import (
    "fmt"
    "time"
)

// Goroutines yield control at these points:
func goroutineYieldPoints() {
    // 1. Channel operations
    ch := make(chan int)
    go func() {
        val := <-ch // Yields here if empty
        fmt.Println(val)
    }()
    
    // 2. Network/IO operations
    go func() {
        time.Sleep(1 * time.Millisecond) // Yields during sleep
    }()
    
    // 3. Function calls (Go 1.14+)
    go func() {
        recursiveFunction() // Can preempt at function boundaries
    }()
    
    // 4. Explicit yield
    go func() {
        for {
            // Do work
            runtime.Gosched() // Explicit yield
        }
    }()
}

func recursiveFunction() {
    // Function calls allow preemption points
}
````

## 4. The Real-World Scenario: Why This Matters

### **Scenario: Web Server Handling 10,000 Concurrent Requests**

````go
// ❌ BAD: Using OS Threads
func handleWithThreads() {
    // Each request = 1 thread
    // 10,000 threads × 2MB = 20GB memory
    // Context switches kill performance
    // Most threads are WAITING (I/O bound)
    
    for request := range requests {
        go handleRequest(request) // If these were threads - 💀
    }
}

// ✅ GOOD: Using Goroutines
func handleWithGoroutines() {
    // 10,000 goroutines × 2KB = 20MB memory
    // Fast context switches
    // While goroutine waits for I/O, thread runs another goroutine!
    
    for request := range requests {
        go handleRequest(request) // Lightweight!
    }
}

func handleRequest(req Request) {
    // 1. Read from database (I/O - goroutine yields)
    data := db.Query() // Thread switches to another goroutine here!
    
    // 2. Call external API (I/O - goroutine yields)
    result := api.Call() // Thread switches again!
    
    // 3. Write response (I/O - goroutine yields)
    response.Write(result) // Thread can do other work!
}
````

## 5. The M:N Scheduler Magic

````go
/*
THREADS (1:1 model)
┌─────────────────────────────────────┐
│ 1000 tasks = 1000 OS threads        │
│ All managed by OS kernel            │
│ Heavy, expensive, limited           │
└─────────────────────────────────────┘

GOROUTINES (M:N model)
┌─────────────────────────────────────┐
│ 1,000,000 goroutines (G)            │
│      mapped to                      │
│ 8 OS threads (M)                    │
│      via                            │
│ 8 logical processors (P)            │
└─────────────────────────────────────┘
*/

package main

import (
    "fmt"
    "runtime"
    "time"
)

func demonstrateMNScheduling() {
    runtime.GOMAXPROCS(4) // 4 processors (P)
    // Go creates ~4 OS threads (M)
    
    // But we can create unlimited goroutines!
    for i := 0; i < 100000; i++ {
        go func(id int) {
            // This goroutine might run on ANY of the 4 threads
            // When it blocks (I/O), the thread picks up another goroutine
            time.Sleep(100 * time.Millisecond)
        }(i)
    }
    
    time.Sleep(1 * time.Second)
    fmt.Printf("100,000 goroutines running on %d threads\n", 
        runtime.GOMAXPROCS(0))
}
````

## 6. Work Stealing: Efficient CPU Utilization

````go
/*
WITHOUT WORK STEALING (Traditional Threads)
Thread 1: [████████████████] Busy
Thread 2: [░░░░░░░░░░░░░░░░] Idle (wasted CPU)
Thread 3: [████████████████] Busy
Thread 4: [░░░░░░░░░░░░░░░░] Idle (wasted CPU)

WITH WORK STEALING (Go Scheduler)
P1: [████] Local queue: G1, G2, G3
P2: [░░░░] Local queue: Empty → Steals G3 from P1!
P3: [████] Local queue: G4, G5, G6
P4: [████] Local queue: G7, G8, G9
*/

func demonstrateWorkStealing() {
    runtime.GOMAXPROCS(4)
    
    // Create uneven work distribution
    for i := 0; i < 1000; i++ {
        go func(id int) {
            if id < 100 {
                // Heavy work
                sum := 0
                for j := 0; j < 1e8; j++ {
                    sum += j
                }
            } else {
                // Light work
                time.Sleep(1 * time.Millisecond)
            }
        }(i)
    }
    
    // Go scheduler automatically balances load across P's
    // Idle processors steal goroutines from busy ones
}
````

## 7. Blocking System Calls: Goroutines Win Again

````go
package main

import (
    "fmt"
    "os"
    "runtime"
    "time"
)

// What happens when goroutine blocks on syscall?
func blockingSystemCall() {
    runtime.GOMAXPROCS(2) // 2 processors, 2 threads initially
    
    go func() {
        // This blocks the OS thread
        file, _ := os.Open("large_file.txt")
        defer file.Close()
        
        // While this goroutine blocks...
        // Go scheduler does something smart:
        // 1. Detaches the P (processor) from this M (thread)
        // 2. Assigns P to a NEW thread or existing spare thread
        // 3. That P continues running other goroutines!
        // 4. When syscall completes, goroutine goes back to run queue
    }()
    
    // These goroutines keep running on other threads!
    for i := 0; i < 100; i++ {
        go func(id int) {
            // CPU work continues uninterrupted
            sum := 0
            for j := 0; j < 1e7; j++ {
                sum += j
            }
        }(i)
    }
    
    time.Sleep(1 * time.Second)
    fmt.Printf("Despite blocking I/O, all goroutines made progress!\n")
}
````

## 8. Concrete Example: Why Goroutines Win

````go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Simulating web server handling concurrent requests

func compareThreadsVsGoroutines() {
    requests := 10000
    
    fmt.Println("=== If we used OS Threads ===")
    fmt.Printf("Memory: %d threads × 2MB = %dGB\n", requests, requests*2/1024)
    fmt.Println("Context switches: Kernel-level, ~1-2 microseconds each")
    fmt.Println("Result: System would likely crash or thrash")
    
    fmt.Println("\n=== With Goroutines ===")
    fmt.Printf("Memory: %d goroutines × 2KB = %dMB\n", requests, requests*2/1024)
    fmt.Println("Context switches: User-level, ~200 nanoseconds")
    
    start := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < requests; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Simulate I/O operations
            time.Sleep(10 * time.Millisecond) // Database query
            time.Sleep(20 * time.Millisecond) // API call
            time.Sleep(5 * time.Millisecond)  // Write response
            
            // While this goroutine sleeps, its thread runs OTHER goroutines!
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("Handled %d requests in %v\n", requests, time.Since(start))
    fmt.Printf("Running on only %d OS threads!\n", runtime.GOMAXPROCS(0))
}
````

## Key Takeaways

| Feature | OS Threads | Goroutines |
|---------|-----------|------------|
| **Stack Size** | 1-2 MB fixed | 2 KB (grows to 1GB) |
| **Creation Cost** | ~1 ms | ~1 µs |
| **Switch Cost** | ~1-2 µs (kernel) | ~200 ns (user space) |
| **Scalability** | ~1000s | Millions |
| **Blocking** | Blocks thread | Only blocks goroutine |
| **Scheduling** | OS kernel | Go runtime |

## The Bottom Line

**Goroutines allow you to write code as if you have unlimited threads, but run it efficiently on limited OS threads.**

```go
// You write this (thinking in terms of concurrent tasks)
for _, task := range millionTasks {
    go processTask(task)
}

// Go runtime efficiently maps this to:
// - A few OS threads (GOMAXPROCS)
// - Fast user-space scheduling
// - Automatic work stealing
// - Efficient I/O handling
```

**Without goroutines, you'd need to manually manage thread pools, queues, and complex scheduling logic. Goroutines abstract all of this away!**

# Queue Sizes and Overflow Handling

## Local Queue Size

- **P's Local Queue**: **256 goroutines** (fixed size)
- **Global Queue**: **Unlimited** (dynamically grows)

```go
// When you create a goroutine:
go handleRequest() // Goes to P's local queue (256 max)

// If local queue is full → Goes to global queue (unlimited)
```

## You're Right - We DON'T Store Threads in Queues!

**We store GOROUTINES (G) in queues, not threads (M)**

```
Queue stores: Goroutines (G) ✅
Queue does NOT store: Threads (M) ❌

┌─────────────────────────────────┐
│ P's Local Queue (256 max)       │
│ [G1][G2][G3]...[G256]           │
└─────────────────────────────────┘
           ↓ Overflow
┌─────────────────────────────────┐
│ Global Queue (unlimited)        │
│ [G257][G258][G259]...           │
└─────────────────────────────────┘
```

## What Happens When System Can't Create More Threads?

**Nothing breaks!** Go has a thread limit but it's high:

```go
// Max OS threads Go will create: ~10,000 (configurable)
runtime.SetMaxThreads(10000)

// But you can have MILLIONS of goroutines
// They just wait in queues to be scheduled
```

## How Web Applications Handle This

````go
package main

import (
    "net/http"
    "time"
)

// Pattern 1: Unlimited Goroutines (Default)
func unlimitedHandler(w http.ResponseWriter, r *http.Request) {
    // Each request = 1 goroutine
    // If 100,000 requests come: 100,000 goroutines created
    // They queue up, none are dropped
    // User waits (request times out after default timeout)
    
    time.Sleep(100 * time.Millisecond) // Simulate work
    w.Write([]byte("Done"))
}

// Pattern 2: Rate Limiting (Production Pattern)
func rateLimitedServer() {
    limiter := make(chan struct{}, 1000) // Max 1000 concurrent requests
    
    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        select {
        case limiter <- struct{}{}: // Try to acquire slot
            defer func() { <-limiter }() // Release slot
            
            // Process request
            handleRequest(w, r)
            
        default:
            // No slot available - reject immediately
            http.Error(w, "Server busy", http.StatusServiceUnavailable)
        }
    })
}

// Pattern 3: Worker Pool (Best for CPU-bound)
func workerPoolServer() {
    jobs := make(chan *http.Request, 10000) // Buffered queue
    
    // Fixed number of workers
    for i := 0; i < 100; i++ {
        go worker(jobs) // Only 100 goroutines
    }
    
    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        select {
        case jobs <- r: // Queue request
            // User waits for worker to pick it up
        case <-time.After(5 * time.Second):
            http.Error(w, "Timeout", http.StatusRequestTimeout)
        }
    })
}
````

## Real-World Behavior

| Scenario | What Happens | User Experience |
|----------|-------------|-----------------|
| **Normal Load** | Goroutines scheduled instantly | Fast response |
| **High Load** | Goroutines queue up | Slower response |
| **Extreme Load** | Queue grows large | Timeout/slow |
| **With Rate Limiting** | Requests rejected when limit hit | 503 error immediately |

## Key Points

1. **Goroutines queue up** - they're NOT dropped by default
2. **Users wait** - until timeout (usually 30-60s)
3. **Production apps use**:
   - Rate limiting
   - Worker pools
   - Load balancers
   - Circuit breakers

```go
// The system NEVER runs out of ability to create goroutines
// It only runs out of CPU/memory to execute them efficiently
// That's why you add rate limiting in production!
```

# Thread Count vs CPU Cores

## Short Answer

**Yes, Go CAN and DOES create more threads than CPU cores.** This is not only possible but often necessary!

## Why More Threads Than CPUs Makes Sense

````go
// Your intuition is right for CPU-bound work:
runtime.GOMAXPROCS(runtime.NumCPU()) // Usually 8-16

// But Go can create MANY more OS threads (M):
runtime.SetMaxThreads(10000) // Default: 10,000 threads

// Why? Because threads get BLOCKED on I/O!
````

## The Key Insight: Blocking I/O

```go
┌──────────────────────────────────────────────┐
│ CPU-Bound: Threads = CPUs makes sense       │
│ Thread never blocks, always computing       │
└──────────────────────────────────────────────┘

┌──────────────────────────────────────────────┐
│ I/O-Bound: Need MORE threads than CPUs      │
│ Thread blocks waiting for I/O response      │
│ CPU sits idle if no extra threads!          │
└──────────────────────────────────────────────┘
```

## What Happens When Thread Blocks

````go
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
````

## The Thread Lifecycle

```
SCENARIO: GOMAXPROCS = 4 (4 CPUs)

Initial State:
P1 → M1 (thread 1)
P2 → M2 (thread 2)
P3 → M3 (thread 3)
P4 → M4 (thread 4)
Total: 4 threads for 4 CPUs ✓

Goroutine on M1 makes blocking syscall (disk read):
❌ M1 is BLOCKED (waiting for disk)
✅ P1 detaches from M1
✅ Go creates M5 (new thread)
✅ P1 attaches to M5
✅ P1 continues running other goroutines!

Now we have:
P1 → M5 (thread 5) - Running
P2 → M2 (thread 2) - Running
P3 → M3 (thread 3) - Running
P4 → M4 (thread 4) - Running
M1 - Blocked on syscall (parked)
Total: 5 threads for 4 CPUs ✓

More blocking calls:
M2 blocks → Creates M6
M3 blocks → Creates M7
...potentially up to 10,000 threads!
```

## Ideal Thread Count

````go
package main

import "runtime"

// For PURE CPU-bound work:
func cpuBoundIdeal() {
    // Threads = CPUs is perfect
    runtime.GOMAXPROCS(runtime.NumCPU())
    // Go will create ~NumCPU threads
    // No need for more since nothing blocks
}

// For I/O-bound work:
func ioBoundIdeal() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    // But Go will create MANY more threads automatically!
    // As goroutines block on I/O, new threads are created
    
    // You don't set thread count directly
    // Go manages it automatically based on blocking patterns
}

// The formula:
/*
Active Threads = GOMAXPROCS + Blocked Threads

Example:
- GOMAXPROCS = 4
- 10 goroutines blocked on network I/O
- Active threads = 4 + 10 = 14 threads

Go creates threads on-demand when blocking happens!
*/
````

## Real-World Example

````go
package main

import (
    "database/sql"
    "fmt"
    "runtime"
    "time"
)

func webServerExample() {
    runtime.GOMAXPROCS(8) // 8-core machine
    
    // Handling 1000 concurrent requests
    for i := 0; i < 1000; i++ {
        go handleRequest(i)
    }
    
    time.Sleep(5 * time.Second)
}

func handleRequest(id int) {
    // 1. Database query (BLOCKS M/thread)
    db.Query("SELECT * FROM users") // Thread blocked for 50ms
    
    // 2. External API call (BLOCKS M/thread)
    http.Get("https://api.example.com") // Thread blocked for 100ms
    
    // 3. Redis call (BLOCKS M/thread)
    redis.Get("key") // Thread blocked for 5ms
    
    // During each block:
    // - This M (thread) is parked
    // - P (processor) detaches and finds another M
    // - If no M available, Go creates a new one
    // - Other goroutines keep running!
}

/*
With 1000 concurrent requests:
- GOMAXPROCS = 8 (8 P's)
- Each request blocks 3 times
- Potentially 100+ threads created
- But only 8 are ACTIVE at any moment
- Rest are BLOCKED waiting for I/O
*/
````

## Monitoring Threads

````go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func monitorThreads() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
        fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
        
        // Note: No direct way to get thread count in Go
        // But you can infer: More blocking = More threads
    }
}
````

## Key Takeaways

| Aspect | Reality |
|--------|---------|
| **Can create more threads than CPUs?** | ✅ Yes, absolutely |
| **Is it useful?** | ✅ Yes, for I/O-bound work |
| **Who manages it?** | Go runtime (automatic) |
| **Do you set it?** | No, it happens automatically |
| **What you set** | GOMAXPROCS (parallel execution) |
| **What Go manages** | Thread creation (based on blocking) |

## The Formula

```go
Total Threads = GOMAXPROCS + Blocked Threads

// Example 1: Pure CPU work
// GOMAXPROCS = 8
// Blocked threads = 0 (nothing blocks)
// Total threads ≈ 8

// Example 2: Web server (I/O heavy)
// GOMAXPROCS = 8
// Blocked threads = 50 (waiting on I/O)
// Total threads ≈ 58

// Go creates threads on-demand!
```

## Bottom Line

```go
// You configure parallelism (GOMAXPROCS):
runtime.GOMAXPROCS(runtime.NumCPU()) // Usually 8-16

// Go manages concurrency (thread creation):
// - Creates threads when goroutines block
// - Reuses threads when they unblock
// - Can create up to 10,000 threads
// - You don't control this directly

// This is the MAGIC of Go's scheduler!
```

**Your intuition is correct for CPU-bound work, but I/O-bound work NEEDS more threads than CPUs because threads spend time blocked waiting!**


I think my understanding as to how context switch happens between go routines is not fully complete. I thought if a go routine is blocked on an IO the OS thread will switch it with a new go routine but you are saying that the thread itself will be switched then in IO bound cases we will see similar performance as thread only system, if we were always creating new threads.a

# Context Switching: The Critical Distinction

You've hit on a **very important nuance**! Let me clarify the difference between **network I/O** (non-blocking) and **blocking syscalls** (blocking).

## Two Types of I/O in Go

### 1. **Network I/O (Non-Blocking) - Most Common**

````go
// This does NOT block the OS thread!
resp, err := http.Get("https://api.com") // Network I/O

// What actually happens:
/*
1. Goroutine initiates network request
2. Go runtime parks the GOROUTINE (not the thread!)
3. Thread (M) picks up ANOTHER goroutine from queue
4. Thread stays active, running other goroutines
5. When network response arrives (via epoll/kqueue)
6. Original goroutine becomes runnable again
*/
````

### 2. **Blocking Syscalls (Rare) - Creates New Thread**

````go
// This DOES block the OS thread
file, _ := os.Open("file.txt")     // File I/O (syscall)
file.Read(buffer)                   // Blocking read

// What happens:
/*
1. Goroutine makes blocking syscall
2. OS thread (M) blocks waiting for kernel
3. Go runtime detaches P from this M
4. P attaches to NEW/existing thread
5. Other goroutines continue on new thread
*/
````

## The Magic: Network I/O is Non-Blocking!

````go
package main

import (
    "fmt"
    "net/http"
    "runtime"
    "time"
)

func demonstrateNonBlockingIO() {
    runtime.GOMAXPROCS(2) // Only 2 threads initially
    
    fmt.Printf("Starting with %d threads\n", 2)
    
    // Launch 1000 HTTP requests
    for i := 0; i < 1000; i++ {
        go func(id int) {
            // Network I/O - Does NOT block thread!
            resp, _ := http.Get("https://httpbin.org/delay/1")
            if resp != nil {
                resp.Body.Close()
            }
            
            // While waiting for response:
            // - This GOROUTINE is parked
            // - The THREAD continues running other goroutines!
            // - NO new thread is created!
        }(i)
    }
    
    time.Sleep(2 * time.Second)
    // Still only ~2 threads! All 1000 requests handled on 2 threads!
}
````

## How Go Achieves This: The Netpoller

```go
/*
Go's Network Poller (epoll on Linux, kqueue on macOS)
┌──────────────────────────────────────────┐
│  Thread 1 (M1 + P1)                      │
│  ┌────────────────────────────────┐      │
│  │ Running G1 (computing)         │      │
│  └────────────────────────────────┘      │
│                                           │
│  G2: Waiting for network (parked)        │
│  G3: Waiting for network (parked)        │
│  G4: Waiting for network (parked)        │
│  ...                                     │
│  G1000: Waiting for network (parked)     │
└──────────────────────────────────────────┘

When network data arrives:
1. OS signals via epoll/kqueue
2. Netpoller marks goroutine as RUNNABLE
3. Scheduler picks it up
4. All on SAME thread! No thread creation!
*/
```

## Visual Comparison

### Network I/O (90% of web apps)

```go
Timeline of 1 Thread handling 3 HTTP requests:

Thread M1:
[G1: Send request]─[G2: Send request]─[G3: Send request]─[G1: Process response]─[G2: Process response]

While G1 waits → Thread runs G2
While G2 waits → Thread runs G3  
While G3 waits → Thread runs G1 (response arrived)

Result: 1 thread, 3 concurrent operations ✅
No thread creation! Just goroutine switching!
```

### Blocking Syscalls (Rare)

```go
Timeline of blocking file I/O:

Thread M1:
[G1: os.Open()]──────[BLOCKED]──────→ (waiting on kernel)

Thread M2 (newly created):
[G2: Running]─[G3: Running]─[G4: Running]

When G1's syscall completes:
M1 returns to pool, G1 becomes runnable

Result: 2 threads needed ❌
Thread creation overhead!
```

## The Complete Picture

````go
package main

import (
    "fmt"
    "net/http"
    "os"
    "runtime"
    "time"
)

func completeExample() {
    runtime.GOMAXPROCS(4) // 4 threads initially
    
    // Scenario 1: Network I/O (non-blocking)
    for i := 0; i < 1000; i++ {
        go func() {
            // No new threads created!
            // Goroutine parks, thread continues with others
            http.Get("https://api.com")
        }()
    }
    
    // Scenario 2: Blocking syscalls
    for i := 0; i < 10; i++ {
        go func() {
            // New threads created! (up to 10 more)
            // Because thread blocks on kernel call
            file, _ := os.Open("large_file.txt")
            if file != nil {
                file.Close()
            }
        }()
    }
    
    time.Sleep(2 * time.Second)
    
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
    fmt.Printf("Total goroutines: %d\n", runtime.NumGoroutine())
    // Threads: ~4 (for network) + ~10 (for file I/O) = ~14
}
````

## Why Go is Still Better Than Thread-per-Request

````go
// Traditional thread-per-request (like Java without NIO):
func traditionalThreadModel() {
    // 1000 requests = 1000 threads
    // Each thread: 2MB stack
    // Total: 2GB memory
    // Context switches: Kernel-level (expensive)
}

// Go with network I/O:
func goNetworkModel() {
    // 1000 requests = 1000 goroutines
    // Running on: ~GOMAXPROCS threads (e.g., 8)
    // Each goroutine: 2KB stack
    // Total: 2MB memory
    // Context switches: User-level (cheap)
    
    // Thread doesn't block, just switches goroutines!
}

// Go with blocking syscalls:
func goBlockingModel() {
    // If you have 100 blocking syscalls happening:
    // 100 threads created temporarily
    // But still better than 1000 threads!
    // And threads are reused after syscall completes
}
````

## Practical Example: Database Connection Pool

````go
package main

import (
    "database/sql"
    "fmt"
    "runtime"
)

func databaseExample() {
    runtime.GOMAXPROCS(8)
    
    // Database with connection pool
    db, _ := sql.Open("postgres", "...")
    db.SetMaxOpenConns(20) // Only 20 connections
    
    // Handle 1000 concurrent requests
    for i := 0; i < 1000; i++ {
        go func(id int) {
            // db.Query() uses network I/O internally
            // Does NOT block OS thread!
            // Goroutine waits, thread serves others
            rows, _ := db.Query("SELECT * FROM users")
            if rows != nil {
                rows.Close()
            }
        }(i)
    }
    
    /*
    Result:
    - 1000 goroutines
    - Running on ~8 threads (GOMAXPROCS)
    - 20 active database connections
    - No thread blocking!
    - Goroutines queue at connection pool, not threads
    */
}
````

## Key Insight: Modern I/O is Non-Blocking

````go
// These are NON-BLOCKING (don't create threads):
http.Get()              // Network I/O
net.Dial()              // TCP connections
db.Query()              // Database (uses network internally)
redis.Get()             // Redis (network)
time.Sleep()            // Timer
<-channel               // Channel operations

// These ARE BLOCKING (may create threads):
os.Open()               // File operations
os.Read()               // File read
syscall.Read()          // Direct syscalls
cgo calls               // C function calls
````

## The Answer to Your Question

> "If we were always creating new threads, wouldn't performance be similar to thread-only systems?"

**No, because:**

1. **90% of web app I/O is network I/O** → Non-blocking, no thread creation
2. **Goroutine switching is 10x faster** than thread switching (200ns vs 2µs)
3. **Threads are reused** after blocking calls complete
4. **You only create threads when absolutely necessary** (actual blocking syscalls)

````go
// Performance comparison:
Traditional: 1000 requests = 1000 threads created
Go Network:  1000 requests = 0 new threads created (reuse existing)
Go Blocking: 100 file ops = 100 threads created temporarily

// Go is hybrid:
// - Network I/O: Pure goroutine switching (fast!)
// - Blocking I/O: Thread creation (only when needed)
````

## Final Mental Model

```
Your corrected understanding:

Network I/O (http.Get, db.Query):
┌─────────────────────────────────┐
│ Goroutine parks                 │
│ Thread continues with others    │ ← THIS is the common case!
│ No thread creation              │
└─────────────────────────────────┘

Blocking Syscall (os.Open):
┌─────────────────────────────────┐
│ Thread blocks                   │
│ New thread created              │ ← Rare, only for syscalls
│ P switches to new thread        │
└─────────────────────────────────┘
```

**Bottom line: Go's scheduler is smart enough to avoid thread creation for the most common case (network I/O), which is why it's so efficient for web applications!**
