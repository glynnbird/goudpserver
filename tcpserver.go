package main

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// listenTCPServer creates a TCP listener on the server's port
func (s *Server) listenTCPServer() (net.Listener, error) {

	// listen on the server's port
	portStr := fmt.Sprintf(":%v", s.port)
	ln, err := net.Listen("tcp", portStr)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

// runTCPServer executes a TCP server. It takes an already-started network
// listener. It accepts socket connections and sets up a go-routine per
// socket to handle incoming messages. Each socket times out after a period
// of inactivity.
func (s *Server) runTCPServer(ln net.Listener) {
	defer s.wg.Done()

	// register the socketsGauge to count the number active sockets
	socketsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goudpserver",
		Subsystem: "tcp_server",
		Name:      "num_sockets",
		Help:      "Number of sockets open in the TCP server",
	})

	// register the tcpRequestDuration histogram to report on request handling performance
	tcpRequestDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "goudpserver",
			Subsystem: "tcp_server",
			Name:      "request_duration_seconds",
			Help:      "Time spent processing a TCP request.",
			Buckets:   []float64{0.00001, 0.00002, 0.00003, 0.00004, 0.00005, 0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
		},
	)
	prometheus.MustRegister(socketsGauge, tcpRequestDuration)

	for {
		// accept TCP connection
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.Info("TCP server closed")
				return // graceful shutdown
			}
			slog.Error("TCP accept error", "error", err)
			continue
		}

		// one go routine per connection
		go func() {
			defer conn.Close()
			defer socketsGauge.Dec()

			// increment socket count
			socketsGauge.Inc()

			// time out the socket after 30 seconds of inactivity
			idleTimeout := 30 * time.Second

			// create line reader
			reader := bufio.NewScanner(conn)
			reader.Buffer(make([]byte, 0, 1024), 1024)
			conn.SetDeadline(time.Now().Add(idleTimeout))

			// read each line
			for reader.Scan() {
				timer := prometheus.NewTimer(tcpRequestDuration)
				conn.SetReadDeadline(time.Now().Add(idleTimeout))
				line := reader.Text()
				// gives a means of replying back to the caller to handleMessage
				replyHandler := ReplyHandler{
					permit: func() {
						_, err := conn.Write([]byte("p\n"))
						timer.ObserveDuration()
						if err != nil {
							slog.Error("TCP failed to send permit response", "error", err)
						}
					},
					deny: func() {
						_, err := conn.Write([]byte("d\n"))
						timer.ObserveDuration()
						if err != nil {
							slog.Error("TCP failed to send deny response", "error", err)
						}
					},
				}

				// parse the message and reply back to the caller
				s.handleMessage("TCP", line, replyHandler)
			}
		}()
	}
}
