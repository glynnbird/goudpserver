package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

// runUDPServer executes a UDP server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine.
func (s *Server) runUDPServer() error {
	defer s.wg.Done()

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

	// wait for messages of up to 128 bytes
	buffer := make([]byte, 128)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("UDP Error: %v\n", err)
			continue
		}

		// clone buffer and send to goroutine to handle the message
		data := append([]byte(nil), buffer[:n]...)
		go func() {
			// gives a means of replying back to the caller to handleMessage
			replyHandler := ReplyHandler{
				permit: func() {
					_, err := conn.WriteToUDP([]byte("p"), addr)
					if err != nil {
						log.Printf("UDP failed to send permit response back to %v\n", addr)
					}
				},
				deny: func() {
					_, err := conn.WriteToUDP([]byte("d"), addr)
					if err != nil {
						log.Printf("UDP failed to send deny response back to %v\n", addr)
					}
				},
			}

			// parse the message and reply back to the caller
			s.handleMessage("UDP", strings.TrimSpace(string(data)), replyHandler)
		}()
	}
}
