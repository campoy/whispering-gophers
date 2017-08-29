package main

import "fmt"

func main() {
	// START OMIT
	ch := make(chan int)

	select {
	case ch <- 42:
		fmt.Println("Send succeeded")
	default:
		fmt.Println("Send failed")
	}
	// END OMIT
}
