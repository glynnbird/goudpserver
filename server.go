package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const refreshInterval = 1 * time.Second

// Message is a parsed incoming message
type Message struct {
	accountName string
	class       string
	capacity    int
	inc         int
}

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
		log.Printf("failed to send deny respose back to %v\n", addr)
	}
}

// permit sends a "p" (permit) message back to the caller
func (s *Server) permit(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("p"), addr)
	if err != nil {
		log.Printf("failed to send deny respose back to %v\n", addr)
	}
}

// parseMessage takes an incoming UDP message string and parses it looking for
// <accountName>,<class>,<capacity>,<inc>\n
// where accountName that uniquely identifies each client, class is l/w/q,
// capacity is the bucket capacity for that class/accountName and inc is
// the amount that is being asked to be removed from the bucket value.
func (s *Server) parseMessage(str string) (*Message, error) {
	// parse the incoming string - account,class,max_per_second,inc_by
	bits := strings.Split(str, ",")
	if len(bits) != 4 {
		return nil, errors.New("message string must contain 4 strings separated by commas")
	}

	// sanity checks
	accountName := bits[0]
	class := bits[1]
	capacityStr := bits[2]
	incrementStr := bits[3]
	if len(accountName) == 0 || len(class) == 0 || len(capacityStr) == 0 || len(incrementStr) == 0 {
		return nil, errors.New("missing account/class/capacity/inc strings")
	}
	if class != "l" && class != "w" && class != "q" {
		return nil, errors.New("class must be l/w/q")
	}
	capacity, err := strconv.Atoi(capacityStr)
	if err != nil {
		return nil, errors.New("cannot convert capacity from string to integer")
	}
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	inc, err := strconv.Atoi(incrementStr)
	if err != nil {
		return nil, errors.New("cannot convert increment from string to integer")
	}
	if inc <= 0 {
		return nil, errors.New("inc must be positive")
	}
	message := Message{
		accountName: accountName,
		class:       class,
		capacity:    capacity,
		inc:         inc,
	}
	return &message, nil

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
	message, err := s.parseMessage(str)
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
