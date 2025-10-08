package main

import "fmt"

// Variables
var x int = 10
var y = 10  // type inferred as int
var n int; // zero value
var s string; // ""
var t bool; // false

// functions
func add(a int, b int) int {
	return a + b
}

func divide(a int, b int) (int, bool) {
	if b == 0 {
		return 0, false
	} else {
		return a/b, true
	}
}


// Structs (custom types)
type User struct {
	Name string
	Age int
}

// Interfaces
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

// func main() {
// 	// z := 30 // short declaration (same as var z = 30), only works inside functions
// 	// fmt.Println("%d Hello World", z)
// 	fmt.Println(add(1, 2))
// 	fmt.Println(divide(10,2))

// 	x = 20 // reassignment

// 	// Conditionals
// 	if x == 10 {
// 	fmt.Println("x is 10")
// 	} else if x == 20 {
// 		fmt.Println("x is 20")
// 	} else {
// 		fmt.Println("x is neither 10 nor 20")
// 	}

// 	// Only Loop Keyword
// 	for i:=0; i<10; i++ {
// 		fmt.Println(i)
// 	}

// 	// slice of integers
// 	nums := []int{1,2,3,4}

// 	// range
// 	for index, value := range nums {
// 		fmt.Println(index, value)
// 	}

// 	// switch
// 	switch day := 3; day {
// 	case 1,2,3:
// 		fmt.Println("Early Week")
// 	default:
// 		fmt.Println("Late Week")
// 	}

// 	// Arrays (fixed size)
// 	var arr0 [5]int = [5]int{1,2,3,4,5}
// 	arr1 := [...] int{1,2,3,4,5,6} // size inferred
// 	arr2 := [5]int{1,2,3,4,5} // short declaration (most common inside functions)
// 	fmt.Println(arr0)
// 	fmt.Println(arr1)
// 	fmt.Println(arr2)

// 	// Slices (dynamic size)
// 	slice0 := []int{1,2,3,4,5} // size inferred
// 	slice1 := make([]int, 5) // make function
// 	slice2 := make([]int, 5, 10) // length 5, capacity 10
// 	slice0 = append(slice0, 6) // append function
// 	fmt.Println(slice0)
// 	fmt.Println(slice1)
// 	fmt.Println(slice2)

// 	// Maps (dictionaries)
// 	dict0 := map[string]int{"one": 1, "two": 2}
// 	dict1 := make(map[string]int)
// 	dict1["three"] = 3
// 	dict1["four"] = 4
// 	fmt.Println(dict0)
// 	fmt.Println(dict1)
// 	// Check if key exists
// 	value, exists := dict0["one"]
// 	if exists {
// 		fmt.Println(value)
// 	}

// 	user1 := User{Name: "Rajat", Age: 30}
// 	user2 := User{Name: "Kriti", Age: 30}

// 	fmt.Println(user1.Name)
// 	fmt.Println(user2.Age)
// 	user1.Age = 31 // update

// 	var user3 User // zero value
// 	fmt.Println(user3) // {"" 0}

// 	// Pointers
// 	a := 10
// 	b := &a // address of a
// 	fmt.Println(a)
// 	fmt.Println(b) // address
// 	fmt.Println(*b) // value at address
// 	*b = 20
// 	fmt.Println(a) // a is now 20

// 	// Method (attached to type)
// 	circle := Circle{Radius: 5.0}
// 	area := circle.Area()  // Called ON the Circle instance
// 	fmt.Println(area)

// 	// Regular Function
// 	// circle2 := Circle{Radius: 5.0}
// 	// area2 := Area(circle2)  // Circle passed AS parameter
// 	// fmt.Println(area2)
// }



