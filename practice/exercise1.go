package main


// **Questions for you**:
// 1. Identify all happens-before relationships in this code
// 2. Is there a race condition? If so, where?
// 3. How does closing channels establish happens-before?

import (
    "sync"
)

type Task struct {
    ID   int
    Data string
}

type Result struct {
    ID    int
    Value string
}

func workerPool() {
    tasks := make(chan Task, 10)
    results := make(chan Result, 10)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go worker(tasks, results, &wg)
    }
    
    // Send tasks
    go func() {
        for i := 0; i < 100; i++ {
            tasks <- Task{ID: i, Data: "task data"}
        }
        close(tasks)
    }()
    
    // Wait and close results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    for result := range results {
        println(result.ID, result.Value)
    }
}

func worker(tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for task := range tasks {
        // Process task
        results <- Result{ID: task.ID, Value: "processed"}
    }
}

