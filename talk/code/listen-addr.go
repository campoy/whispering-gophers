package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on", l.Addr())
}
