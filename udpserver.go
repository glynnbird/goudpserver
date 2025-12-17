package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
)

func (s *Server) listenUDPServer() (*net.UDPConn, error) {
	// listen on the server's port
	portStr := fmt.Sprintf(":%v", s.port)
	address, err := net.ResolveUDPAddr("udp", portStr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// runUDPServer executes a UDP server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine.
func (s *Server) runUDPServer(conn *net.UDPConn) {
	defer s.wg.Done()

	// wait for messages of up to 128 bytes
	buffer := make([]byte, 128)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.Info("UDP server closed")
				return // graceful shutdown
			}
			slog.Error("UDP read error", "error", err)
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
						slog.Error("UDP failed to send permit response back", "addr", addr, "error", err)
					}
				},
				deny: func() {
					_, err := conn.WriteToUDP([]byte("d"), addr)
					if err != nil {
						slog.Error("UDP failed to send deny response back", "addr", addr, "error", err)
					}
				},
			}

			// parse the message and reply back to the caller
			s.handleMessage("UDP", strings.TrimSpace(string(data)), replyHandler)
		}()
	}
}
