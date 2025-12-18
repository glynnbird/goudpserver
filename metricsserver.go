package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) runMetrics(ctx context.Context) {
	defer s.wg.Done()

	// register prometheus metrics
	// register the udpRequestDuration histogram to report on request handling performance
	s.udpRequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "goudpserver",
			Subsystem: "udp_server",
			Name:      "request_duration_seconds",
			Help:      "Time spent processing a UDP request.",
			Buckets:   []float64{0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
		},
	)

	// register the tcpRequestDuration histogram to report on request handling performance
	s.tcpRequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "goudpserver",
			Subsystem: "tcp_server",
			Name:      "request_duration_seconds",
			Help:      "Time spent processing a TCP request.",
			Buckets:   []float64{0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
		},
	)

	// register the socketsGauge to count the number active sockets
	s.socketsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goudpserver",
		Subsystem: "tcp_server",
		Name:      "num_sockets",
		Help:      "Number of sockets open in the TCP server",
	})
	prometheus.MustRegister(s.udpRequestDuration, s.tcpRequestDuration, s.socketsGauge)

	// create http server
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
