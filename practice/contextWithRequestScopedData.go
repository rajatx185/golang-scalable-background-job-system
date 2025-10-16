// package main

// import (
// 	"fmt"
// 	"context"
// )

// type contextKey string

// const (
// 	userIdKey contextKey = "userID"
// 	requestIdKey contextKey = "requestID"
// )

// func main() {
// 	// Create context with values
// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, userIdKey, "rajatx185")
// 	ctx = context.WithValue(ctx, requestIdKey, "req-12345")

// 	// pass to functions
// 	handleRequest(ctx)
// }

// func handleRequest(ctx context.Context) {
// 	// Retrieve values from context
// 	userId := ctx.Value(userIdKey).(string)
// 	requestId := ctx.Value(requestIdKey).(string)

// 	fmt.Printf("Handling request %s for user %s\n", requestId, userId)

// 	// Pass context to another function
// 	processData(ctx)
// }


// func processData(ctx context.Context) {
// 	// can still access context values
// 	userId := ctx.Value((userIdKey))
// 	fmt.Println("Processing data for user:", userId)
// }