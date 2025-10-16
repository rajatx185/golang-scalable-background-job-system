# What "Visible via Cache Coherency" Means

Let me break down exactly what happens when writes become visible to other cores through cache coherency.

## The Core Concept

**"Visible via cache coherency"** means: Other CPU cores can now observe the updated value through the cache coherency protocol, without necessarily going to main memory.

## How Cache Coherency Works

### The MESI Protocol (Most Common)

Each cache line in every core has a state:

```
M = Modified   - This core has the only valid copy, it's been modified
E = Exclusive  - This core has the only copy, it's clean (matches memory)
S = Shared     - Multiple cores have this cache line, all clean
I = Invalid    - This cache line is stale/invalid
```

### Step-by-Step Example

````go
// Initial state - both cores have x = 0 cached
var x int64 = 0

Core 1 Cache:  x = 0 [Shared]
Core 2 Cache:  x = 0 [Shared]
Memory:        x = 0
````

#### Step 1: Core 1 Writes x = 42

```
Core 1 executes: x = 42

What happens:
1. Core 1 wants to write to x
2. Core 1 checks cache line state: [Shared]
3. Core 1 must gain exclusive ownership

Core 1 sends: "Invalidate Request" on bus
              ↓
       ┌──────┴──────┐
       │ System Bus  │ (interconnect between cores)
       └──────┬──────┘
              ↓
Core 2 receives: "Invalidate x"
Core 2 marks: x [Invalid]

Result:
Core 1 Cache:  x = 42 [Modified]  ← Only valid copy
Core 2 Cache:  x = ?? [Invalid]   ← Knows it's stale
Memory:        x = 0               ← Not updated yet!
```

**Key Point**: Core 2 **knows** its cached value is invalid, but hasn't fetched the new value yet.

#### Step 2: Core 2 Reads x

```
Core 2 executes: read x

What happens:
1. Core 2 checks cache: x [Invalid] - can't use this!
2. Core 2 sends "Read Request" on bus
              ↓
       ┌──────┴──────┐
       │ System Bus  │
       └──────┬──────┘
              ↓
Core 1 receives: "Read Request for x"
Core 1 responds: Sends x = 42 on bus
                 Changes state to [Shared]

Result:
Core 1 Cache:  x = 42 [Shared]    ← Still has it
Core 2 Cache:  x = 42 [Shared]    ← Got new value
Memory:        x = 0 or 42        ← May or may not be updated
```

**This is "visible via cache coherency"**: Core 2 got the updated value **directly from Core 1's cache**, not from memory!

## Visual Timeline: Without Memory Barrier

```
Timeline (No Barrier):
═══════════════════════════════════════════════════════

Core 1:
T1:  x = 42 → Store Buffer (not even in L1 yet!)
T2:  flag = 1 → Store Buffer
T3:  Store Buffer drains flag = 1 to L1 first (!)
     Cache: flag = 1 [Modified]
T4:  Invalidate message sent for flag
T100: Eventually x = 42 drains to L1
     Cache: x = 42 [Modified]

Core 2:
T5:  Read flag
     Detects [Invalid], requests from Core 1
     Gets flag = 1 ✓
T6:  Read x
     Still [Shared] with old value
     Reads x = 0 ❌ WRONG!
     
Problem: flag became visible before x!
```

## Visual Timeline: With Memory Barrier (Atomic Store)

```
Timeline (With Barrier):
═══════════════════════════════════════════════════════

Core 1:
T1:  x = 42 → Store Buffer
T2:  y = 100 → Store Buffer
T3:  atomic.Store(&flag, 1) ← BARRIER
     │
     ├─ Step 1: DRAIN store buffer
     │   - x = 42 → L1 Cache [Modified]
     │   - y = 100 → L1 Cache [Modified]
     │   - Send invalidate messages on bus
     │
     ├─ Step 2: Wait for acknowledgments
     │   - Core 2 acknowledges x [Invalid]
     │   - Core 2 acknowledges y [Invalid]
     │
     └─ Step 3: Now set flag
         - flag = 1 → L1 Cache [Modified]
         - Send invalidate message for flag

Core 2:
T10: Read flag (atomic.Load)
     Detects [Invalid]
     Sends read request on bus
     Gets flag = 1 from Core 1 ✓
     
T11: Read x
     Detects [Invalid] (from earlier invalidation)
     Sends read request on bus
     Gets x = 42 from Core 1 ✓
     
T12: Read y
     Detects [Invalid]
     Sends read request on bus
     Gets y = 100 from Core 1 ✓

Result: Sees all updates correctly!
```

## The Protocol Messages on the Bus

### Without Barrier (Race Condition Possible)

```
System Bus Messages (ordered by time):
─────────────────────────────────────────────────────

T1: [Core 1] Write x=42 (local store buffer only)
T2: [Core 1] Write flag=1 (local store buffer only)
T3: [Core 1 → Bus] "Invalidate: flag" 
    ↓
T4: [Core 2] Marks flag [Invalid]
T5: [Core 2 → Bus] "Read Request: flag"
    ↓
T6: [Core 1 → Bus] "Response: flag=1"
    ↓
T7: [Core 2] Receives flag=1 ✓
T8: [Core 2 → Bus] "Read Request: x"
    ↓
T9: [Core 1 → Bus] "Response: x=0"  ❌ Still old value!
    (x=42 not yet flushed from store buffer)
```

### With Barrier (Guaranteed Ordering)

```
System Bus Messages (ordered by time):
─────────────────────────────────────────────────────

T1: [Core 1] Write x=42 (store buffer)
T2: [Core 1] Write y=100 (store buffer)
T3: [Core 1] atomic.Store(&flag, 1)
    ↓ BARRIER FORCES DRAIN ↓
T4: [Core 1 → Bus] "Invalidate: x"
    ↓
T5: [Core 2] Marks x [Invalid]
T6: [Core 1 → Bus] "Invalidate: y"
    ↓
T7: [Core 2] Marks y [Invalid]
T8: [Core 1] x=42 now in L1 [Modified] ✓
T9: [Core 1] y=100 now in L1 [Modified] ✓
    ↓ ALL DRAINED, NOW PROCEED ↓
T10: [Core 1 → Bus] "Invalidate: flag"
     ↓
T11: [Core 2] Marks flag [Invalid]
T12: [Core 1] flag=1 now in L1 [Modified] ✓

Now when Core 2 reads:
T13: [Core 2 → Bus] "Read Request: flag"
     ↓
T14: [Core 1 → Bus] "Response: flag=1"
T15: [Core 2 → Bus] "Read Request: x"
     ↓
T16: [Core 1 → Bus] "Response: x=42" ✓ Correct!
T17: [Core 2 → Bus] "Read Request: y"
     ↓
T18: [Core 1 → Bus] "Response: y=100" ✓ Correct!
```

## What "Visible" Really Means

### Three Levels of Visibility

````go
// Core 1 writes
x = 42

// Level 1: In Store Buffer
// - Visible ONLY to Core 1
// - Not even in Core 1's L1 cache yet
// - Other cores: Cannot see

// Level 2: In Core 1's L1 Cache [Modified]
// - Visible to Core 1
// - Other cores have [Invalid] state
// - Other cores: Know it's stale, can request it

// Level 3: Globally Visible (via cache coherency)
// - In Core 1's cache [Modified] or [Shared]
// - Invalidate messages sent
// - Other cores: Can fetch via bus request
// - This is what "visible via cache coherency" means!

// Level 4: In Main Memory (optional)
// - Eventually written back
// - But not necessary for visibility!
````

### The Key Insight

**"Visible via cache coherency" means:**

```
Core 2 doesn't need to read from RAM.
It can get the value directly from Core 1's cache
via the cache coherency protocol (MESI).

The value is "visible" because:
1. Core 1's cache has it in [Modified] state
2. Core 2 knows its copy is [Invalid]
3. Core 2 can request it via the bus
4. Core 1 will respond with the updated value

NO MAIN MEMORY ACCESS NEEDED!
```

## Practical Example

````go
package main

import (
    "fmt"
    "sync/atomic"
    "runtime"
)

var (
    x    int64
    flag int64
)

func writer() {
    x = 42
    atomic.StoreInt64(&flag, 1)  // Memory barrier
}

func reader() {
    if atomic.LoadInt64(&flag) == 1 {  // Memory barrier
        fmt.Println(x)  // Will print 42
    }
}

func main() {
    runtime.GOMAXPROCS(2)  // Use 2 CPU cores
    
    go writer()
    go reader()
}
````

**What happens behind the scenes:**

```
Cache Coherency Messages:
──────────────────────────────────────────────────

1. Writer (Core 1):
   x = 42 → Store Buffer
   atomic.Store(&flag, 1) triggers:
   
   Step A: Drain x = 42 to L1 cache
   Step B: Send "Invalidate x" message on bus
           Core 2 receives → marks x [Invalid]
   
   Step C: Wait for acknowledgment
           Core 2 sends ACK
   
   Step D: Now write flag = 1 to L1 cache
   Step E: Send "Invalidate flag" message on bus
           Core 2 receives → marks flag [Invalid]

2. Reader (Core 2):
   atomic.Load(&flag) triggers:
   
   Step A: Check cache → flag [Invalid]
   Step B: Send "Read Request: flag" on bus
   Step C: Core 1 responds with flag = 1
   Step D: Cache now has flag = 1 [Shared]
   
   Step E: Load sees flag = 1, executes println(x)
   Step F: Check cache → x [Invalid]
   Step G: Send "Read Request: x" on bus
   Step H: Core 1 responds with x = 42
   Step I: Cache now has x = 42 [Shared]
   Step J: Print 42 ✓

Result: Value was transferred Core 1 → Core 2
        via cache coherency protocol
        WITHOUT going to main memory!
```

## Summary

**"Visible via cache coherency to other cores" means:**

1. **The value is in a cache** (usually L1 or L3), not necessarily main memory
2. **The cache coherency protocol** (MESI, MOESI, etc.) ensures consistency
3. **Other cores can fetch it** via bus requests/snooping
4. **Invalidation messages** tell other cores their cached copies are stale
5. **Cache-to-cache transfers** happen directly, bypassing memory

The memory barrier ensures this visibility happens in the correct order:
- Forces store buffer to drain to cache
- Sends invalidation messages to other cores
- Ensures cache coherency protocol has propagated the updates
- Only then allows subsequent operations to proceed

So when Core 2 reads after the barrier, the cache coherency protocol **guarantees** it will see the updated values, either from another core's cache or from memory.