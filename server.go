package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// Server is a data structure that holds information about our UDP server, including which
// port it listens on and a map of Account structs, one for each user account
type Server struct {
	port     int
	accounts *AccountMap
}

// NewServer createa a new server struct, given the port
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
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			s.accounts.Range(func(key string, acc Account) bool {
				acc.reset()
				s.accounts.Store(key, acc)
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
	conn.WriteToUDP([]byte("d"), addr)
}

// deny sends a "p" (permit) message back to the caller
func (s *Server) permit(conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("p"), addr)
}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleUDPMessage(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	accepted := false
	str := strings.TrimSpace(string(data))
	defer func() {
		log.Printf("addr: %s message: %s accepted: %v", addr, str, accepted)
	}()
	// parse the incoming string - account,class,max_per_second,inc_by
	bits := strings.Split(str, ",")
	if len(bits) != 4 {
		return
	}

	// sanity checks
	account := bits[0]
	class := bits[1]
	max_per_second := bits[2]
	incrementStr := bits[3]
	if len(account) == 0 || len(class) == 0 || len(max_per_second) == 0 ||
		(class != "l" && class != "w" && class != "q") {
		s.deny(conn, addr)
		return
	}
	limit, err := strconv.Atoi(max_per_second)
	if err != nil {
		s.deny(conn, addr)
		return
	}
	inc, err := strconv.Atoi(incrementStr)
	if err != nil || inc <= 0 {
		s.deny(conn, addr)
		return
	}

	// check the map
	acc, _ := s.accounts.Load(account)
	switch class {
	case "l":
		accepted = acc.Lookups.dec(inc, limit)
	case "w":
		accepted = acc.Writes.dec(inc, limit)
	case "q":
		accepted = acc.Queries.dec(inc, limit)
	}
	s.accounts.Store(account, acc)

	// jsonStr, _ := json.Marshal(acc)

	// log.Printf("%v\n", string(jsonStr))
	if accepted {
		s.permit(conn, addr)
	} else {
		s.deny(conn, addr)
	}

}
