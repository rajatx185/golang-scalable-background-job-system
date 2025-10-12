// package main

// import (
// 	"fmt"
// 	"context"
// 	"time"
// )


// func main() {
// 	deadline := time.Now().Add(2 * time.Second)
// 	ctx, cancel := context.WithDeadline(context.Background(), deadline)
// 	defer cancel()

// 	go work(ctx)

// 	time.Sleep(1 * time.Second)
// 	fmt.Println("Main function completed")
// }

// func work(ctx context.Context) {
// 	select {
// 		case <- ctx.Done():
// 			fmt.Println("Work cancelled:", ctx.Err())
// 			return
// 	}
// }