package main

import (
	"sync"
	"testing"
)

// setup tests
func Test_server_new(t *testing.T) {
	port := 8888
	server := NewServer(port)
	if server.port != port {
		t.Errorf("Expected server port to be %v, got %v", port, server.port)
	}
	if len(server.accounts.accounts) != 0 {
		t.Errorf("Expected server account map to %v length, got %v", 0, len(server.accounts.accounts))
	}
}

func Test_server_permit_deny_count(t *testing.T) {
	port := 8888
	server := NewServer(port)
	permitCount := 0
	denyCount := 0
	replyHandler := ReplyHandler{
		permit: func() {
			permitCount++
		},
		deny: func() {
			denyCount++
		},
	}
	// try 2000 lookups/writes/queries for a bucket with a capacity of 10000
	for i := 0; i < 2000; i++ {
		server.handleMessage("test", "gb,l,1000,1", replyHandler)
		server.handleMessage("test", "gb,w,1000,1", replyHandler)
		server.handleMessage("test", "gb,q,1000,1", replyHandler)
	}
	if permitCount != 3000 {
		t.Errorf("Expected permit count to be %v, got %v", 1000, permitCount)
	}
	if denyCount != 3000 {
		t.Errorf("Expected deny count to be %v, got %v", 1000, denyCount)
	}

	// reset and try again
	server.accounts.Reset()
	permitCount = 0
	denyCount = 0

	// try 2000 writes for a bucket with a capacity of 10000
	for i := 0; i < 2000; i++ {
		server.handleMessage("test", "gb,l,1000,1", replyHandler)
		server.handleMessage("test", "gb,w,1000,1", replyHandler)
		server.handleMessage("test", "gb,q,1000,1", replyHandler)
	}
	if permitCount != 3000 {
		t.Errorf("Expected permit count to be %v, got %v", 1000, permitCount)
	}
	if denyCount != 3000 {
		t.Errorf("Expected deny count to be %v, got %v", 1000, denyCount)
	}
}

func Test_server_permit_deny_count_parallel(t *testing.T) {
	port := 8888
	server := NewServer(port)
	permitCount := 0
	denyCount := 0
	var muPermit sync.Mutex
	var muDeny sync.Mutex
	replyHandler := ReplyHandler{
		permit: func() {
			muPermit.Lock()
			defer muPermit.Unlock()
			permitCount++
		},
		deny: func() {
			muDeny.Lock()
			defer muDeny.Unlock()
			denyCount++
		},
	}

	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			server.handleMessage("test", "gb,l,50000,1", replyHandler)
		}
	}()
	wg.Wait()

	if permitCount != 50000 {
		t.Errorf("Expected permit count to be %v, got %v", 50000, permitCount)
	}
	if denyCount != 100000 {
		t.Errorf("Expected deny count to be %v, got %v", 100000, denyCount)
	}
}
