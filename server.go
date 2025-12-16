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
	accounts *AccountMap
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
	accountsMap := AccountMap{}
	server := Server{
		port:     port,
		accounts: &accountsMap,
	}
	return &server
}

// Run executes the server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine. Another goroutine
// resets each Account's buckets periodically.
func (s *Server) Run() {

	var wg sync.WaitGroup

	// reset the counters every second
	wg.Go(func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			s.accounts.Range(func(key string, acc Account) bool {
				acc.reset()
				return true
			})
		}
	})

	// run the UDP server
	wg.Go(func() {
		err := s.runUDPServer()
		if err != nil {
			fmt.Println(err)
		}
	})

	// run the TCP server
	wg.Go(func() {
		err := s.runTCPServer()
		if err != nil {
			fmt.Println(err)
		}
	})

	// wait for all goroutines to finish
	wg.Wait()

}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleMessage(protocol string, str string, replyer ReplyHandler) {
	//  trim \n
	permitted := false
	var err error

	// deferred logging
	defer func() {
		log.Printf("protocol: %s message: %s permitted: %v err: %v", protocol, str, permitted, err)
	}()

	// parse the incoming message
	message, err := parseMessage(str)
	if err != nil {
		replyer.deny()
		return
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	acc, ok := s.accounts.Load(message.accountName)

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	permitted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// if this is a new account, it needs storing in the sync map
	if !ok {
		s.accounts.Store(message.accountName, acc)
	}

	// permit or deny reply
	if permitted {
		replyer.permit()
	} else {
		replyer.deny()
	}
}
