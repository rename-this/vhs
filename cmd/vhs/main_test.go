package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
	"gotest.tools/v3/assert"
)

func TestNewRootCmd(t *testing.T) {
	// TODO(andrewhare): Is there a good way to validate flags?
	// At this point just ensure that the function doesn't crash.
	_ = newRootCmd()
}

func TestRoot(t *testing.T) {
	cases := []struct {
		desc                  string
		init                  func(context.Context) *session.Config
		cfg                   *session.Config
		inputLine             string
		outputLines           []string
		validate              func(*session.Config, *bytes.Buffer)
		flowErrContains       string
		initializeErrContains string
	}{
		{
			desc: "http with middleware",
			init: func(ctx context.Context) *session.Config {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, time.Now().UnixNano())
				}))

				go func() {
					defer s.Close()
					var (
						client = s.Client()
						ticker = time.Tick(time.Second)
					)
					for {
						select {
						case <-ticker:
							_, err := client.Get(s.URL)
							assert.NilError(t, err)
						case <-ctx.Done():
							return
						}
					}
				}()

				return &session.Config{
					FlowDuration:       10 * time.Second,
					InputDrainDuration: 5 * time.Second,
					ShutdownDuration:   5 * time.Second,
					Addr:               strings.TrimLeft(s.URL, "http://"),
					CaptureResponse:    true,
					TCPTimeout:         5 * time.Minute,
					HTTPTimeout:        30 * time.Second,
					ProfileHTTPAddr:    ":81112",
					ProfilePathCPU:     "/tmp/vhs_cpu_test.prof",
					ProfilePathMemory:  "/tmp/vhs_mem_test.prof",
				}
			},
			inputLine: "tcp|http",
			outputLines: []string{
				"json|testout",
			},
			validate: func(cfg *session.Config, buf *bytes.Buffer) {
				scanner := bufio.NewScanner(buf)
				for scanner.Scan() {
					var r httpx.Request
					err := json.Unmarshal(scanner.Bytes(), &r)
					assert.NilError(t, err)
				}
				assert.NilError(t, scanner.Err())
			},
		},
		{
			desc: "http with middleware and prometheus",
			init: func(ctx context.Context) *session.Config {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, time.Now().UnixNano())
				}))

				go func() {
					defer s.Close()
					var (
						client = s.Client()
						ticker = time.Tick(time.Second)
					)
					for {
						select {
						case <-ticker:
							_, err := client.Get(s.URL)
							assert.NilError(t, err)
						case <-ctx.Done():
							return
						}
					}
				}()

				return &session.Config{
					FlowDuration:       10 * time.Second,
					InputDrainDuration: 5 * time.Second,
					ShutdownDuration:   5 * time.Second,
					Addr:               strings.TrimLeft(s.URL, "http://"),
					CaptureResponse:    true,
					TCPTimeout:         5 * time.Minute,
					HTTPTimeout:        30 * time.Second,
					PrometheusAddr:     ":8080",
					Middleware:         "../../testdata/http_middleware.bash",
				}
			},
			inputLine: "tcp|http",
			outputLines: []string{
				"json|testout",
			},
			validate: func(cfg *session.Config, buf *bytes.Buffer) {
				scanner := bufio.NewScanner(buf)
				for scanner.Scan() {
					var r httpx.Request
					err := json.Unmarshal(scanner.Bytes(), &r)
					assert.NilError(t, err)
					assert.Equal(t, true, strings.Contains(r.Body, "hijacked"))
					if r.Response != nil {
						assert.Equal(t, true, strings.Contains(r.Response.Body, "hijacked"))
					}
				}
				assert.NilError(t, scanner.Err())

				// TODO(ztreinhart): We need a better way to assert Prometheus metrics.
				promURL := fmt.Sprintf("http://localhost%s/metrics", cfg.PrometheusAddr)
				r, err := http.Get(promURL)
				assert.NilError(t, err)
				b, _ := ioutil.ReadAll(r.Body)
				r.Body.Close()
				assert.Equal(t, true, strings.Contains(string(b), "vhs_http_latency_seconds"))
			},
		},
		{
			desc: "missing middleware",
			init: func(ctx context.Context) *session.Config {
				return &session.Config{
					Middleware: "../../testdata/no_such_file",
				}
			},
			initializeErrContains: "no such file or directory",
		},
		{
			desc: "middleware crash immediately",
			init: func(ctx context.Context) *session.Config {
				return &session.Config{
					FlowDuration: time.Second,
					Middleware:   "../../testdata/crash_immediately.bash",
				}
			},
			inputLine:       "tcp|http",
			flowErrContains: "middleware crashed: exit status 1",
		},
		{
			desc: "middleware crash eventually",
			init: func(ctx context.Context) *session.Config {
				return &session.Config{
					FlowDuration: 2 * time.Second,
					Middleware:   "../../testdata/crash_eventually.bash",
				}
			},
			inputLine:       "tcp|http",
			flowErrContains: "middleware crashed: exit status 2",
		},
		{
			desc: "bad input line",
			init: func(ctx context.Context) *session.Config {
				return &session.Config{}
			},
			inputLine:             "---",
			initializeErrContains: "invalid source: ---",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var (
				cfg    = c.init(ctx)
				parser = defaultParser()
				buf    bytes.Buffer
			)

			parser.Sinks["testout"] = func(session.Context) (flow.Sink, error) {
				return ioutilx.NopWriteCloser(&buf), nil
			}

			var logBuf bytes.Buffer
			err := root(cfg, c.inputLine, c.outputLines, parser, &logBuf)
			if c.initializeErrContains != "" {
				assert.ErrorContains(t, err, c.initializeErrContains)
				return
			}

			assert.NilError(t, err)

			if c.flowErrContains == "" {
				c.validate(cfg, &buf)
			} else {
				assert.Assert(t, strings.Contains(logBuf.String(), c.flowErrContains))
			}
		})
	}
}
