package httpx

import (
	"testing"
	"time"

	"github.com/gramLabs/vhs/session"

	"gotest.tools/v3/assert"
)

var refTime = time.Date(2020, 10, 12, 12, 0, 0, 0, time.UTC)

func TestMetrics(t *testing.T) {
	cases := []struct {
		desc     string
		messages []Message

		counters  map[metricsLabels]int64
		durations map[metricsLabels][]float64
	}{
		{
			desc: "latency only",
			messages: []Message{
				&Request{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime,
					Method:       "POST",
					URL:          newURL("/test"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   200,
				},
			},
			// duration:  0.5,
			durations: map[metricsLabels][]float64{
				metricsLabels{
					method: "POST",
					code:   "200",
					path:   "/test",
				}: {0.5},
			},
			// count: 1
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: "POST",
					code:   "200",
					path:   "/test",
				}: 1,
			},
		},
		{
			desc: "timeout only",
			messages: []Message{
				&Request{
					ConnectionID: "222",
					ExchangeID:   "1",
					Created:      refTime.Add(750 * time.Millisecond),
					Method:       "GET",
					Response:     nil,
					URL:          newURL("/test"),
				},
			},
			// duration: none
			durations: map[metricsLabels][]float64{},
			// count: one timeout
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: "GET",
					code:   "",
					path:   "/test",
				}: 1,
			},
		},
		{
			desc: "combined",
			messages: []Message{
				&Request{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime,
					Method:       "POST",
					URL:          newURL("/test1"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   200,
				},
				&Request{
					ConnectionID: "222",
					ExchangeID:   "1",
					Created:      refTime.Add(750 * time.Millisecond),
					Method:       "GET",
					Response:     nil,
					URL:          newURL("/test2"),
				},
			},
			// duration: one measurement, 0.5s
			durations: map[metricsLabels][]float64{
				metricsLabels{
					method: "POST",
					code:   "200",
					path:   "/test1",
				}: {0.5},
			},
			// count: 1 timeout, 1 code 200
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: "POST",
					code:   "200",
					path:   "/test1",
				}: 1,
				metricsLabels{
					method: "GET",
					code:   "",
					path:   "/test2",
				}: 1,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx, _, _ := session.NewContexts(&session.Config{
				HTTPTimeout: 50 * time.Millisecond, // Short correlator time so we can actually get some timeouts.
			}, nil)

			backend := newTestMetricsBackend()
			metrics := &Metrics{
				in:  make(chan interface{}),
				met: backend,
			}
			go metrics.Init(ctx, nil)

			for _, m := range c.messages {
				metrics.In() <- m
				time.Sleep(10 * time.Millisecond)
			}

			// Sleep long enough to let a prune cycle run in the correlator.
			time.Sleep(300 * time.Millisecond)

			ctx.Cancel()

			assert.DeepEqual(t, c.durations, backend.durations)
			assert.DeepEqual(t, c.counters, backend.counters)

		})
	}
}

// testMetricsBackend is a metricsBackend for testing
type testMetricsBackend struct {
	counters  map[metricsLabels]int64
	durations map[metricsLabels][]float64
}

// Make sure testMetricsBackend implements metricsBackend
var _ metricsBackend = &testMetricsBackend{}

// newTestMetricsBackend creates a new test metrics backend.
func newTestMetricsBackend() *testMetricsBackend {
	return &testMetricsBackend{
		counters:  make(map[metricsLabels]int64),
		durations: make(map[metricsLabels][]float64),
	}
}

// IncrementCounter increments the http exchange counter with the specified labels.
func (t *testMetricsBackend) IncrementCounter(l metricsLabels) {
	t.counters[l]++
}

// AddDuration adds the measured http exchange duration with the specified labels.
func (t *testMetricsBackend) AddDuration(l metricsLabels, f float64) {
	t.durations[l] = append(t.durations[l], f)
}
