package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 5; i++ {
		t := <-ticker.C
		fmt.Println("Ticked at", t)
	}
}
