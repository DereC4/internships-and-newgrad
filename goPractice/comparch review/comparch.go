package main

import "fmt"

// When passing structs to functions, Go copies the data
type User struct {
	Username string
	Age      int
	Active   bool
}

// alignment / order matters, just like in C!!!
type BadStruct struct {
	A bool  // 1 byte  (+ 7 bytes padding)
	B int64 // 8 bytes
	C bool  // 1 byte  (+ 7 bytes padding)
}

type GoodStrut struct {
	B int64 // 8 bytes
	A bool  // 1 byte
	C bool  // 1 byte  (+ 6 bytes padding)
}

func main() {
	// gonna review pointers from freshman year of college lol
	x := 42
	pointer1 := &x
	fmt.Println(pointer1)
	fmt.Println(x)
	fmt.Println(*pointer1 + 1)

	u1 := User{
		Username: "alice_dev",
		Age:      28,
		Active:   true,
	}
	u1.Active = false
}
