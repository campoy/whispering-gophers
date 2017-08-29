// This program listens to the host and port specified by the -listen flag and
// dumps any incoming data to standard output.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var addr = flag.String("listen", "localhost:8000", "server listen address")

type dumpWriter struct {
	c net.Conn
	w io.Writer
}

func (w dumpWriter) Write(v []byte) (int, error) {
	fmt.Fprintf(w.w, "[%v->%v] ", w.c.RemoteAddr(), w.c.LocalAddr())
	return w.w.Write(v)
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on", l.Addr())
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go io.Copy(dumpWriter{c, os.Stdout}, c)
	}
}
