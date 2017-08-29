package main

import (
	"compress/gzip"
	"log"
	"strings"
)

func main() {
	log.Println("Opening gzip stream...")
	_, err := gzip.NewReader(strings.NewReader("not a gzip stream!"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("OK!")
}
