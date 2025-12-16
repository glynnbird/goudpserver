package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// runTCPServer executes a TCP server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine.
func (s *Server) runTCPServer() error {
	// listen on the server's port
	portStr := fmt.Sprintf(":%v", s.port)
	ln, err := net.Listen("tcp", portStr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("TCP accept error:", err)
			continue
		}

		go func() {
			defer conn.Close()
			reader := bufio.NewScanner(conn)
			reader.Buffer(make([]byte, 0, 1024), 1024)
			for reader.Scan() {
				line := reader.Text()
				// gives a means of replying back to the caller to handleMessage
				replyHandler := ReplyHandler{
					permit: func() {
						_, err := conn.Write([]byte("p"))
						if err != nil {
							log.Printf("TCP failed to send permit response\n")
						}
					},
					deny: func() {
						_, err := conn.Write([]byte("d"))
						if err != nil {
							log.Printf("TCP failed to send deny response back to \n")
						}
					},
				}

				// parse the message and reply back to the caller
				s.handleMessage("TCP", line, replyHandler)
			}
		}()
	}
}
