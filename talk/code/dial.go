package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	c, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(c, "GET /")
	io.Copy(os.Stdout, c)
	c.Close()
}
