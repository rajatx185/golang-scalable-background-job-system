// // Goroutine 1
// mu.Lock()     // Got the lock
// counter++     // Working...

// // Goroutine 2 (tries to lock while G1 holds it)
// mu.Lock()     // Lock taken! Need to wait...
//               // ↓ THIS IS WHERE IT GETS HEAVY
//               // 1. System call to kernel (expensive!)
//               // 2. Kernel puts G2 to sleep (context switch)
//               // 3. Kernel scheduler decides when to wake G2
//               // 4. When G1 unlocks, kernel wakes G2 (another context switch)
//               // 5. G2 finally gets the lock
