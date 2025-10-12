// package main

// import (
// 	"context"
// 	"fmt"
// 	"time"
// )

// func main() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	for i := 1; i <= 5; i++ {
// 		go worker(ctx, i)
// 	}

// 	// Wait for context to Timeout
// 	<- ctx.Done()
// 	fmt.Println("Main all workers should stop now")

// 	// Give all workers time to cleanup
// 	time.Sleep(500*time.Millisecond)
// }

// func worker(ctx context.Context, id int) {
// 	for {
// 		select {
// 			case <- ctx.Done():
// 				fmt.Printf("Worker %d stopping:\n", id)
// 				return
// 			default:
// 				fmt.Printf("Worker %d working...\n", id)
// 				time.Sleep(500 * time.Millisecond)
// 		}
// 	}
// }