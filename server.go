package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const refreshInterval = 1 * time.Second

// Server is a data structure that holds information about our UDP server, including which
// port it listens on and a map of Account structs, one for each user account
type Server struct {
	port     int
	accounts map[string]Account
	mu       sync.RWMutex
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
	accountsMap := make(map[string]Account)
	server := Server{
		port:     port,
		accounts: accountsMap,
	}
	return &server
}

// Run executes the server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine. Another goroutine
// resets each Account's buckets periodically.
func (s *Server) Run() {

	// reset the counters every second
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			s.mu.RLock()
			for _, acc := range s.accounts {
				acc.reset()
			}
			s.mu.RUnlock()
		}
	}()

	// run the UDP server
	go func() {
		err := s.runUDPServer()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// run the TCP server
	go func() {
		err := s.runTCPServer()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// wait for all goroutines to finish
	s.wg.Add(3)
	s.wg.Wait()

}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleMessage(protocol string, str string, replyer ReplyHandler) {
	permitted := false
	var err error

	// deferred logging
	defer func() {
		log.Printf("protocol: %s message: %s permitted: %v err: %v", protocol, str, permitted, err)
	}()

	// parse the incoming message
	message, err := parseMessage(str)
	if err != nil {
		str = "invalid"
		replyer.deny()
		return
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	s.mu.RLock()
	acc, ok := s.accounts[message.accountName]
	s.mu.RUnlock()
	// if the key isn't in our map, create a new Account
	if !ok {
		acc = *NewAccount(message.accountName)
	}

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	permitted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// if this is a new account, it needs storing in the sync map
	if !ok {
		s.mu.Lock()
		s.accounts[message.accountName] = acc
		s.mu.Unlock()
	}

	// permit or deny reply
	if permitted {
		replyer.permit()
	} else {
		replyer.deny()
	}
}
