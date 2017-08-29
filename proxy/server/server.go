// The server command is a multiplexer service for proxied TCP connections.
// Its clients access it through the code.google.com/p/whispering-gophers/proxy package.
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var (
	listenAddr = flag.String("addr", "localhost:2000", "listen address")
	testMode   = flag.Bool("test", false, "print listen address (for integration test)")
)

func main() {
	flag.Parse()

	s := NewServer()
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	if *testMode {
		fmt.Println(l.Addr())
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.Serve(c)
	}
}

// Server tracks Listeners and the address pool.
type Server struct {
	mu     sync.Mutex
	key    map[string]*Listener
	addr   map[string]*Listener
	lastIP net.IP
}

func NewServer() *Server {
	return &Server{
		key:    map[string]*Listener{},
		addr:   map[string]*Listener{},
		lastIP: net.IP{10, 0, 0, 0},
	}
}

func (s *Server) Serve(c net.Conn) {
	var cmd, arg string
	_, err := fmt.Fscan(c, &cmd, &arg)
	if err != nil {
		log.Println("%v: bad command:", c.RemoteAddr(), err)
		c.Close()
		return
	}
	switch cmd {
	case "LISTEN":
		s.Listen(c)
	case "ACCEPT":
		s.Accept(c, arg)
	case "CLOSE":
		s.Close(c, arg)
	case "DIAL":
		s.Dial(c, arg)
	default:
		log.Println("%v: bad command:", c.RemoteAddr(), cmd)
		c.Close()
	}
}

func (s *Server) Listen(c net.Conn) {
	s.mu.Lock()
	incIP(s.lastIP)
	addr, key := s.lastIP.String(), genkey()
	l := NewListener(addr)
	s.key[key] = l
	s.addr[addr] = l
	s.mu.Unlock()
	fmt.Fprintln(c, addr, key)
	c.Close()
}

func incIP(ip net.IP) {
	ip[3]++
	if ip[3] == 0 {
		ip[3] = 1
		ip[2]++
		if ip[2] == 0 {
			ip[1]++
		}
	}
}

func genkey() string {
	b := make([]byte, 8)
	n, _ := rand.Read(b)
	return fmt.Sprintf("%x", b[:n])
}

func (s *Server) Accept(c net.Conn, key string) {
	defer c.Close()

	s.mu.Lock()
	l, ok := s.key[key]
	s.mu.Unlock()
	if !ok {
		fmt.Fprintln(c, "ERROR unknown key")
		return
	}

	ch := make(chan net.Conn)
	l.accept <- ch
	c2 := <-ch
	if c2 == nil {
		fmt.Fprintln(c2, "ERROR duplicate accept")
		return
	}
	defer c2.Close()
	fmt.Fprintln(c2, "OK")
	fmt.Fprintln(c, c2.RemoteAddr())

	errc := make(chan error, 1)
	go cp(errc, c, c2)
	go cp(errc, c2, c)
	if err := <-errc; err != nil {
		log.Println(err)
	}
}

func cp(errc chan error, w io.Writer, r io.Reader) {
	_, err := io.Copy(w, r)
	errc <- err
}

func (s *Server) Close(c net.Conn, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.key[key]
	if !ok {
		fmt.Fprintln(c, "ERROR unknown key")
		return
	}
	l.close <- true
	delete(s.key, key)
	delete(s.addr, l.Addr)
}

func (s *Server) Dial(c net.Conn, addr string) {
	s.mu.Lock()
	l, ok := s.addr[addr]
	s.mu.Unlock()
	if !ok {
		fmt.Fprintln(c, "ERROR unknown address")
		return
	}
	l.dial <- c
}

// Listener represents an listening TCP socket.
type Listener struct {
	Addr   string
	accept chan chan net.Conn
	dial   chan net.Conn
	close  chan bool
}

func NewListener(addr string) *Listener {
	l := &Listener{
		Addr:   addr,
		accept: make(chan chan net.Conn),
		dial:   make(chan net.Conn),
		close:  make(chan bool),
	}
	go l.loop()
	return l
}

func (l *Listener) loop() {
	var acpt chan net.Conn
	var dial []net.Conn
	for {
		select {
		case ch := <-l.accept:
			if acpt != nil {
				acpt <- nil
			}
			acpt = ch
			if len(dial) > 0 {
				acpt <- dial[0]
				dial = dial[1:]
				acpt = nil
			}
		case c := <-l.dial:
			if acpt != nil {
				acpt <- c
				acpt = nil
			} else {
				dial = append(dial, c)
			}
		case <-l.close:
			if acpt != nil {
				acpt <- nil
			}
			for _, c := range dial {
				c.Close()
			}
			return
		}

	}
}
