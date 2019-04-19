package main

import (
	"fmt"
	"github.com/kjk/winapigen/w"
)

var (
	// import a type from w to make sure it compiles
	f w.HWND
)

func main() {
	fmt.Printf("Hello from w tester\n")
}