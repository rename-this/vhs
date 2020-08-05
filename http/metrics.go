package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/gramLabs/vhs/output"
	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/modifier"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Ensure Metrics conforms to Format interface.
var _ format.Format = &Metrics{}

// Metrics is a format that calculates HTTP metrics for Prometheus monitoring
// Note that this format does not modify data passing through it, it merely extracts metrics.
// Also note that this is a "dead end" format: its output io.Reader is never updated and remains empty.
type Metrics struct {
	c        *Correlator
	latency  *prometheus.GaugeVec
	timeouts *prometheus.CounterVec
	in chan interface{}
	out chan io.Reader
	errs chan error
}

// NewMetricsPipe creates a new *output.Pipe for calculating HTTP metrics
func NewMetricsPipe(reqTimeout time.Duration) *output.Pipe {
	return output.NewPipe(NewMetrics(reqTimeout), nil, modifier.NopWriteCloser(ioutil.Discard))
}

// NewMetrics creates a new Metrics format.
func NewMetrics(reqTimeout time.Duration) *Metrics {
	return &Metrics{
		c: NewCorrelator(reqTimeout),
		latency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "vhs",
			Subsystem: "http",
			Name: "latency_seconds",
			Help: "Latency of http exchanges captured by VHS.",
		}, []string{"method", "code"}),
		timeouts: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "vhs",
			Subsystem: "http",
			Name: "timeouts_total",
			Help: "Total count of timed-out http exchanges captured by VHS.",
		}, []string{"method"}),
		in: make(chan interface{}),
		out: make(chan io.Reader),
		errs: make(chan error),
	}
}

// In returns the input channel.
func (m *Metrics) In() chan<- interface{} { return m.in }

// Errors returns the errors channel.
func (m *Metrics) Errors() <-chan error { return m.errs }

// Init initializes the metrics format and registers the metrics with Prometheus
func (m *Metrics) Init(ctx context.Context, _ io.Writer) {
	go m.c.Start(ctx)

	for {
		select {
		case n := <-m.in:
			switch msg := n.(type) {
			case Message:
				m.c.Messages <- msg
			}
		case r := <-m.c.Exchanges:
			m.calcMetrics(r)
		case <-ctx.Done():
			// (ztreinhart): return something downstream?
			return
		}
	}
}

// Calculates the desired metrics. Currently calculates latency between request and response and number of timeouts.
func (m *Metrics) calcMetrics(req *Request) {
	if req.Response != nil {
		m.latency.With(prometheus.Labels{
			"method": req.Method,
			"code":   fmt.Sprintf("%d", req.Response.StatusCode),
		}).Set(req.Response.Created.Sub(req.Created).Seconds())
	} else {
		m.timeouts.With(prometheus.Labels{
			"method": req.Method,
		}).Inc()
	}
}
