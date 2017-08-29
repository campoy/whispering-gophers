package main

import (
	"fmt"
	"time"
)

func main() {
	go say("let's go!", 3*time.Second)
	go say("ho!", 2*time.Second)
	go say("hey!", 1*time.Second)
	time.Sleep(4 * time.Second)
}

func say(text string, duration time.Duration) {
	time.Sleep(duration)
	fmt.Println(text)
}
