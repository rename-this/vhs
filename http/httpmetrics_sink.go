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
	//More here later.
	c       *Correlator
	latency *prometheus.GaugeVec
}

// NewHttpmetrics returns a new Httpmetrics.
func NewHttpmetrics(reqTimeout time.Duration) *Httpmetrics {
	return &Httpmetrics{
		c: NewCorrelator(reqTimeout),
		latency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vhs_http_latency_seconds",
			Help: "Latency of http exchanges captured by VHS.",
		}, []string{"method", "code"}),
	}
}

// Init stub
func (h *Httpmetrics) Init(ctx context.Context) {
	prometheus.MustRegister(h.latency)

	go h.c.Start(ctx)

	go func() {
		for req := range h.c.Exchanges {
			h.calcMetrics(req)
		}
	}()
}

// Write stub
func (h *Httpmetrics) Write(n interface{}) error {
	switch m := n.(type) {
	case Message:
		h.c.Messages <- m
	}
	return nil
}

// Flush stub
func (*Httpmetrics) Flush() error {
	return nil
}

// Metric calculation stub
func (h *Httpmetrics) calcMetrics(req *Request) {
	if req.Response != nil {
		//latency := req.Response.Created.Sub(req.Created).Seconds()
		//fmt.Printf("%v:%d latency: %ds\n", req.ConnectionID, req.ExchangeID, latency.Milliseconds())
		h.latency.With(prometheus.Labels{
			"method": req.Method,
			"code":   fmt.Sprintf("%d", req.Response.StatusCode),
		}).Set(req.Response.Created.Sub(req.Created).Seconds())
	} else {
		//fmt.Println("No response?!")
	}
}
