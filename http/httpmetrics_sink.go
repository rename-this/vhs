package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gramLabs/vhs/sink"
	"github.com/prometheus/client_golang/prometheus"
)

// Ensure Httpmetrics conforms to Sink interface.
var _ sink.Sink = &Httpmetrics{}

// Httpmetrics is a sink that calculates HTTP metrics to be made available to Prometheus
type Httpmetrics struct {
	c        *Correlator
	latency  *prometheus.GaugeVec
	timeouts *prometheus.CounterVec
}

// NewHttpmetrics returns a new Httpmetrics.
func NewHttpmetrics(reqTimeout time.Duration) *Httpmetrics {
	return &Httpmetrics{
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
func (h *Httpmetrics) Init(ctx context.Context) {
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
func (h *Httpmetrics) Write(n interface{}) error {
	switch m := n.(type) {
	case Message:
		h.c.Messages <- m
	}
	return nil
}

// Flush is a no-op.
func (*Httpmetrics) Flush() error { return nil }

// Calculates the desired metrics. Currently calculates latency between request and response and number of timeouts.
func (h *Httpmetrics) calcMetrics(req *Request) {
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
