package main

import (
	"log"

	"code.google.com/p/whispering-gophers/util"
)

func main() {
	l, err := util.Listen()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on", l.Addr())
}
