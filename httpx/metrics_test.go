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
			ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{
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
		testTimeout = 10 * time.Second
		httpTimeout = 50 * time.Millisecond

		numSenders           = 4
		numMessagesPerSender = 250

		numExchangesSent int64 = 0
		numTimeoutsSent  int64 = 0

		timeoutCountCh = make(chan struct{})
		messageCountCh = make(chan struct{})
		errs           = make(chan error)
	)

	genctx := session.NewContexts(&session.Config{
		Debug:             false,
		DebugHTTPMessages: false,
	}, &session.FlowConfig{
		HTTPTimeout: 50 * time.Millisecond, // Short correlator time so we can actually get some timeouts.
	}, errs)

	metricsctx := session.NewContexts(&session.Config{
		Debug:             false,
		DebugHTTPMessages: false,
	}, &session.FlowConfig{
		HTTPTimeout: 50 * time.Millisecond, // Short correlator time so we can actually get some timeouts.
	}, errs)

	backend := newTestMetricsBackend()
	metrics := &Metrics{
		in:  make(chan interface{}),
		met: backend,
	}

	go metrics.Init(metricsctx, nil)

	for i := 0; i < numSenders; i++ {
		go func() {
			for msgCount := 0; msgCount < numMessagesPerSender; msgCount++ {
				connID := ksuid.New().String()
				eID := ksuid.New().String()
				timeout := rand.Intn(2)

				// Send Request
				metrics.In() <- &Request{
					ConnectionID: connID,
					ExchangeID:   eID,
					Created:      refTime,
					Method:       http.MethodGet,
					URL:          newURL("/test"),
				}

				if timeout == 1 {
					timeoutCountCh <- struct{}{}
					continue
				}

				// Send Response
				metrics.In() <- &Response{
					ConnectionID: connID,
					ExchangeID:   eID,
					Created:      refTime.Add(10 * time.Millisecond),
					StatusCode:   http.StatusOK,
				}

				messageCountCh <- struct{}{}
			}
		}()
	}

	for {
		select {
		case <-messageCountCh:
			numExchangesSent++
		case <-timeoutCountCh:
			numTimeoutsSent++
		case err := <-errs:
			assert.NilError(t, err)
		case <-time.After(testTimeout):
			break
		}

		if numExchangesSent+numTimeoutsSent == int64(numSenders*numMessagesPerSender) {
			break
		}
	}

	genctx.Cancel()
	// This should be max httpTimeout, but need a little buffer to handle progression through the correlator
	// In testing locally 250ms seems to be adequate
	time.Sleep(5 * httpTimeout)

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

	backend.Lock()
	assert.DeepEqual(t, backend.counters, refCount)
	backend.Unlock()
}

// testMetricsBackend is a metricsBackend for testing
type testMetricsBackend struct {
	sync.Mutex
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
	t.Lock()
	t.counters[l]++
	t.Unlock()
}

// AddDuration adds the measured http exchange duration with the specified labels.
func (t *testMetricsBackend) AddDuration(l metricsLabels, f float64) {
	t.Lock()
	t.durations[l] = append(t.durations[l], f)
	t.Unlock()
}
