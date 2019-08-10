package main

import (
	"testing"
	"time"
)

func TestPeers(t *testing.T) {
	peers := &Peers{m: make(map[string]chan<- Message)}
	done := make(chan bool, 1)

	var chA, chB <-chan Message
	go func() {
		defer func() { done <- true }()
		if chA = peers.Add("a"); chA == nil {
			t.Fatal(`peers.Add("a") returned nil, want channel`)
		}
	}()
	go func() {
		defer func() { done <- true }()
		if chB = peers.Add("b"); chB == nil {
			t.Fatal(`peers.Add("b") returned nil, want channel`)
		}
	}()
	<-done
	<-done
	if chA == chB {
		t.Fatal(`peers.Add("b") returned same channel as "a"!`)
	}
	if ch := peers.Add("a"); ch != nil {
		t.Fatal(`second peers.Add("a") returned non-nil channel, want nil`)
	}
	if ch := peers.Add("b"); ch != nil {
		t.Fatal(`second peers.Add("b") returned non-nil channel, want nil`)
	}

	list := peers.List()
	if len(list) != 2 {
		t.Fatalf("peers.List() returned a list of length %d, want 2", len(list))
	}

	go func() {
		for _, ch := range list {
			select {
			case ch <- Message{Body: "foo"}:
			case <-time.After(10 * time.Millisecond):
			}
		}
		done <- true
	}()
	select {
	case m := <-chA:
		if m.Body != "foo" {
			t.Fatalf("received message %q, want %q", m.Body, "foo")
		}
	case <-done:
		t.Fatal(`didn't receive message on "a" channel`)
	}
	<-done

	peers.Remove("a")

	list = peers.List()
	if len(list) != 1 {
		t.Fatalf("peers.List() returned a list of length %d, want 1", len(list))
	}

	go func() {
		select {
		case list[0] <- Message{Body: "bar"}:
		case <-time.After(10 * time.Millisecond):
		}
		done <- true
	}()
	select {
	case m := <-chB:
		if m.Body != "bar" {
			t.Fatalf("received message %q, want %q", m.Body, "bar")
		}
	case <-done:
		t.Fatal(`didn't receive message on "b" channel`)
	}
}
