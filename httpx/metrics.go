package httpx

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Ensure Metrics conforms to Format interface.
var _ flow.OutputFormat = &Metrics{}

// Metrics is a format that calculates HTTP metrics for Prometheus monitoring
// Note that this format does not modify data passing through it, it merely extracts metrics.
// Also note that this is a "dead end" format: its output io.Reader is never updated and remains empty.
type Metrics struct {
	c  *Correlator
	in chan interface{}

	met metricsBackend
}

// NewMetricsOutput creates a new *output.Pipe for calculating HTTP metrics
func NewMetricsOutput() *flow.Output {
	return flow.NewOutput(NewMetrics(), nil, ioutilx.NopWriteCloser(ioutil.Discard))
}

// NewMetrics creates a new Metrics format.
func NewMetrics() *Metrics {
	return &Metrics{
		in: make(chan interface{}),

		met: &promBackend{
			Count: promauto.NewCounterVec(prometheus.CounterOpts{
				Namespace: "vhs",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total count of http requests captured by VHS.",
			}, []string{"method", "code", "path"}),

			Duration: promauto.NewSummaryVec(prometheus.SummaryOpts{
				Namespace: "vhs",
				Subsystem: "http",
				Name:      "duration_seconds",
				Help:      "Duration of http exchanges seen by VHS.",
				Objectives: map[float64]float64{
					0.5:    0.05,
					0.75:   0.01,
					0.9:    0.005,
					0.95:   0.005,
					0.99:   0.001,
					0.999:  0.0001,
					0.9999: 0.00001,
				},
			}, []string{"method", "code", "path"}),
		},
	}
}

// In returns the input channel.
func (m *Metrics) In() chan<- interface{} { return m.in }

// Init initializes the metrics format and registers the metrics with Prometheus
func (m *Metrics) Init(ctx session.Context, _ io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "http_metrics").
		Logger()

	ctx.Logger.Debug().Msg("init")

	c := NewCorrelator(ctx.Config.HTTPTimeout)
	c.Start(ctx)

	ctx.Logger.Debug().Msg("correlator started")

	for {
		select {
		case n := <-m.in:
			switch msg := n.(type) {
			case Message:
				c.Messages <- msg
				if ctx.Config.DebugHTTPMessages {
					ctx.Logger.Debug().Interface("m", msg).Msg("received message")
				} else {
					ctx.Logger.Debug().Msg("received message")
				}
			}
		case r := <-c.Exchanges:
			calcMetrics(ctx, r, m.met)
			if ctx.Config.DebugHTTPMessages {
				ctx.Logger.Debug().Interface("r", r).Msg("received request from correlator")
			} else {
				ctx.Logger.Debug().Msg("received request from correlator")
			}
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}

// Calculates the desired metrics. Currently calculates latency between request and response and number of timeouts.
func calcMetrics(ctx session.Context, req *Request, met metricsBackend) {
	if req.Response == nil {
		met.IncrementCounter(metricsLabels{
			method: req.Method,
			code:   "",
			path:   req.URL.Path,
		})
		return
	}

	met.IncrementCounter(metricsLabels{
		method: req.Method,
		code:   fmt.Sprintf("%d", req.Response.StatusCode),
		path:   req.URL.Path,
	})

	met.AddDuration(metricsLabels{
		method: req.Method,
		code:   fmt.Sprintf("%d", req.Response.StatusCode),
		path:   req.URL.Path,
	}, req.Response.Created.Sub(req.Created).Seconds())

	ctx.Logger.Debug().Msg("calculated")
}

// metricsBackend abstracts the basic metric update calls. Mostly for easier testing.
type metricsBackend interface {
	IncrementCounter(metricsLabels)
	AddDuration(metricsLabels, float64)
}

// metricsLabels is the set of labels we're currently using
type metricsLabels struct {
	method string
	code   string
	path   string
}

// promBackend wraps a few details of the interface to Prometheus client code.
type promBackend struct {
	Count    *prometheus.CounterVec
	Duration *prometheus.SummaryVec
}

// Make sure promBackend implements metricsBackend
var _ metricsBackend = &promBackend{}

// IncrementCounter increments the prometheus HTTP message counter associated with the specified labels.
func (p *promBackend) IncrementCounter(labels metricsLabels) {
	p.Count.With(prometheus.Labels{
		"method": labels.method,
		"code":   labels.code,
		"path":   labels.path,
	}).Inc()
}

// AddDuration adds the measured HTTP duration to the prometheus summary with the specified labels.
func (p *promBackend) AddDuration(labels metricsLabels, duration float64) {
	p.Duration.With(prometheus.Labels{
		"method": labels.method,
		"code":   labels.code,
		"path":   labels.path,
	}).Observe(duration)
}
