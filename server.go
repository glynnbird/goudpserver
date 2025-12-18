package main

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// refreshInterval is how frequently the account's buckets are refreshed (topped up)
const refreshInterval = 1 * time.Second

// Server is a data structure that holds information about our UDP server, including which
// port it listens on and a map of Account structs, one for each user account
type Server struct {
	port              int
	accounts          *AccountMap
	wg                sync.WaitGroup
	messagesProcessed *prometheus.CounterVec
	messagesErrored   *prometheus.CounterVec
	messagesHandled   *prometheus.CounterVec
}

// a ReplyHandler is a struct which has two functions that reply permit or deny back to the
// caller. Passing a ReplyHandler to a goroutine saves having to pass connection and address
// pointers around.
type ReplyHandler struct {
	permit func()
	deny   func()
}

// NewServer creates a new server struct, given the port
func NewServer(port int) *Server {
	messagesProcessed := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goudpserver",
		Subsystem: "messages",
		Name:      "received",
		Help:      "Total number of messages received",
	}, []string{"protocol"})
	messagesErrored := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goudpserver",
		Subsystem: "messages",
		Name:      "errored",
		Help:      "Total number of messages errored",
	}, []string{"reason"})
	messagesHandled := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goudpserver",
		Subsystem: "messages",
		Name:      "handled",
		Help:      "Total number of messages handled",
	}, []string{"class", "permitted"})
	accountsPtr := NewAccountMap()
	server := Server{
		port:              port,
		accounts:          accountsPtr,
		messagesProcessed: messagesProcessed,
		messagesErrored:   messagesErrored,
		messagesHandled:   messagesHandled,
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

	// we have three goroutines to wait for: TCP/UDP servers and the reset tiner
	s.wg.Add(3)

	// run the UDP server
	go func() {
		var err error
		udpConn, err = s.listenUDPServer()
		if err != nil {
			slog.Error("UDP listen error", "error", err)
		}
		s.runUDPServer(udpConn)
	}()

	// run the TCP server
	go func() {
		var err error
		tcpListener, err = s.listenTCPServer()
		if err != nil {
			slog.Error("TCP listen error", "error", err)
		}
		s.runTCPServer(tcpListener)
	}()

	// reset the accounts every second
	go s.RunTimer(ctx)

	// wait until the app indicates it wants to terminate
	<-ctx.Done()

	// close the servers
	slog.Info("Terminating")
	tcpListener.Close()
	udpConn.Close()

	// wait for all goroutines to finish
	s.wg.Wait()
	slog.Info("goroutines stopped")
}

// handle is run as a goroutine to handle a single incoming message
func (s *Server) handleMessage(protocol string, str string, replyer ReplyHandler) {
	permitted := false
	var err error

	// deferred logging
	defer func() {
		if err != nil {
			slog.Error("Error handling message", "protocol", protocol, "error", err)
		} else {
			slog.Info("Message", "protocol", protocol, "message", str, "permitted", permitted)
		}
	}()

	// parse the incoming message
	s.messagesProcessed.WithLabelValues(protocol).Inc()
	message, err := parseMessage(str)
	if err != nil {
		s.messagesErrored.WithLabelValues(err.Error()).Inc()
		replyer.deny()
		return
	}

	// locate the account in the sync map (or create a new one if it's not there already)
	acc := s.accounts.LoadOrStore(message.accountName)

	// get a decision on whether there is enough Value left in the bucket to decrement it by "inc"
	permitted = acc.Buckets[message.class].dec(message.inc, message.capacity)

	// permit or deny reply
	if permitted {
		s.messagesHandled.WithLabelValues(message.class, "p").Inc()
		replyer.permit()
	} else {
		s.messagesHandled.WithLabelValues(message.class, "d").Inc()
		replyer.deny()
	}
}
