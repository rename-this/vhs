package httpx

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/rename-this/vhs/session"

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
					Method:       http.MethodPost,
					URL:          newURL("/test"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   http.StatusOK,
				},
			},
			// duration:  0.5,
			durations: map[metricsLabels][]float64{
				metricsLabels{
					method: http.MethodPost,
					code:   strconv.Itoa(http.StatusOK),
					path:   "/test",
				}: {0.5},
			},
			// count: 1
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: http.MethodPost,
					code:   strconv.Itoa(http.StatusOK),
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
					Method:       http.MethodGet,
					Response:     nil,
					URL:          newURL("/test"),
				},
			},
			// duration: none
			durations: map[metricsLabels][]float64{},
			// count: one timeout
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: http.MethodGet,
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
					Method:       http.MethodPost,
					URL:          newURL("/test1"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   http.StatusOK,
				},
				&Request{
					ConnectionID: "222",
					ExchangeID:   "1",
					Created:      refTime.Add(750 * time.Millisecond),
					Method:       http.MethodGet,
					Response:     nil,
					URL:          newURL("/test2"),
				},
			},
			// duration: one measurement, 0.5s
			durations: map[metricsLabels][]float64{
				metricsLabels{
					method: http.MethodPost,
					code:   strconv.Itoa(http.StatusOK),
					path:   "/test1",
				}: {0.5},
			},
			// count: 1 timeout, 1 code 200
			counters: map[metricsLabels]int64{
				metricsLabels{
					method: http.MethodPost,
					code:   strconv.Itoa(http.StatusOK),
					path:   "/test1",
				}: 1,
				metricsLabels{
					method: http.MethodGet,
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

func TestStressMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.") // Can be skipped if too time consuming.
	}

	var (
		httptimeout = 500 * time.Millisecond
		testTimeout = 10 * time.Second
		numSenders  = 4

		numExchangesSent int64 = 0
		numTimeoutsSent  int64 = 0
		ExSentMutex      sync.Mutex
		TimeoutSentMutex sync.Mutex
	)

	errs := make(chan error)
	genctx, metricsctx, _ := session.NewContexts(&session.Config{
		Debug:             false,
		DebugHTTPMessages: false,
		HTTPTimeout:       50 * time.Millisecond, // Short correlator time so we can actually get some timeouts.
	}, errs)

	backend := newTestMetricsBackend()
	metrics := &Metrics{
		in:  make(chan interface{}),
		met: backend,
	}
	go metrics.Init(metricsctx, nil)

	for i := 0; i < numSenders; i++ {
		go func() {
			for {
				select {
				case <-genctx.StdContext.Done():
					return
				default:
					connID := ksuid.New().String()
					eID := ksuid.New().String()
					timeout := rand.Intn(2)

					if timeout == 0 {
						metrics.In() <- &Request{
							ConnectionID: connID,
							ExchangeID:   eID,
							Created:      refTime,
							Method:       http.MethodGet,
							URL:          newURL("/test"),
						}
						metrics.In() <- &Response{
							ConnectionID: connID,
							ExchangeID:   eID,
							Created:      refTime.Add(10 * time.Millisecond),
							StatusCode:   http.StatusOK,
						}

						ExSentMutex.Lock()
						numExchangesSent++
						ExSentMutex.Unlock()

					} else {
						metrics.In() <- &Request{
							ConnectionID: connID,
							ExchangeID:   eID,
							Created:      refTime,
							Method:       http.MethodGet,
							URL:          newURL("/test"),
						}

						TimeoutSentMutex.Lock()
						numTimeoutsSent++
						TimeoutSentMutex.Unlock()
					}
				}
			}
		}()
	}

	time.Sleep(testTimeout)
	genctx.Cancel()
	time.Sleep(6 * httptimeout)
	metricsctx.Cancel()
	time.Sleep(100 * time.Millisecond)

	ExSentMutex.Lock()
	defer ExSentMutex.Unlock()
	TimeoutSentMutex.Lock()
	defer TimeoutSentMutex.Unlock()

	refCount := map[metricsLabels]int64{
		// Complete exchanges
		metricsLabels{
			method: http.MethodGet,
			code:   strconv.Itoa(http.StatusOK),
			path:   "/test",
		}: numExchangesSent,
		// Timeouts
		metricsLabels{
			method: http.MethodGet,
			code:   "",
			path:   "/test",
		}: numTimeoutsSent,
	}

	assert.DeepEqual(t, backend.counters, refCount)

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
