package main

import (
	"flag"
	"log"
	"os"

	"code.google.com/p/whispering-gophers/util"
)

func init() {
	// Trick to stuff the command-line flags, because this example
	// is run from the present tool interface.
	os.Args = []string{"register", "-master", "localhost:8000"}
}

func main() {
	flag.Parse() // To parse the util package's -master flag.
	l, err := util.Listen()
	if err != nil {
		log.Fatal(err)
	}
	err = util.RegisterPeer(l.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
}
