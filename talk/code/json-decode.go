package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Site struct {
	Title string
	URL   string
}

const stream = `
	{"Title": "The Go Programming Language", "URL": "http://golang.org"}
	{"Title": "Google", "URL": "http://google.com"}
`

func main() {
	dec := json.NewDecoder(strings.NewReader(stream))
	for {
		var s Site
		if err := dec.Decode(&s); err != nil {
			log.Fatal(err)
		}
		fmt.Println(s.Title, s.URL)
	}
}
