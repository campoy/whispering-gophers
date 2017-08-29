// Solution to part 4 of the Whispering Gophers code lab.
//
// This program is a combination of parts 2 and 3.
//
// It listens on the host and port specified by the -listen flag.
// For each incoming connection, it launches a goroutine that reads and decodes
// JSON-encoded messages from the connection and prints them to standard
// output.
// It concurrently makes a connection the host and port specified by the -dial
// flag, reads lines from standard input, and writes JSON-encoded messages to
// the network connection.
//
// You can test it by running part3 in one terminal:
// 	$ part3 -listen=localhost:8000
// Running this program in another terminal:
// 	$ part4 -listen=localhost:8001 -dial=localhost:8000
// And running part2 in another terminal:
// 	$ part2 -dial=localhost:8001
// Lines typed in the second terminal should appear as JSON objects in the
// first terminal, and those typed at the third terminal should appear in the
// second.
//
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	listenAddr = flag.String("listen", "", "host:port to listen on")
	dialAddr   = flag.String("dial", "", "host:port to dial")
)

type Message struct {
	Body string
}

func main() {
	flag.Parse()

	go dial(*dialAddr)

	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serve(c)
	}
}

func serve(c net.Conn) {
	defer c.Close()
	d := json.NewDecoder(c)
	for {
		var m Message
		err := d.Decode(&m)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%#v\n", m)
	}
}

func dial(addr string) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(os.Stdin)
	e := json.NewEncoder(c)
	for s.Scan() {
		m := Message{Body: s.Text()}
		err := e.Encode(m)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
}
