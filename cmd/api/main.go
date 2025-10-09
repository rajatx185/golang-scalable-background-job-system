package main

import (
    "fmt"
    "github.com/rajatx185/golang-scalable-background-job-system/internal/queue"
)

func main() {
    fmt.Println("API starting")
    _ = queue1.Enqueue(/*...*/)
}