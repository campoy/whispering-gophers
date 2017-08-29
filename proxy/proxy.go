// Package proxy provides Dial and Listen functions that behave like net.Dial
// and net.Listen, but all connections are made through a multiplexing proxy
// service that provides a virtual address space.
// Addresses obtained through this package should be treated as opaque strings,
// even though they look like regular IP addresses.
//
// The package registers a "-proxy" command-line flag that specifies the
// address of the proxy service.
package proxy

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var proxyAddr = flag.String("proxy", "localhost:2000", "remote proxy address")

// Dial opens a connection to the specified address.
func Dial(address string) (net.Conn, error) {
	c, err := net.Dial("tcp", *proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	c = logConn{"dial", c}
	_, err = fmt.Fprintf(c, "DIAL %v\n", address)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	var status string
	_, err = fmt.Fscan(c, &status)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("bad response from proxy: %v", err)
	}
	if status != "OK" {
		c.Close()
		return nil, fmt.Errorf("bad response from proxy: %v", status)
	}
	return &conn{Conn: c, remote: addr(address)}, nil
}

// Listen opens a listening socket.
// Use the Addr method of the returned Listener to obtain the listen address.
func Listen() (net.Listener, error) {
	c, err := net.Dial("tcp", *proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	c = logConn{"list", c}
	defer c.Close()
	_, err = fmt.Fprintln(c, "LISTEN nop")
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	l := &listener{}
	_, err = fmt.Fscan(c, &l.addr, &l.key)
	if err != nil {
		return nil, fmt.Errorf("bad response from proxy: %v", err)
	}
	return l, nil
}

type listener struct {
	key  string
	addr addr
}

var _ net.Listener = &listener{}

func (l *listener) Accept() (c net.Conn, err error) {
	c, err = net.Dial("tcp", *proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	c = logConn{"acpt", c}
	_, err = fmt.Fprintf(c, "ACCEPT %v\n", l.key)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("connecting to proxy: %v", err)
	}
	pc := &conn{Conn: c, local: l.addr}
	_, err = fmt.Fscan(c, &pc.remote)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("bad response from proxy: %v", err)
	}
	return pc, nil
}

func (l *listener) Close() error {
	c, err := net.Dial("tcp", *proxyAddr)
	if err != nil {
		return fmt.Errorf("connecting to proxy: %v", err)
	}
	c = logConn{"clse", c}
	defer c.Close()
	_, err = fmt.Fprintf(c, "CLOSE %v\n", l.key)
	if err != nil {
		return fmt.Errorf("bad response from proxy: %v", err)
	}
	return nil
}

func (l *listener) Addr() net.Addr {
	return addr(l.addr)
}

type conn struct {
	net.Conn
	local, remote addr
}

var _ net.Conn = &conn{}

func (c *conn) LocalAddr() net.Addr  { return c.local }
func (c *conn) RemoteAddr() net.Addr { return c.remote }

type addr string

func (a addr) Network() string { return "proxy" }
func (a addr) String() string  { return string(a) }

const verbose = false // log network traffic; for debugging proxy/server

type logConn struct {
	prefix string
	net.Conn
}

func (c logConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if verbose {
		log.Printf("%v wr %q (%v)", c.prefix, b, err)
	}
	return
}

func (c logConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if verbose {
		log.Printf("%v rd %q (%v)", c.prefix, b[:n], err)
	}
	return
}

func (c logConn) Close() (err error) {
	err = c.Conn.Close()
	if verbose {
		log.Printf("%v cl (%v)", c.prefix, err)
	}
	return
}
