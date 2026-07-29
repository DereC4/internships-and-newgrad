package main

import "fmt"

func main() {
	// gonna review pointers from freshman year of college lol
	x := 42
	pointer1 := &x
	fmt.Println(pointer1)
	fmt.Println(x)
	fmt.Println(*pointer1 + 1)
}
