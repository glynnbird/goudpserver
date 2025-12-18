package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RunMetrics(ctx context.Context) {

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsSrv := &http.Server{
		Addr:    ":2112",
		Handler: metricsMux,
	}

	go func() {
		slog.Info("metrics listening on :2112")
		if err := metricsSrv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				slog.Info("metrics server closed")
				return
			}
			slog.Error("metrics server", "error", err)
			return
		}
	}()

	<-ctx.Done()
	metricsSrv.Shutdown(ctx)
}
