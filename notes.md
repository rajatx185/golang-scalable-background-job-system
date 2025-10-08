
---

## 🧠 **Go Concurrency — Core Concepts Revision Notes**

### 1. **Goroutines vs Threads**

* **Goroutines** are lightweight functions managed by the Go runtime.
* **Threads** are OS-managed and much heavier (MBs of stack memory vs a few KBs for goroutines).
* The Go runtime multiplexes thousands of goroutines over a small pool of threads.
* You can spawn hundreds of thousands of goroutines without major overhead.
* **Trade-off:** You lose fine-grained control over thread scheduling but gain simplicity and scalability.

📌 **Analogy:**

> A goroutine is to a thread what a coroutine is to a process — a lightweight, cooperative execution unit.

---

### 2. **GOMAXPROCS**

* Defines how many OS threads can **execute Go code in parallel** (true parallelism).
* Default = number of CPU cores.
* Controls the number of “P” (processor contexts) in Go’s scheduler.
* Concurrency ≠ Parallelism:

  * Concurrency = many tasks *in progress*
  * Parallelism = many tasks *executing simultaneously*

🧩 **Scheduler Model (G-M-P):**

| Component | Meaning             | Role                                    |
| --------- | ------------------- | --------------------------------------- |
| **G**     | Goroutine           | Logical unit of work                    |
| **M**     | Machine (OS Thread) | Executes Go code                        |
| **P**     | Processor           | Holds the local run queue of goroutines |

* `GOMAXPROCS` = number of **P** (and hence maximum parallel goroutines).
* The scheduler uses **work-stealing** to balance load between Ps.

---

### 3. **Data Races & Memory Safety**

* A **data race** occurs when:

  1. Two goroutines access the same memory location.
  2. At least one access is a write.
  3. There’s no synchronization (no happens-before order).
* Results: unpredictable values, memory corruption, crashes.
* **Scheduler does not prevent races** — synchronization is the developer’s job.

🧰 **Tools for Avoiding Races**

| Mechanism       | Use Case                                      | Example                              |
| --------------- | --------------------------------------------- | ------------------------------------ |
| **sync.Mutex**  | Protect shared state (maps, structs)          | `mu.Lock(); counter++ ; mu.Unlock()` |
| **sync/atomic** | Fast atomic ops for simple integers           | `atomic.AddInt64(&counter, 1)`       |
| **Channels**    | Message passing; one goroutine owns the state | `ch <- value` / `<-ch`               |

🧩 **Race Detector:**

```bash
go test -race ./...
```

---

### 4. **Go Memory Model (Happens-Before)**

* Defines how reads/writes to shared variables are ordered.
* Synchronization (mutexes, atomics, channels) creates **happens-before relationships**, guaranteeing visibility and safety.
* If two operations are not ordered by happens-before → possible data race.

**Key Happens-Before Rules:**

* Unlock happens-before the next Lock on the same mutex.
* Send happens-before the corresponding receive on the same channel.
* Close of a channel happens-before a receive returning zero value.
* Writes before goroutine start are visible in that goroutine.

---

### 5. **Design Trade-offs**

| Technique                  | Pros                                  | Cons                          |
| -------------------------- | ------------------------------------- | ----------------------------- |
| **Mutexes**                | Simple, correct for complex state     | Can block or cause contention |
| **Atomics**                | Very fast, low overhead               | Only for simple scalar data   |
| **Channels / Actor Model** | Easier reasoning, avoids shared state | May add latency / complexity  |

---

### 6. **Quick Code Patterns**

**Unsafe (racy):**

```go
counter++ // multiple goroutines
```

**With Mutex:**

```go
mu.Lock()
counter++
mu.Unlock()
```

**With Atomic:**

```go
atomic.AddInt64(&counter, 1)
```

**With Channel (actor):**

```go
ch <- struct{}{} // send increment request
```

---

### 7. **Quick Interview Pointers**

* “Goroutines are scheduled by Go’s runtime, not the OS.”
* “GOMAXPROCS limits how many threads can run Go code simultaneously.”
* “A data race happens when concurrent reads/writes lack synchronization.”
* “Channels establish happens-before relationships automatically.”
* “Use `go test -race` religiously during development.”

---

Would you like our **next Socratic exploration** to go into:

1. **Mutexes & lock contention** — how they actually behave under Go’s scheduler, and when they hurt performance,
   or
2. **Channels & communication patterns** — how Go models concurrency through message passing instead of shared memory?
