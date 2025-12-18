package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
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

	// register the tcpRequestDuration histogram to report on request handling performance
	udpRequestDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "goudpserver",
			Subsystem: "udp_server",
			Name:      "request_duration_seconds",
			Help:      "Time spent processing a UDP request.",
			Buckets:   []float64{0.00001, 0.00002, 0.00003, 0.00004, 0.00005, 0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
		},
	)
	prometheus.MustRegister(udpRequestDuration)

	// wait for messages of up to 128 bytes
	buffer := make([]byte, 128)
	for {

		n, addr, err := conn.ReadFromUDP(buffer)
		timer := prometheus.NewTimer(udpRequestDuration)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.Info("UDP server closed")
				return // graceful shutdown
			}
			slog.Error("UDP read error", "error", err)
			continue
		}

		// clone buffer and send to goroutine to handle the message
		data := make([]byte, n)
		copy(data, buffer[:n])
		go func(t *prometheus.Timer) {
			// gives a means of replying back to the caller to handleMessage
			replyHandler := ReplyHandler{
				permit: func() {
					_, err := conn.WriteToUDP([]byte("p"), addr)
					t.ObserveDuration()
					if err != nil {
						slog.Error("UDP failed to send permit response", "addr", addr, "error", err)
					}
				},
				deny: func() {
					_, err := conn.WriteToUDP([]byte("d"), addr)
					t.ObserveDuration()
					if err != nil {
						slog.Error("UDP failed to send deny response", "addr", addr, "error", err)
					}
				},
			}

			// parse the message and reply back to the caller
			s.handleMessage("UDP", strings.TrimSpace(string(data)), replyHandler)
		}(timer)
	}
}
