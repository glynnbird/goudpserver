package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Replace the default logger for tests
	slog.SetDefault(slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError, // suppress Info, Debug, Warn
		}),
	))

	os.Exit(m.Run())
}
