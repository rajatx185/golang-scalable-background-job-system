// package main

// import (
// 	"fmt"
// 	"context"
// 	"time"
// )


// func main() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

//     result := make(chan string)
// 	go longTask(ctx, result)

// 	select {
// 	case res := <-result:
// 		fmt.Println("Got Result:", res)
// 	case <-ctx.Done():
// 		fmt.Println("Timeout Reason:", ctx.Err())
// 	}
// }

// func longTask(ctx context.Context, result chan string) {
// 	select {
// 	case <- time.After(4 * time.Second):
// 		result <- "Task Completed"
// 	case <- ctx.Done():
// 		fmt.Println("longTask cancelled:", ctx.Err())
// 		return
// 	}
// }