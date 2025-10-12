package main 

import (
	"context"
	"fmt"
	"time"
	"net/http"
	"io"
)

func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
 // Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel() // Ensure resources are cleaned up

// Create HTTP request with context
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
	return "", err
}

// Execute request
resp, err := http.DefaultClient.Do(req)
if err != nil {
	return "", err
}
defer resp.Body.Close()

// Read response
body, err := io.ReadAll(resp.Body)
if err != nil {
	return "", err
}


// Return response body as string
return string(body), nil
}


func main() {
	// Try to fetch with 5-second timeout
	result, err := fetchWithTimeout("https://httpbin.org/delay/2", 5*time.Second)
	if err != nil {
		fmt.Println("Error:", err) // Will Timeout
		return
	}
	fmt.Println("Fetched content:", result)
}

