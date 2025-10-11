package main

import (
	"fmt"
	"context"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker(ctx)
	
	time.Sleep((2 * time.Second))

	fmt.Println("cancelling context")
	cancel()

	time.Sleep((2 * time.Second))
	fmt.Println("exiting main")
}

func worker (ctx context.Context) {
	for {
		select {
		case <- ctx.Done():
			fmt.Println("Worker: context cancelled, stopping work")
			fmt.Println("Reason:", ctx.Err())
			return
		default:
			fmt.Println("Worker: working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}