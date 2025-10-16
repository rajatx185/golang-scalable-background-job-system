# Go Context (context.Context) - Complete Guide

Let me explain **Context** in Go - one of the most important patterns for managing goroutines, cancellations, and timeouts.

## What is Context?

**Context = A way to pass cancellation signals, deadlines, and request-scoped values across API boundaries and goroutines**

Think of it as a "control signal" that travels with your request through your application.

## Simple Analogy: Group Trip 🚌

```
You're organizing a bus tour with 20 people

Context = Walkie-talkie that everyone carries

Uses:
1. Cancellation: "Trip cancelled! Everyone return to bus!"
2. Timeout: "We leave in 30 minutes, be back by then"
3. Values: "Your group number is 5, your guide is Sarah"

Without context:
- No way to tell everyone to stop
- People wander off indefinitely
- Can't coordinate group activities
```

## The Context Interface

```go
type Context interface {
    // Deadline returns the time when work should be cancelled
    Deadline() (deadline time.Time, ok bool)
    
    // Done returns a channel that's closed when work should be cancelled
    Done() <-chan struct{}
    
    // Err returns why the context was cancelled
    Err() error
    
    // Value returns the value associated with key
    Value(key interface{}) interface{}
}
```

## Creating Contexts

### 1. **Background Context** (Root)

```go
package main

import (
    "context"
    "fmt"
)

func main() {
    // Create root context (never cancelled)
    ctx := context.Background()
    
    // Use in main or top-level requests
    processRequest(ctx)
}

func processRequest(ctx context.Context) {
    fmt.Println("Processing...")
}
```

**When to use:**
- `main()` function
- Top-level request handlers
- Tests
- Initial context for goroutines

### 2. **TODO Context** (Placeholder)

```go
func main() {
    // Use when you don't know which context to use yet
    ctx := context.TODO()
    
    // Placeholder during development
    doSomething(ctx)
}
```

**When to use:**
- During development when you're unsure
- Refactoring code to add context support

## Context Cancellation

### Manual Cancellation with `WithCancel`

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    
    // Start worker goroutine
    go worker(ctx)
    
    // Let it work for 2 seconds
    time.Sleep(2 * time.Second)
    
    // Cancel the context
    fmt.Println("Cancelling...")
    cancel()  // This signals worker to stop!
    
    // Wait a bit to see cancellation effect
    time.Sleep(1 * time.Second)
    fmt.Println("Done")
}

func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():  // Context cancelled!
            fmt.Println("Worker: Received cancellation signal")
            fmt.Println("Worker: Reason:", ctx.Err())
            return  // Exit goroutine
        default:
            fmt.Println("Worker: Working...")
            time.Sleep(500 * time.Millisecond)
        }
    }
}
```

**Output:**
```
Worker: Working...
Worker: Working...
Worker: Working...
Worker: Working...
Cancelling...
Worker: Received cancellation signal
Worker: Reason: context canceled
Done
```

### Visual Flow:

```
Time →
────────────────────────────────────────

Main:     Create ctx → Start worker → Sleep → cancel() → Sleep → Exit
                       ↓                       ↓
Worker:               Work → Work → Work → CANCELLED → Exit
                                            ↑
                      ctx.Done() channel closes here!
```

## Timeout Context

### `WithTimeout` - Automatic Cancellation

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Create context with 3-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()  // Always call cancel to free resources!
    
    // Start long-running task
    result := make(chan string)
    go longTask(ctx, result)
    
    // Wait for result or timeout
    select {
    case res := <-result:
        fmt.Println("Got result:", res)
    case <-ctx.Done():
        fmt.Println("Timeout! Reason:", ctx.Err())
    }
}

func longTask(ctx context.Context, result chan string) {
    // Simulate work that takes 5 seconds
    select {
    case <-time.After(5 * time.Second):
        result <- "Task completed"
    case <-ctx.Done():
        fmt.Println("Task cancelled:", ctx.Err())
        return
    }
}
```

**Output:**
```
Task cancelled: context deadline exceeded
Timeout! Reason: context deadline exceeded
```

### `WithDeadline` - Cancel at Specific Time

```go
func main() {
    // Cancel at specific time
    deadline := time.Now().Add(2 * time.Second)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()
    
    go work(ctx)
    
    time.Sleep(3 * time.Second)
}

func work(ctx context.Context) {
    select {
    case <-ctx.Done():
        fmt.Println("Deadline exceeded!")
        return
    }
}
```

## Real-World Example: HTTP Request with Timeout

```go
package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
)

func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // Create HTTP request with context
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }
    
    // Execute request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    return string(body), nil
}

func main() {
    // Try to fetch with 5-second timeout
    result, err := fetchWithTimeout("https://httpbin.org/delay/10", 5*time.Second)
    if err != nil {
        fmt.Println("Error:", err)  // Will timeout!
        return
    }
    
    fmt.Println("Result:", result)
}
```

**What happens:**
1. Request starts
2. Server takes 10 seconds to respond
3. After 5 seconds, context times out
4. HTTP request is cancelled
5. Error returned: `context deadline exceeded`

## Context Values (Request-Scoped Data)

### Passing Values Through Context

```go
package main

import (
    "context"
    "fmt"
)

// Define custom type for context keys (best practice)
type contextKey string

const (
    userIDKey    contextKey = "userID"
    requestIDKey contextKey = "requestID"
)

func main() {
    // Create context with values
    ctx := context.Background()
    ctx = context.WithValue(ctx, userIDKey, "user123")
    ctx = context.WithValue(ctx, requestIDKey, "req-456")
    
    // Pass to functions
    handleRequest(ctx)
}

func handleRequest(ctx context.Context) {
    // Retrieve values
    userID := ctx.Value(userIDKey).(string)
    requestID := ctx.Value(requestIDKey).(string)
    
    fmt.Printf("Handling request %s for user %s\n", requestID, userID)
    
    // Pass to deeper functions
    processData(ctx)
}

func processData(ctx context.Context) {
    // Can still access values
    userID := ctx.Value(userIDKey)
    fmt.Println("Processing data for user:", userID)
}
```

**Output:**
```
Handling request req-456 for user user123
Processing data for user: user123
```

### ⚠️ Important: Context Values Best Practices

```go
// ❌ BAD: Don't use for essential function parameters
func processUser(ctx context.Context) {
    userID := ctx.Value("userID").(string)  // Unclear, error-prone
    // ...
}

// ✓ GOOD: Use explicit parameters for essential data
func processUser(ctx context.Context, userID string) {
    // Clear and type-safe
    // ...
}

// ✓ GOOD: Context values for request-scoped metadata
// - Request IDs
// - Authentication tokens
// - Trace IDs
// - Request deadlines
```

## Multiple Goroutines with Context

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    // Start multiple workers
    for i := 1; i <= 5; i++ {
        go worker(ctx, i)
    }
    
    // Wait for context to timeout
    <-ctx.Done()
    fmt.Println("Main: All workers should stop now")
    
    // Give workers time to cleanup
    time.Sleep(500 * time.Millisecond)
}

func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d: Stopping (%v)\n", id, ctx.Err())
            return
        default:
            fmt.Printf("Worker %d: Working...\n", id)
            time.Sleep(500 * time.Millisecond)
        }
    }
}
```

**Output:**
```
Worker 1: Working...
Worker 2: Working...
Worker 3: Working...
Worker 4: Working...
Worker 5: Working...
Worker 1: Working...
Worker 2: Working...
...
Worker 1: Stopping (context deadline exceeded)
Worker 3: Stopping (context deadline exceeded)
Worker 5: Stopping (context deadline exceeded)
Worker 2: Stopping (context deadline exceeded)
Worker 4: Stopping (context deadline exceeded)
Main: All workers should stop now
```

## Context Tree (Parent-Child Relationship)

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // Root context
    root := context.Background()
    
    // Create child context with timeout
    parent, cancel1 := context.WithTimeout(root, 5*time.Second)
    defer cancel1()
    
    // Create grandchild context with shorter timeout
    child, cancel2 := context.WithTimeout(parent, 2*time.Second)
    defer cancel2()
    
    // Child will timeout first (2 seconds)
    go task(child, "Child")
    
    // Parent will timeout later (5 seconds)
    go task(parent, "Parent")
    
    time.Sleep(6 * time.Second)
}

func task(ctx context.Context, name string) {
    <-ctx.Done()
    fmt.Printf("%s context cancelled: %v\n", name, ctx.Err())
}
```

**Output:**
```
Child context cancelled: context deadline exceeded   (at 2s)
Parent context cancelled: context deadline exceeded  (at 5s)
```

### Visual Tree:

```
context.Background() (never cancelled)
    ↓
WithTimeout(5s) ─────────────────────→ Cancelled at 5s
    ↓
WithTimeout(2s) ──────────→ Cancelled at 2s
    
Rule: Child contexts inherit parent cancellation
      (but can have shorter deadlines)
```

## Real-World Example: Database Query with Timeout

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

type User struct {
    ID   int
    Name string
}

func getUserByID(ctx context.Context, db *sql.DB, userID int) (*User, error) {
    // Create query context with timeout
    queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    var user User
    
    // Execute query with context
    err := db.QueryRowContext(queryCtx, 
        "SELECT id, name FROM users WHERE id = ?", userID).
        Scan(&user.ID, &user.Name)
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func main() {
    // db, _ := sql.Open("mysql", "connection_string")
    // defer db.Close()
    
    // Request-level context with 5-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Query inherits timeout (but adds its own 2s limit)
    // user, err := getUserByID(ctx, db, 123)
    // if err != nil {
    //     fmt.Println("Error:", err)
    //     return
    // }
    
    // fmt.Printf("User: %+v\n", user)
}
```

## HTTP Server with Context

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Get request context (automatically created by http.Server)
    ctx := r.Context()
    
    // Add request ID to context
    ctx = context.WithValue(ctx, "requestID", generateRequestID())
    
    // Simulate long operation
    result := make(chan string, 1)
    go processRequest(ctx, result)
    
    // Wait for result or client disconnect
    select {
    case res := <-result:
        fmt.Fprintf(w, "Result: %s\n", res)
    case <-ctx.Done():
        // Client disconnected or request cancelled
        fmt.Println("Request cancelled:", ctx.Err())
        http.Error(w, "Request cancelled", http.StatusRequestTimeout)
    }
}

func processRequest(ctx context.Context, result chan string) {
    requestID := ctx.Value("requestID").(string)
    
    // Simulate work
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            fmt.Printf("Request %s: Stopping early\n", requestID)
            return
        case <-time.After(500 * time.Millisecond):
            fmt.Printf("Request %s: Working step %d\n", requestID, i+1)
        }
    }
    
    result <- "Done!"
}

func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

**What happens:**
- Client makes request
- Server starts processing
- If client disconnects, `ctx.Done()` closes
- Server stops processing immediately
- Saves resources!

## Context Patterns

### Pattern 1: Propagating Cancellation

```go
func orchestrator(ctx context.Context) {
    // Start multiple tasks
    go task1(ctx)
    go task2(ctx)
    go task3(ctx)
    
    // When ctx is cancelled, all tasks stop
}
```

### Pattern 2: Fan-Out with Timeout

```go
func processItems(ctx context.Context, items []Item) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    errCh := make(chan error, len(items))
    
    for _, item := range items {
        go func(it Item) {
            errCh <- processItem(ctx, it)
        }(item)
    }
    
    for range items {
        if err := <-errCh; err != nil {
            cancel()  // Cancel remaining work
            return err
        }
    }
    
    return nil
}
```

### Pattern 3: Graceful Shutdown

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Handle OS signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigCh
        fmt.Println("Shutting down...")
        cancel()  // Signal all goroutines to stop
    }()
    
    // Start workers with context
    go worker(ctx)
    
    // Wait for shutdown
    <-ctx.Done()
    fmt.Println("Cleanup complete")
}
```

## Common Mistakes

### ❌ Mistake 1: Not calling cancel()

```go
// Memory leak! Context resources not freed
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// forgot defer cancel()
```

### ❌ Mistake 2: Storing context in struct

```go
// DON'T do this
type Server struct {
    ctx context.Context  // ❌ Bad!
}

// DO this instead
type Server struct {
    // No context field
}

func (s *Server) Process(ctx context.Context) {  // ✓ Pass as parameter
    // ...
}
```

### ❌ Mistake 3: Using nil context

```go
// ❌ Don't pass nil
doWork(nil)

// ✓ Use context.Background() or context.TODO()
doWork(context.Background())
```

## Summary

### Context Use Cases:

```
✓ Cancellation signals across goroutines
✓ Request timeouts
✓ Graceful shutdown
✓ Request-scoped values (trace IDs, auth tokens)
✓ Preventing goroutine leaks
✓ Database query timeouts
✓ HTTP request cancellation
```

### Key Rules:

```
1. Always pass context as first parameter
2. Never store context in structs
3. Always call cancel() (use defer)
4. Don't pass nil context
5. Use WithValue sparingly (request metadata only)
6. Child contexts inherit parent cancellation
7. Cancelling parent cancels all children
```

### Context Flow:

```
main() 
  └─ context.Background()
       └─ WithTimeout(5s)
            ├─ HTTP Handler
            │   └─ WithTimeout(2s)
            │       └─ Database Query
            ├─ Worker Goroutine 1
            ├─ Worker Goroutine 2
            └─ Worker Goroutine 3
            
When parent cancelled → all children stop!
```

**Memory trick**: 
- "Context = Control signal for cooperative cancellation"
- "Always first parameter, never in struct"
- "Cancel = Clean up, don't forget defer!"

Context is essential for writing robust, production-ready Go applications! 🎯


