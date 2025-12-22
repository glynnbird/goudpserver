package main

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"
)

// refreshInterval is how frequently the account's buckets are refreshed (topped up)
const refreshInterval = 1 * time.Second

// responses
const permitResponse = "p"
const denyResponse = "d"

// Server is a data structure that holds information about our UDP server, including which
// port it listens on and a map of Account structs, one for each user account
type Server struct {
	port     int
	accounts *AccountMap
	wg       sync.WaitGroup
	met      *metrics
}

// NewServer creates a new server struct, given the port
func NewServer(port int, met *metrics) *Server {

	accountsPtr := NewAccountMap()
	server := Server{
		port:     port,
		accounts: accountsPtr,
		met:      met,
	}
	return &server
}

// RunTimer resets the accountMap's buckets every second
func (s *Server) RunTimer(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// loop until the context is done, i.e. the application is ready to quit
	for {
		select {
		case <-ticker.C:
			s.accounts.Reset()
			slog.Debug("Reset")
		case <-ctx.Done():
			return
		}
	}
}

// Run executes the server. It starts listening on the specified port,
// dispatching incoming messages to its own goroutine. Another goroutine
// resets each Account's buckets periodically.
func (s *Server) Run(ctx context.Context) {
	var udpConn *net.UDPConn
	var tcpListener net.Listener

	// we have four goroutines to wait for:
	//   - TCP server
	//   - UDP server
	//   - reset timer
	//   - prometheus metrics server

	// start prometheus metrics
	s.wg.Add(1)
	go s.runMetrics(ctx)

	// run the UDP server
	s.wg.Add(1)
	go func() {
		var err error
		udpConn, err = s.listenUDPServer()
		if err != nil {
			slog.Error("UDP listen error", "error", err)
		}
		s.runUDPServer(ctx, udpConn)
	}()

	// run the TCP server
	s.wg.Add(1)
	go func() {
		var err error
		tcpListener, err = s.listenTCPServer()
		if err != nil {
			slog.Error("TCP listen error", "error", err)
		}
		s.runTCPServer(ctx, tcpListener)
	}()

	// reset the accounts every second
	s.wg.Add(1)
	go s.RunTimer(ctx)

	// wait for all goroutines to finish
	s.wg.Wait()
	slog.Info("goroutines stopped")
}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleMessage(protocol string, str string) string {
	permitted := false
	var err error

	// parse the incoming message
	s.met.messagesProcessed.WithLabelValues(protocol).Inc()
	message, err := parseMessage(str)
	if err != nil {
		s.met.messagesErrored.WithLabelValues(err.Error()).Inc()
		slog.Error("Error handling message", "protocol", protocol, "error", err)
		return denyResponse
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	acc, newAccountCreated := s.accounts.LoadOrStore(message.accountName)
	if newAccountCreated {
		s.met.accountGauge.Inc()
	}

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	permitted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// permit or deny reply
	slog.Info("Message", "protocol", protocol, "message", str, "permitted", permitted)
	if permitted {
		s.met.messagesHandled.WithLabelValues(message.class, permitResponse).Inc()
		return permitResponse
	} else {
		s.met.messagesHandled.WithLabelValues(message.class, denyResponse).Inc()
		return denyResponse
	}
}
