package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

// runTCPServer executes a TCP server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine.
func (s *Server) runTCPServer() error {
	defer s.wg.Done()

	// listen on the server's port
	portStr := fmt.Sprintf(":%v", s.port)
	ln, err := net.Listen("tcp", portStr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		// accept TCP connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println("TCP accept error:", err)
			continue
		}

		// one go routine per connection
		go func() {
			defer conn.Close()
			// time out the socket after 30 seconds of inactivity
			idleTimeout := 30 * time.Second

			// create line reader
			reader := bufio.NewScanner(conn)
			reader.Buffer(make([]byte, 0, 1024), 1024)

			// read each line
			for reader.Scan() {
				conn.SetReadDeadline(time.Now().Add(idleTimeout))
				line := reader.Text()
				// gives a means of replying back to the caller to handleMessage
				replyHandler := ReplyHandler{
					permit: func() {
						_, err := conn.Write([]byte("p\n"))
						if err != nil {
							log.Printf("TCP failed to send permit response\n")
						}
					},
					deny: func() {
						_, err := conn.Write([]byte("d\n"))
						if err != nil {
							log.Printf("TCP failed to send deny response\n")
						}
					},
				}

				// parse the message and reply back to the caller
				s.handleMessage("TCP", line, replyHandler)
			}
		}()
	}
}
