package main

import (
	"fmt"
	"log"
	"net"
	"strings"
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
func (s *Server) Run() error {
	// listen on the server's port
	portStr := fmt.Sprintf(":%v", s.port)
	address, err := net.ResolveUDPAddr("udp", portStr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	// reset the counters every second
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			s.accounts.Range(func(key string, acc Account) bool {
				acc.reset()
				return true
			})
		}
	}()

	// wait for messages of up to 128 bytes
	buffer := make([]byte, 128)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// clone buffer and send to goroutine to handle the message
		data := append([]byte(nil), buffer[:n]...)
		go func() {
			replyer := ReplyHandler{
				permit: func() {
					_, err := conn.WriteToUDP([]byte("p"), addr)
					if err != nil {
						log.Printf("failed to send permit response back to %v\n", addr)
					}
				},
				deny: func() {
					_, err := conn.WriteToUDP([]byte("d"), addr)
					if err != nil {
						log.Printf("failed to send deny response back to %v\n", addr)
					}
				},
			}

			s.handleUDPMessage(strings.TrimSpace(string(data)), replyer)
		}()
	}
}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleUDPMessage(str string, replyer ReplyHandler) {
	//  trim \n
	permitted := false
	var err error

	// deferred logging
	defer func() {
		log.Printf("message: %s permitted: %v err: %v", str, permitted, err)
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
