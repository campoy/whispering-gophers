package main

import (
	"encoding/json"
	"log"
	"os"
)

type Site struct {
	Title string
	URL   string
}

var sites = []Site{
	{"The Go Programming Language", "http://golang.org"},
	{"Google", "http://google.com"},
}

func main() {
	enc := json.NewEncoder(os.Stdout)
	for _, s := range sites {
		err := enc.Encode(s)
		if err != nil {
			log.Fatal(err)
		}
	}
}
