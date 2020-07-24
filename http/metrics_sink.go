package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gramLabs/vhs/sink"
	"github.com/prometheus/client_golang/prometheus"
)

// Ensure MetricsSink conforms to Sink interface.
var _ sink.Sink = &MetricsSink{}

// MetricsSink is a sink that calculates HTTP metrics to be made available to Prometheus
type MetricsSink struct {
	c        *Correlator
	latency  *prometheus.GaugeVec
	timeouts *prometheus.CounterVec
}

// NewMetrics returns a new MetricsSink.
func NewMetrics(reqTimeout time.Duration) *MetricsSink {
	return &MetricsSink{
		c: NewCorrelator(reqTimeout),
		latency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vhs_http_latency_seconds",
			Help: "Latency of http exchanges captured by VHS.",
		}, []string{"method", "code"}),
		timeouts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "vhs_http_timeouts_total",
			Help: "Total count of timed-out http exchanges captured by VHS.",
		}, []string{"method"}),
	}
}

// Init initializes the sink and registers the Prometheus metrics
func (h *MetricsSink) Init(ctx context.Context) {
	prometheus.MustRegister(h.latency)
	prometheus.MustRegister(h.timeouts)

	go h.c.Start(ctx)

	go func() {
		for req := range h.c.Exchanges {
			h.calcMetrics(req)
		}
	}()
}

// Write adds a new message on which metrics will be calculated.
func (h *MetricsSink) Write(n interface{}) error {
	switch m := n.(type) {
	case Message:
		h.c.Messages <- m
	}
	return nil
}

// Flush is a no-op.
func (*MetricsSink) Flush() error { return nil }

// Calculates the desired metrics. Currently calculates latency between request and response and number of timeouts.
func (h *MetricsSink) calcMetrics(req *Request) {
	if req.Response != nil {
		h.latency.With(prometheus.Labels{
			"method": req.Method,
			"code":   fmt.Sprintf("%d", req.Response.StatusCode),
		}).Set(req.Response.Created.Sub(req.Created).Seconds())

	} else {
		h.timeouts.With(prometheus.Labels{
			"method": req.Method,
		}).Inc()
	}
}
