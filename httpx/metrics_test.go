package httpx

import (
	"testing"
	"time"

	"github.com/gramLabs/vhs/session"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gotest.tools/v3/assert"
)

var refTime = time.Date(2020, 10, 12, 12, 0, 0, 0, time.UTC)

func TestMetrics(t *testing.T) {
	cases := []struct {
		desc     string
		messages []Message
		latency  float64
		timeouts float64
	}{
		{
			desc: "latency only",
			messages: []Message{
				&Request{
					ConnectionID: "111",
					ExchangeID:   0,
					Created:      refTime,
					Method:       "POST",
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   0,
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   200,
				},
			},
			latency:  0.5,
			timeouts: 0,
		},
		{
			desc: "timeout only",
			messages: []Message{
				&Request{
					ConnectionID: "222",
					ExchangeID:   1,
					Created:      refTime.Add(750 * time.Millisecond),
					Method:       "GET",
					Response:     nil,
				},
			},
			latency:  0,
			timeouts: 1,
		},
		{
			desc: "combined",
			messages: []Message{
				&Request{
					ConnectionID: "111",
					ExchangeID:   0,
					Created:      refTime,
					Method:       "POST",
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   0,
					Created:      refTime.Add(500 * time.Millisecond),
					StatusCode:   200,
				},
				&Request{
					ConnectionID: "222",
					ExchangeID:   1,
					Created:      refTime.Add(750 * time.Millisecond),
					Method:       "GET",
					Response:     nil,
				},
			},
			latency:  0.5,
			timeouts: 1,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx, _, _ := session.NewContexts(&session.Config{
				HTTPTimeout: 50 * time.Millisecond, // Short correlator time so we can actually get some timeouts.
			}, nil)

			met := NewMetrics()
			go met.Init(ctx, nil)

			for _, m := range c.messages {
				met.In() <- m
				time.Sleep(10 * time.Millisecond)
			}

			//Sleep long enough to let a prune cycle run in the correlator.
			time.Sleep(300 * time.Millisecond)

			ctx.Cancel()

			latencyMetric := met.latency.With(prometheus.Labels{
				"method": "POST",
				"code":   "200",
			})

			timeoutMetric := met.timeouts.With(prometheus.Labels{
				"method": "GET",
			})

			assert.Equal(t, c.latency, testutil.ToFloat64(latencyMetric))
			assert.Equal(t, c.timeouts, testutil.ToFloat64(timeoutMetric))

			prometheus.Unregister(met.latency)
			prometheus.Unregister(met.timeouts)
		})
	}
}
