// Solution to part 1 of the Whispering Gophers code lab.
// This program reads from standard input and writes JSON-encoded messages to
// standard output. For example, this input line:
//	Hello!
// Produces this output:
//	{"Body":"Hello!"}
//
package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

type Message struct {
	Body string
}

func main() {
	// TODO: Create a new bufio.Scanner reading from the standard input.
	// TODO: Create a new json.Encoder writing into the standard output.
	for /* TODO: Iterate over every line in the scanner */ {
		// TODO: Create a new message with the read text.
		// TODO: Encode the message, and check for errors!
	}
	// TODO: Check for a scan error.
}
