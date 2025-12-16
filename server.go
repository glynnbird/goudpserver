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
		go s.handleUDPMessage(conn, addr, data)
	}
}

// deny sends a "d" (deny) message back to the caller
func (s *Server) deny(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("d"), addr)
	if err != nil {
		log.Printf("failed to send deny response back to %v\n", addr)
	}
}

// permit sends a "p" (permit) message back to the caller
func (s *Server) permit(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("p"), addr)
	if err != nil {
		log.Printf("failed to send deny response back to %v\n", addr)
	}
}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleUDPMessage(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	//  trim \n
	accepted := false
	str := strings.TrimSpace(string(data))
	var err error

	// deferred logging
	defer func() {
		log.Printf("addr: %s message: %s accepted: %v err: %v", addr, str, accepted, err)
	}()

	// parse the incoming message
	message, err := parseMessage(str)
	if err != nil {
		s.deny(conn, addr)
		return
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	acc, ok := s.accounts.Load(message.accountName)

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	accepted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// if this is a new account, it needs storing in the sync map
	if !ok {
		s.accounts.Store(message.accountName, acc)
	}

	// permit or deny reply
	if accepted {
		s.permit(conn, addr)
	} else {
		s.deny(conn, addr)
	}
}
