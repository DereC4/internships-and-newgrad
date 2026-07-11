package main

import (
	"errors"
	"fmt"
)

// fmt is the standard library package for formatted I/O.

func add(a, b int) (float64, error) {
	// You wrap multiple return types in parentheses.
	if b == -1 {
		// Return a default value and a new error object
		return 0.0, errors.New("Custom Error I made")
	}

	return float64(a + b), nil
}

func main() {
	// no semicolons in golang
	fmt.Println("Hello World!")

	// keyword first, type comes last
	var x int = 3
	var defaultInt int       // 0
	var defaultString string // ""
	var defaultBool bool     // false

	// shorthand variable declaration
	urmom := "Go"

	// Unused Variables are Compiler Errors
	// No Implicit Type Conversion, must cast to use different types
	fmt.Println(x)
	fmt.Println(defaultInt)
	fmt.Println(defaultString)
	fmt.Println(defaultBool)
	fmt.Println(urmom)
	fmt.Println(add(30000, 39))

}
