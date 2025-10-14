# Race Detector in Go

The **race detector** is a powerful tool built into Go that helps identify **data races** in concurrent programs during runtime.

## What is a Data Race?

A data race occurs when:
- Two or more goroutines access the same variable concurrently
- At least one of the accesses is a write
- The accesses are not synchronized

## Using the Race Detector

### Basic Usage

```bash
# Run your program with race detection
go run -race main.go

# Test with race detection
go test -race

# Build with race detection
go build -race
```

## Example: Detecting a Race

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    counter := 0
    
    // Start multiple goroutines that modify counter
    for i := 0; i < 5; i++ {
        go func() {
            counter++ // Race condition!
        }()
    }
    
    time.Sleep(time.Second)
    fmt.Println("Counter:", counter)
}
```

**Running with race detector:**
```bash
go run -race main.go
```

**Output will show:**
```
WARNING: DATA RACE
Write at 0x... by goroutine X:
Read at 0x... by goroutine Y:
```

## Fixing the Race

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    counter := 0
    var mu sync.Mutex
    
    for i := 0; i < 5; i++ {
        go func() {
            mu.Lock()
            counter++ // Now safe!
            mu.Unlock()
        }()
    }
    
    time.Sleep(time.Second)
    mu.Lock()
    fmt.Println("Counter:", counter)
    mu.Unlock()
}
```

## Important Notes

- **Performance Impact**: Race detector slows execution (5-10x) and increases memory usage
- **Runtime Detection Only**: Only catches races that actually occur during execution
- **Not for Production**: Only use during development and testing
- **Coverage Matters**: Write comprehensive tests to trigger all code paths

## Best Practices

1. **Always run tests with `-race`** in CI/CD pipelines
2. Use it during development when writing concurrent code
3. Combine with good test coverage
4. Fix all races before production deployment

The race detector is essential for building reliable concurrent Go applications! 🏁