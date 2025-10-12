package main

import (
	"fmt"
	"context"
	"time"
	"net/http"
)

func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func handler(w http.ResponseWriter, r *http.Request){
	// Get request context (automatically created by http.Server)
	ctx := r.Context()

	// Add request ID to context (in real world, generate a unique ID)
	ctx = context.WithValue(ctx, "requestID", generateRequestID())

	// Simulate a long-running operation
	result := make(chan string, 1)
	go processRequest(ctx, result)

	// Wait for result or client disconnect
	select {
	case res := <- result:
		fmt.Fprintf(w, "Request completed: %s\n", res)
	case <- ctx.Done():
		// Client disconnected or request cancelled
		fmt.Printf("Request %s cancelled: %v\n", ctx.Value("requestID"), ctx.Err())
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

func processRequest(ctx context.Context, result chan string) {
	requestID := ctx.Value("requestID").(string)

	// Simulate work
	for i:=0; i<5; i++ {
		select {
		case <-ctx.Done():
			// Context cancelled or timed out
			fmt.Printf("Request %s stopping early: %v\n", requestID, ctx.Err())
			return
		case <-time.After(100 * time.Millisecond):
			// Simulate work by sleeping
			fmt.Printf("Request %s processing step %d\n", requestID, i+1)
		}
	}
	
	// Send result
	result <- "Done!"
}



func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server at :8080")
	http.ListenAndServe(":8080", nil)
}