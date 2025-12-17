package main

import (
	"log/slog"
	"sync"
	"time"
)

const refreshInterval = 1 * time.Second

// Server is a data structure that holds information about our UDP server, including which
// port it listens on and a map of Account structs, one for each user account
type Server struct {
	port     int
	accounts *AccountMap
	wg       sync.WaitGroup
}

// a ReplyHandler is a struct which has two functions that reply permit or deny back to the
// caller. Passing a ReplyHandler to a goroutine saves having to pass connection and address
// pointers around.
type ReplyHandler struct {
	permit func()
	deny   func()
}

// NewServer creates a new server struct, given the port
func NewServer(port int) *Server {
	accountsPtr := NewAccountMap()
	server := Server{
		port:     port,
		accounts: accountsPtr,
	}
	return &server
}

// Run executes the server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine. Another goroutine
// resets each Account's buckets periodically.
func (s *Server) Run() {

	// reset the counters every second
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			s.accounts.Reset()
		}
	}()

	// run the UDP server
	s.wg.Add(1)
	go func() {
		err := s.runUDPServer()
		if err != nil {
			slog.Error("UDP server error", "error", err)
		}
	}()

	// run the TCP server
	s.wg.Add(1)
	go func() {
		err := s.runTCPServer()
		if err != nil {
			slog.Error("TCP server error", "error", err)
		}
	}()

	// wait for all goroutines to finish
	s.wg.Wait()

}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleMessage(protocol string, str string, replyer ReplyHandler) {
	permitted := false
	var err error

	// deferred logging
	defer func() {
		if err != nil {
			slog.Error("Error handling message", "protocol", protocol, "error", err)
		} else {
			slog.Info("Handled message", "protocol", protocol, "message", str, "permitted", permitted)
		}
	}()

	// parse the incoming message
	message, err := parseMessage(str)
	if err != nil {
		str = "invalid"
		replyer.deny()
		return
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	acc := s.accounts.LoadOrStore(message.accountName)

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	permitted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// permit or deny reply
	if permitted {
		replyer.permit()
	} else {
		replyer.deny()
	}
}
