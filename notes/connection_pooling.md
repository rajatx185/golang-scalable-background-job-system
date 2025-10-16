# Database Connection Pooling Explained

## Short Answer

**No, goroutines don't share a single connection.** They wait in line to acquire one of the 20 available connections. It's like a taxi stand with 20 taxis and 1000 people waiting.

## How Connection Pooling Works

````go
package main

import (
    "database/sql"
    "fmt"
    "runtime"
    "sync"
    "time"
)

func connectionPoolingExplained() {
    runtime.GOMAXPROCS(8)
    
    db, _ := sql.Open("postgres", "...")
    db.SetMaxOpenConns(20) // Max 20 CONCURRENT connections
    
    var wg sync.WaitGroup
    
    // Launch 1000 goroutines
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Step 1: Try to acquire a connection from pool
            // If all 20 are busy, THIS GOROUTINE WAITS (parks)
            // Thread continues running OTHER goroutines!
            
            rows, _ := db.Query("SELECT * FROM users WHERE id = $1", id)
            
            // Step 2: Got a connection! Execute query
            // Connection is EXCLUSIVELY used by this goroutine
            
            if rows != nil {
                rows.Close() // Step 3: Release connection back to pool
            }
            
            // Step 4: Another waiting goroutine can now acquire it
            
        }(i)
    }
    
    wg.Wait()
    
    /*
    Timeline:
    
    Time 0: 1000 goroutines created
    
    Time 1: First 20 goroutines acquire connections
    - G1 → Connection 1
    - G2 → Connection 2
    - ...
    - G20 → Connection 20
    - G21-G1000: Waiting in queue (parked, NOT blocking threads!)
    
    Time 2: G1 completes query, releases Connection 1
    - G21 acquires Connection 1
    - G22-G1000: Still waiting
    
    Time 3: G2 and G3 complete
    - G22 acquires Connection 2
    - G23 acquires Connection 3
    - G24-G1000: Still waiting
    
    ...continues until all 1000 goroutines process
    */
}
````

## Visual Representation

```
Connection Pool (20 connections):
┌────────────────────────────────────────┐
│ [C1][C2][C3]...[C20]                   │ ← Only 20 connections to DB
└────────────────────────────────────────┘
     ↑    ↑    ↑       ↑
     │    │    │       │
   ┌─┴┐ ┌─┴┐ ┌─┴┐    ┌─┴┐
   │G1│ │G2│ │G3│... │G20│  ← Using connections
   └──┘ └──┘ └──┘    └───┘

Waiting Queue (980 goroutines):
┌────────────────────────────────────────┐
│ [G21][G22][G23]...[G1000]              │ ← Waiting for available connection
└────────────────────────────────────────┘
  (Goroutines parked, threads running others)

When G1 finishes:
- G1 releases C1 back to pool
- G21 wakes up and acquires C1
- G21 moves from waiting to active
```

## Key Points

````go
// Connection pooling is NOT the same as goroutine pooling!

// Goroutines: 1000 created
// Threads: ~8 (GOMAXPROCS)
// DB Connections: Max 20 concurrent

/*
Relationship:
- 1000 goroutines compete for 20 connections
- When goroutine can't get connection, it WAITS
- The THREAD doesn't wait, it runs other goroutines!
- Once connection is released, waiting goroutine resumes
*/
````

## What Happens When Goroutine Waits for Connection

````go
package main

import (
    "database/sql"
    "fmt"
    "time"
)

func whatHappensWhenWaiting() {
    db, _ := sql.Open("postgres", "...")
    db.SetMaxOpenConns(5) // Only 5 connections
    
    // Launch 10 goroutines
    for i := 0; i < 10; i++ {
        go func(id int) {
            fmt.Printf("Goroutine %d: Trying to acquire connection...\n", id)
            
            // Internally, db.Query() does:
            // 1. pool.getConnection() - tries to get connection
            // 2. If all busy, goroutine BLOCKS HERE (parks)
            // 3. Thread moves to another goroutine
            // 4. When connection available, goroutine wakes up
            
            start := time.Now()
            rows, _ := db.Query("SELECT pg_sleep(2)") // 2 second query
            
            fmt.Printf("Goroutine %d: Got connection after %v\n", id, time.Since(start))
            
            if rows != nil {
                rows.Close()
            }
            
        }(i)
    }
    
    time.Sleep(10 * time.Second)
    
    /*
    Output will show:
    - First 5 goroutines get connections immediately
    - Next 5 wait ~2 seconds (until first batch completes)
    - But THREADS don't block, they run other work!
    */
}
````

## Connection Pool Internals (Simplified)

````go
// Simplified version of how sql.DB works internally

type ConnectionPool struct {
    connections chan *Connection // Buffered channel of size MaxOpenConns
    maxOpen     int
}

func (p *ConnectionPool) GetConnection() *Connection {
    select {
    case conn := <-p.connections:
        // Got available connection
        return conn
        
    default:
        // No connection available
        // Goroutine BLOCKS here waiting on channel
        // Thread continues with other goroutines!
        conn := <-p.connections // Blocks until available
        return conn
    }
}

func (p *ConnectionPool) ReleaseConnection(conn *Connection) {
    // Put connection back in pool
    p.connections <- conn // Another goroutine can now acquire it
}

// Usage:
func query() {
    conn := pool.GetConnection()    // May wait here (goroutine parks)
    defer pool.ReleaseConnection(conn)
    
    // Use connection exclusively
    conn.Execute("SELECT * FROM users")
}
````

## Complete Example with Timing

````go
package main

import (
    "database/sql"
    "fmt"
    "runtime"
    "sync"
    "time"
)

func completePoolingExample() {
    runtime.GOMAXPROCS(4) // 4 threads
    
    db, _ := sql.Open("postgres", "...")
    db.SetMaxOpenConns(10)           // Max 10 concurrent connections
    db.SetMaxIdleConns(10)           // Keep 10 idle connections
    db.SetConnMaxLifetime(time.Hour) // Connection lifetime