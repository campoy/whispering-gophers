package main

import (
	"fmt"
	"time"
)

func worker(id string, work chan string) {
	for s := range work {
		fmt.Println(id, "received", s)
	}
}

func main() {
	workers := make(map[string]chan string)
	for i := 0; i < 4; i++ {
		id := fmt.Sprint("worker", i)
		ch := make(chan string)
		go worker(id, ch)
		workers[id] = ch
	}
	for i := 0; ; i++ {
		for _, ch := range workers {
			ch <- fmt.Sprint("task", i)
		}
		time.Sleep(time.Second)
	}
}
