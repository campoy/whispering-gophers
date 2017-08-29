package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// START OMIT
	var (
		count int
		mu    sync.Mutex // protects count
	)
	for i := 0; i < 10; i++ {
		go func() {
			for {
				mu.Lock()
				count++
				mu.Unlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
	time.Sleep(time.Second)
	mu.Lock()
	fmt.Println(count)
	mu.Unlock()
	// END OMIT
}
