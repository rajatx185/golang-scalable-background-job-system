package main

import (
	"fmt"
	"sync"
)


// var data = make (map[string]int)
// var mu sync.RWMutex


func read(key string) int {
	mu.RLock()		 // Acquire read lock
	defer mu.RUnlock() // Release read lock when function exits
	return data[key]
}

func write(key string, value int) {
	mu.Lock()		 // Acquire write lock
	defer mu.Unlock() // Release write lock when function exits
	data[key] = value
}

// func main() {
// 	var wg sync.WaitGroup
// 	for i:=0; i<5; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			write(fmt.Sprint(id), id)
// 			fmt.Println("read:", read(fmt.Sprint(id)))
// 		}(i)
// 	}	
// 	wg.Wait()
// 	fmt.Println("Final data:", data)
// }

// How RWMutex Works Internally:
// Think of RWMutex as having two types of locks:

// 1. Read Lock (RLock) - Shared Lock
// Multiple goroutines can hold read locks simultaneously
// Readers don't block each other
// Readers block writers (and vice versa)
// 2. Write Lock (Lock) - Exclusive Lock
// Only ONE goroutine can hold a write lock
// Writers block everyone (readers and other writers)
// Scenario 1: Multiple Readers (No Writers)
// ─────────────────────────────────────────
// Reader 1: RLock() ──────────────→ RUnlock()
// Reader 2:    RLock() ──────────→ RUnlock()
// Reader 3:       RLock() ─────→ RUnlock()
//                  ↑
//           All reading simultaneously!

// 		  		Scenario 2: Writer Blocks Everyone
// 		─────────────────────────────────────────
// 		Writer:   Lock() ═══════════════════════→ Unlock()
// 		Reader 1:    RLock() BLOCKED... waits... ──→ RUnlock()
// 		Reader 2:       RLock() BLOCKED... waits... ──→ RUnlock()
// 						 ↑
// 				  Writer has exclusive access

// 				Scenario 3: Readers Block Writer
// 				─────────────────────────────────────────
// 				Reader 1: RLock() ──────────────→ RUnlock()
// 				Reader 2:    RLock() ──────────→ RUnlock()
// 				Writer:         Lock() BLOCKED... waits... ═══→ Unlock()
// 								 ↑
// 						  Writer waits for all readers to finish
						  
						  
// 						go func(id int) {
// 							defer wg.Done()
// 							write(fmt.Sprint(id), id)              // Step 1: Exclusive write
// 							fmt.Println("read:", read(fmt.Sprint(id))) // Step 2: Shared read
// 						}(i)
						
// 												Goroutine 0:
// 						  write("0", 0)
// 							mu.Lock() ══════════════════════════════════════════════════════════════╗
// 							  data["0"] = 0                                                          ║
// 							mu.Unlock() ════════════════════════════════════════════════════════════╝
// 						  read("0")
// 							mu.RLock() ──────────────────────────────╗
// 							  value = data["0"]  // Gets 0            ║
// 							mu.RUnlock() ────────────────────────────╝
// 						  print "read: 0"
						
// 						Goroutine 1 (runs concurrently):
// 						  write("1", 1)
// 							mu.Lock() BLOCKED if G0 is writing... waits... ═══════════════════════╗
// 							  data["1"] = 1                                                        ║
// 							mu.Unlock() ═════════════════════════════════════════════════════════╝
// 						  read("1")
// 							mu.RLock() ──────────────────────────────╗ (Can run concurrently with G0's read!)
// 							  value = data["1"]  // Gets 1            ║
// 							mu.RUnlock() ────────────────────────────╝
// 						  print "read: 1"


// 						  Let's say Goroutine 0, 1, 2 all reach their 

// 						  						// All three can execute simultaneously!
// 						Goroutine 0: mu.RLock() ─────────────→ mu.RUnlock()
// 						Goroutine 1:    mu.RLock() ─────────→ mu.RUnlock()
// 						Goroutine 2:       mu.RLock() ────→ mu.RUnlock()
// 											  ↑
// 									   All reading data map concurrently
// 									   (Safe because no one is writing)

// 									   But if Goroutine 3 tries to write during this:

// Goroutine 0: mu.RLock() ─────────────→ mu.RUnlock()
// Goroutine 1:    mu.RLock() ─────────→ mu.RUnlock()
// Goroutine 2:       mu.RLock() ────→ mu.RUnlock()
// Goroutine 3:          mu.Lock() ⏸️ BLOCKED... waits for all reads to finish... ═══→
//                          ↑
//                  Writer must wait!

// 				 Without RWMutex (using regular Mutex):
// // ❌ Only ONE at a time (slow for read-heavy workloads)
// Goroutine 0: mu.Lock() ═══════════════════════════════════════════════════════→ mu.Unlock()
// Goroutine 1:                mu.Lock() BLOCKED... ═══════════════════════════════→ mu.Unlock()
// Goroutine 2:                                         mu.Lock() BLOCKED... ═══════→

// With RWMutex (your code):

// // ✓ Multiple readers can run simultaneously (fast!)
// Goroutine 0 read: mu.RLock() ─────────────→ mu.RUnlock()
// Goroutine 1 read:    mu.RLock() ─────────→ mu.RUnlock()
// Goroutine 2 read:       mu.RLock() ────→ mu.RUnlock()
// Goroutine 3 write:         mu.Lock() BLOCKED... waits... ═══════════════════════→

// The Key Rules:
// Multiple readers can access data simultaneously ✓
// One writer gets exclusive access (no readers, no other writers) ✓
// Readers block writers, writers block everyone ✓

// Think of RWMutex as having two types of locks:

// 1. Read Lock (RLock) - Shared Lock
// Multiple goroutines can hold read locks simultaneously
// Readers don't block each other
// Readers block writers (and vice versa)
// 2. Write Lock (Lock) - Exclusive Lock
// Only ONE goroutine can hold a write lock
// Writers block everyone (readers and other writers)
// The Key Rules:
// Multiple readers can access data simultaneously ✓
// One writer gets exclusive access (no readers, no other writers) ✓
// Readers block writers, writers block everyone ✓
// Analogy:
// Think of a library:

// Reading a book (RLock): Many people can read different books at the same time
// Writing/updating catalog (Lock): Library needs to close (kick everyone out) to update the system