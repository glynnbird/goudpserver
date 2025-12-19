package main

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// metrics collects together all the prometheus metrics in one place
type metrics struct {
	accountGauge       prometheus.Gauge
	messagesProcessed  *prometheus.CounterVec
	messagesErrored    *prometheus.CounterVec
	messagesHandled    *prometheus.CounterVec
	udpRequestDuration prometheus.Histogram
	tcpRequestDuration prometheus.Histogram
	socketsGauge       prometheus.Gauge
}

var (
	once             sync.Once
	metricsSingleton *metrics
)

// NewMetrics creates a metrics struct and returns a point to a "singleton", to prevent
// multiple invocations re-registering metrics.
func NewMetrics() *metrics {

	once.Do(func() {
		m := metrics{}
		metricsSingleton = &m
		m.messagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goudpserver",
			Subsystem: "messages",
			Name:      "received",
			Help:      "Total number of messages received",
		}, []string{"protocol"})
		m.messagesErrored = promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goudpserver",
			Subsystem: "messages",
			Name:      "errored",
			Help:      "Total number of messages errored",
		}, []string{"reason"})
		m.messagesHandled = promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goudpserver",
			Subsystem: "messages",
			Name:      "handled",
			Help:      "Total number of messages handled",
		}, []string{"class", "permitted"})
		m.accountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goudpserver",
			Subsystem: "account_map",
			Name:      "num_keys",
			Help:      "Number of entries in our AccountMap",
		})
		m.udpRequestDuration = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "goudpserver",
				Subsystem: "udp_server",
				Name:      "request_duration_seconds",
				Help:      "Time spent processing a UDP request.",
				Buckets:   []float64{0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
			},
		)
		m.tcpRequestDuration = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "goudpserver",
				Subsystem: "tcp_server",
				Name:      "request_duration_seconds",
				Help:      "Time spent processing a TCP request.",
				Buckets:   []float64{0.0001, 0.0002, 0.0003, 0.0004, 0.0005},
			},
		)
		m.socketsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goudpserver",
			Subsystem: "tcp_server",
			Name:      "num_sockets",
			Help:      "Number of sockets open in the TCP server",
		})
		prometheus.MustRegister(
			m.accountGauge,
			m.udpRequestDuration,
			m.tcpRequestDuration,
			m.socketsGauge)
	})

	return metricsSingleton
}
