package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/gcs"
	"github.com/gramLabs/vhs/gzipx"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/jsonx"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/tcp"
	"go.uber.org/multierr"

	_ "net/http/pprof"

	"github.com/go-errors/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	errBufSize = 10
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	var (
		cmd = &cobra.Command{
			Short: "A tool for capturing and recording network traffic.",
		}

		cfg         = &session.Config{}
		inputLine   string
		outputLines []string
	)

	cmd.PersistentFlags().DurationVar(&cfg.FlowDuration, "flow-duration", 10*time.Second, "The length of the running command.")
	cmd.PersistentFlags().DurationVar(&cfg.InputDrainDuration, "input-drain-duration", 2*time.Second, "A grace period to allow for a inputs to drain.")
	cmd.PersistentFlags().DurationVar(&cfg.ShutdownDuration, "shutdown-duration", 2*time.Second, "A grace period to allow for a clean shutdown.")
	cmd.PersistentFlags().StringVar(&cfg.Addr, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	cmd.PersistentFlags().BoolVar(&cfg.CaptureResponse, "capture-response", false, "Capture the responses.")
	cmd.PersistentFlags().StringVar(&cfg.Middleware, "middleware", "", "A path to an executable that VHS will use as middleware.")
	cmd.PersistentFlags().DurationVar(&cfg.TCPTimeout, "tcp-timeout", 5*time.Minute, "A length of time after which unused TCP connections are closed.")
	cmd.PersistentFlags().DurationVar(&cfg.HTTPTimeout, "http-timeout", 30*time.Second, "A length of time after which an HTTP request is considered to have timed out.")
	cmd.PersistentFlags().StringVar(&cfg.PrometheusAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")
	cmd.PersistentFlags().StringVar(&cfg.GCSBucketName, "gcs-bucket-name", "", "Bucket name for Google Cloud Storage")
	cmd.PersistentFlags().StringVar(&cfg.GCSObjectName, "gcs-object-name", "", "Object name for Google Cloud Storage")
	cmd.PersistentFlags().StringVar(&inputLine, "input", "", "Input description.")
	cmd.PersistentFlags().StringSliceVar(&outputLines, "output", nil, "Output description.")

	cmd.PersistentFlags().BoolVar(&cfg.BufferOutput, "buffer-output", false, "Buffer output until the end of the flow.")
	cmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Emit debug logging.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugPackets, "debug-packets", false, "Emit all packets as debug logs.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugHTTPMessages, "debug-http-messages", false, "Emit all parsed HTTP messages as debug logs.")

	cmd.PersistentFlags().StringVar(&cfg.ProfilePathCPU, "profile-path-cpu", "", "Output CPU profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfilePathMemory, "profile-path-memory", "", "Output memory profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfileHTTPAddr, "profile-http-address", "", "Expose profile data on this address.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return root(cfg, inputLine, outputLines, defaultParser())
	}

	return cmd
}

func root(cfg *session.Config, inputLine string, outputLines []string, parser *flow.Parser) error {
	var (
		errs                     = make(chan error, errBufSize)
		allErrs                  error
		ctx, inputCtx, outputCtx = session.NewContexts(cfg, errs)
	)

	go func() {
		for err := range errs {
			if err != nil {
				allErrs = multierr.Append(allErrs, err)
			}
		}
	}()

	ctx.Logger.Debug().Msg("hello, vhs")

	m, err := startMiddleware(ctx)
	if err != nil {
		return errors.Errorf("failed to start middleware: %v", err)
	}

	f, err := parser.Parse(ctx, inputLine, outputLines)
	if err != nil {
		return errors.Errorf("failed to initialize: %v", err)
	}

	ctx.Logger.Debug().Msg("flow created")

	// Add the metrics pipe if the user has enabled Prometheus metrics.
	if cfg.PrometheusAddr != "" {
		endpoint := "/metrics"
		ctx.Logger.Debug().Msgf("listening for prometheus on %s%s", cfg.PrometheusAddr, endpoint)

		f.Outputs = append(f.Outputs, httpx.NewMetricsOutput())

		mux := http.NewServeMux()
		mux.Handle(endpoint, promhttp.Handler())

		go func() {
			err := http.ListenAndServe(cfg.PrometheusAddr, mux)
			if errors.Is(err, http.ErrServerClosed) {
				ctx.Logger.Error().Err(err).Msg("failed to listen and serve promentheus endpoint")
			}
		}()
	}

	if cfg.ProfileHTTPAddr != "" {
		go func() {
			err := http.ListenAndServe(cfg.ProfileHTTPAddr, nil)
			if errors.Is(err, http.ErrServerClosed) {
				ctx.Logger.Error().Err(err).Msg("failed to listen and serve pprof endpoint")
			}
		}()
	}

	if cfg.ProfilePathCPU != "" {
		f, err := os.Create(cfg.ProfilePathCPU)
		if err != nil {
			return errors.Errorf("failed to create %s: %v", cfg.ProfilePathCPU, err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return errors.Errorf("failed to start CPU profile: %v", err)
		}
		ctx.Logger.Debug().Str("path", cfg.ProfilePathCPU).Msg("CPU profiling enabled")
		defer pprof.StopCPUProfile()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx.Logger.Debug().Msgf("shutdown initiated, exiting in %s",
			ctx.Config.InputDrainDuration+ctx.Config.ShutdownDuration)
		ctx.Cancel()
	}()

	f.Run(ctx, inputCtx, outputCtx, m)

	if cfg.ProfilePathMemory != "" {
		f, err := os.Create(cfg.ProfilePathMemory)
		if err != nil {
			return errors.Errorf("failed to create %s: %v", cfg.ProfilePathMemory, err)
		}
		defer f.Close()

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return errors.Errorf("failed to start write heap profile: %v", err)
		}

		ctx.Logger.Debug().Str("path", cfg.ProfilePathMemory).Msg("memory profile written")
	}

	return allErrs
}

func startMiddleware(ctx session.Context) (middleware.Middleware, error) {
	if ctx.Config.Middleware == "" {
		ctx.Logger.Debug().Msg("no middleware configured")
		return nil, nil
	}

	m, err := middleware.New(ctx, ctx.Config.Middleware)
	if err != nil {
		return nil, errors.Errorf("failed to create middleware: %w", err)
	}

	ctx.Logger.Debug().Msg("middleware created")

	if err := m.Start(); err != nil {
		return nil, errors.Errorf("failed to start middleware: %w", err)
	}

	ctx.Logger.Debug().Msg("middleware started")

	go func() {
		if err := m.Wait(); err != nil {
			ctx.Errors <- errors.Errorf("middleware crashed: %w", err)
		}
	}()

	return m, nil
}

func defaultParser() *flow.Parser {
	return &flow.Parser{
		Sources: map[string]flow.SourceCtor{
			"tcp": tcp.NewSource,
			"gcs": gcs.NewSource,
		},

		InputFormats: map[string]flow.InputFormatCtor{
			"http": httpx.NewInputFormat,
		},

		OutputFormats: map[string]flow.OutputFormatCtor{
			"har":     httpx.NewHAR,
			"json":    jsonx.NewOutputFormat,
			"jsonbuf": jsonx.NewBufferedOutputFormat,
		},

		Sinks: map[string]flow.SinkCtor{
			"gcs": gcs.NewSink,
			"stdout": func(_ session.Context) (flow.Sink, error) {
				return os.Stdout, nil
			},
			"discard": func(_ session.Context) (flow.Sink, error) {
				return ioutilx.NopWriteCloser(ioutil.Discard), nil
			},
		},

		InputModifiers: map[string]flow.InputModifierCtor{
			"gzip": gzipx.NewInputModifier,
		},

		OutputModifiers: map[string]flow.OutputModifierCtor{
			"gzip": gzipx.NewOutputModifier,
		},
	}
}
