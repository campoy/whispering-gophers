// Solution to part 5 of the Whispering Gophers code lab.
//
// This program extends part 4.
//
// It listens on an available public IP and port, and for each incoming
// connection it decodes JSON-encoded messages and writes them to standard
// output.
// It simultaneously makes a connection to the host and port specified by -peer,
// reads lines from standard input, and writes JSON-encoded messages to the
// network connection.
// The messages include the listen address. For example:
// 	{"Addr": "127.0.0.1:41232", "Body": "Hello"!}
//
// You can test this program by listening with the dump program:
// 	$ dump -listen=localhost:8000
// In another terminal, running this program:
// 	$ part5 -peer=localhost:8000
// And in a third terminal, running part2 with the address printed by part5:
// 	$ part2 -dial=192.168.1.200:54312
// Lines typed in the third terminal should appear in the second, and those
// typed in the second window should appear in the first.
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

	"code.google.com/p/whispering-gophers/util"
)

var (
	peerAddr = flag.String("peer", "", "peer host:port")
	self     string
)

type Message struct {
	Addr string
	Body string
}

func main() {
	flag.Parse()

	// TODO: Create a new listener using util.Listen and put it in a variable named l.
	// TODO: Set the global variable self with the address of the listener.
	// TODO: Print the address to the standard output

	go dial(*peerAddr)

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
		m := Message{
			// TODO: Put the self variable in the new Addr field.
			Body: s.Text(),
		}
		err := e.Encode(m)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
}
