package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const defaultPort string = "8081"

func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // kill
	)
	defer stop()

	// look for override of default port using environment variable
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = defaultPort
	}

	// convert to string to integer
	slog.Info("Metrics")
	slog.Info("Listening on", "port", portStr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		slog.Error("Cannot parse PORT environment variable as integer", "error", err)
		os.Exit(1)
	}

	// start prometheus metrics
	slog.Info("Metrics")
	go RunMetrics(ctx)

	// run the server
	server := NewServer(port)
	server.Run(ctx)
	slog.Info("shutdown complete")
}
