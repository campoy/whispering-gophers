package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/whispering-gophers/util"
)

var (
	httpAddr = flag.String("http", "localhost:8080", "HTTP server address")
	peerAddr = flag.String("peer", "", "peer host:port")
	dedup    = flag.Bool("dedup", true, "de-duplicate messages")
	self     string
)

type Message struct {
	ID   string
	Addr string
	Body string
}

func main() {
	flag.Parse()

	l, err := util.Listen()
	if err != nil {
		log.Fatal(err)
	}
	self = l.Addr().String()
	log.Println("Listening on", self)

	if *peerAddr != "" {
		go dial(*peerAddr)
	}
	go readInput()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go serve(c)
		}
	}()

	http.HandleFunc("/", rootHandler)
	http.Handle("/log", websocket.Handler(logHandler))
	err = http.ListenAndServe(*httpAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

var peers = &Peers{m: make(map[string]chan<- Message)}

type Peers struct {
	m  map[string]chan<- Message
	mu sync.RWMutex
}

func (p *Peers) Add(addr string) <-chan Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.m[addr]; ok {
		return nil
	}
	ch := make(chan Message)
	p.m[addr] = ch
	return ch
}

func (p *Peers) Remove(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, addr)
}

func (p *Peers) List() []chan<- Message {
	p.mu.RLock()
	defer p.mu.RUnlock()
	l := make([]chan<- Message, 0, len(p.m))
	for _, ch := range p.m {
		l = append(l, ch)
	}
	return l
}

func broadcast(m Message) {
	for _, ch := range peers.List() {
		select {
		case ch <- m:
		default:
			// Okay to drop messages sometimes.
		}
	}
}

func serve(c net.Conn) {
	log.Println("<", c.RemoteAddr(), "accepted connection")
	d := json.NewDecoder(c)
	for {
		var m Message
		err := d.Decode(&m)
		if err != nil {
			log.Println("<", c.RemoteAddr(), "error:", err)
			break
		}
		if Seen(m.ID) {
			continue
		}
		log.Printf("< %v received: %v", c.RemoteAddr(), m)
		fmt.Println(m.Body)
		broadcast(m)
		go dial(m.Addr)
	}
	c.Close()
	log.Println("<", c.RemoteAddr(), "close")
}

func readInput() {
	r := bufio.NewReader(os.Stdin)
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		m := Message{
			ID:   util.RandomID(),
			Addr: self,
			Body: s[:len(s)-1],
		}
		Seen(m.ID)
		broadcast(m)
	}
}

func dial(addr string) {
	if addr == self {
		return // Don't try to dial self.
	}

	ch := peers.Add(addr)
	if ch == nil {
		return // Peer already connected.
	}
	defer peers.Remove(addr)

	log.Println(">", addr, "dialling")
	c, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(">", addr, "dial error:", err)
		return
	}
	log.Println(">", addr, "connected")
	defer func() {
		c.Close()
		log.Println(">", addr, "closed")
	}()

	e := json.NewEncoder(c)
	for m := range ch {
		err := e.Encode(m)
		if err != nil {
			log.Println(">", addr, "error:", err)
			return
		}
	}
}

var seenIDs = struct {
	m map[string]bool
	sync.Mutex
}{m: make(map[string]bool)}

func Seen(id string) bool {
	if !*dedup || id == "" {
		return false
	}
	seenIDs.Lock()
	ok := seenIDs.m[id]
	seenIDs.m[id] = true
	seenIDs.Unlock()
	return ok
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var data = struct {
		Addr string
		Self string
	}{
		Addr: *httpAddr,
		Self: self,
	}
	err := rootTemplate.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html><head>
	<script>
var log, websocket;

function onMessage(e) {
	log.innerText += e.data;
	log.scrollTop = log.scrollHeight;
}

function init() {
	log = document.getElementById("log");
	websocket = new WebSocket("ws://{{.Addr}}/log");
	websocket.onmessage = onMessage;
	websocket.onclose = console.log;
}

window.addEventListener("load", init, false);
	</script>
	<style>
body {
	font-family: sans-serif;
}
#self, #log {
	position: absolute;
}
#self {
	height: 15%;
	font-size: 50px;
	text-align: center;
}
#log {
	top: 15%;
	height: 80%;
	font-size: 20px;
	overflow: auto;
}
	</style>
</head><body>
	<div id="self">{{.Self}}</div>
	<div id="log"></div>
</body>
</html>
`))

type Logger struct {
	m  map[string]chan<- []byte
	mu sync.Mutex
}

var logger = &Logger{m: make(map[string]chan<- []byte)}

func init() {
	log.SetFlags(0)
	log.SetOutput(io.MultiWriter(os.Stdout, logger.Writer()))
}

func (l *Logger) Writer() io.Writer {
	r, w := io.Pipe()
	go func() {
		b := make([]byte, 1<<16)
		for {
			n, err := r.Read(b)
			if err != nil {
				fmt.Println(err)
				return
			}
			l.mu.Lock()
			for _, ch := range l.m {
				b2 := make([]byte, n)
				copy(b2, b)
				select {
				case ch <- b2:
				case <-time.After(100 * time.Millisecond):
					println("drop")
				}
			}
			l.mu.Unlock()
		}
	}()
	return w
}

func (l *Logger) WriteTo(w io.Writer) error {
	id := util.RandomID()
	ch := make(chan []byte)
	l.mu.Lock()
	l.m[id] = ch
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		delete(l.m, id)
		l.mu.Unlock()
	}()
	for b := range ch {
		_, err := w.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func logHandler(ws *websocket.Conn) {
	err := logger.WriteTo(ws)
	if err != nil {
		fmt.Println(err)
	}
}
