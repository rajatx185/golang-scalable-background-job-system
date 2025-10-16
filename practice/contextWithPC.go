// package main

// import (
// 	"context"
// 	"fmt"
// 	"time"
// )

// func main() {
//  root := context.Background()

//  // create child context
//  parent, cancel1 := context.WithTimeout(root, 1*time.Second)
//  defer cancel1()

//  // create grandchild context
//  child, cancel2 := context.WithTimeout(parent, 6*time.Second)
//  defer cancel2()

//  go task(child, "child")

//  go task(parent, "parent")

//  time.Sleep(5 * time.Second)
// }


// func task (ctx context.Context, name string) {
// 	<- ctx.Done()
// 	fmt.Printf("%s context cancelled: %v\n", name, ctx.Err())
// }