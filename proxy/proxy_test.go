package proxy

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

const serverPath = "code.google.com/p/whispering-gophers/proxy/server"

func TestIntegration(t *testing.T) {
	dir, err := ioutil.TempDir("", "proxy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "server")
	build := exec.Command("go", "build", "-o", bin, serverPath)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("Building server: %v", err)
	}

	server := exec.Command(bin, "-addr=localhost:0", "-test")
	stdout, err := server.StdoutPipe()
	if err != nil {
		t.Fatalf("Server stdout pipe: %v", err)
	}
	server.Stderr = os.Stderr
	if err := server.Start(); err != nil {
		t.Fatalf("Starting server: %v", err)
	}
	defer server.Process.Kill()

	if _, err := fmt.Fscan(stdout, proxyAddr); err != nil {
		t.Fatalf("Scanning server address: %v", err)
	}
	t.Logf("Server running on %v", *proxyAddr)

	l, err := Listen()
	if err != nil {
		t.Fatal(err)
	}

	const a, b = "Hello", "Ahoy!"
	const reps = 200

	var wg sync.WaitGroup
	defer wg.Wait()

	for i := 0; i < reps; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c, err := Dial(l.Addr().String())
			if err != nil {
				t.Errorf("Dial error: %v", err)
				return
			}
			if _, err := fmt.Fprintln(c, a, n); err != nil {
				t.Errorf("Dialer write error: %v", err)
				return
			}
			var s string
			var n2 int
			if _, err := fmt.Fscan(c, &s, &n2); err != nil {
				t.Errorf("Dialler scan error: %v", err)
				return
			}
			if s != b {
				t.Errorf("Dialler read %q, want %q", s, b)
				return
			}
			if n != n2 {
				t.Errorf("Dialler read %v, want %v", n, n2)
				return
			}
			if err := c.Close(); err != nil {
				t.Errorf("Dialler close error: %v", err)
			}
		}(i)
	}

	for i := 0; i < reps; i++ {
		c, err := l.Accept()
		if err != nil {
			t.Fatal("Accept error:", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s string
			var n int
			if _, err := fmt.Fscan(c, &s, &n); err != nil {
				t.Errorf("Listener scan error: %v", err)
				return
			}
			if s != a {
				t.Errorf("Listener read %q, want %q", s, a)
				return
			}
			if _, err := fmt.Fprintln(c, b, n); err != nil {
				t.Errorf("Listener write error: %v", err)
				return
			}
			if err := c.Close(); err != nil {
				t.Errorf("Listener conn close error: %v", err)
			}
		}()
	}

	if err := l.Close(); err != nil {
		t.Fatal("Listener close error: ", err)
	}
}
