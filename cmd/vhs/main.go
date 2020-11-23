package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/rename-this/vhs/capture"
	"github.com/rename-this/vhs/file"
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/gcs"
	"github.com/rename-this/vhs/gzipx"
	"github.com/rename-this/vhs/httpx"
	"github.com/rename-this/vhs/internal/ioutilx"
	"github.com/rename-this/vhs/jsonx"
	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/s3compat"
	"github.com/rename-this/vhs/session"
	"github.com/rename-this/vhs/tcp"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	errBufSize = 10
)

func main() {
	newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	var (
		cmd = &cobra.Command{
			Short: "A tool for capturing and recording network traffic.",
		}

		cfg         = &session.Config{}
		flowCfg     = &session.FlowConfig{}
		inputLine   string
		outputLines []string
	)

	cmd.PersistentFlags().DurationVar(&flowCfg.SourceDuration, "source-duration", math.MaxInt64, "The length of the source is left open. Leave this empty to read to EOF.")
	cmd.PersistentFlags().DurationVar(&flowCfg.DrainDuration, "drain-duration", 10*time.Second, "A grace period to allow for data to drain to sink(s).")
	cmd.PersistentFlags().StringVar(&flowCfg.Addr, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	cmd.PersistentFlags().BoolVar(&flowCfg.CaptureResponse, "capture-response", false, "Capture the responses.")
	cmd.PersistentFlags().StringVar(&flowCfg.Middleware, "middleware", "", "A path to an executable that VHS will use as middleware.")
	cmd.PersistentFlags().DurationVar(&flowCfg.TCPTimeout, "tcp-timeout", 5*time.Minute, "A length of time after which unused TCP connections are closed.")
	cmd.PersistentFlags().DurationVar(&flowCfg.HTTPTimeout, "http-timeout", 30*time.Second, "A length of time after which an HTTP request is considered to have timed out.")
	cmd.PersistentFlags().StringVar(&cfg.PrometheusAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")
	cmd.PersistentFlags().StringVar(&flowCfg.GCSBucketName, "gcs-bucket-name", "", "Bucket name for Google Cloud Storage")
	cmd.PersistentFlags().StringVar(&flowCfg.GCSObjectName, "gcs-object-name", "", "Object name for Google Cloud Storage")
	cmd.PersistentFlags().StringVar(&flowCfg.InputFile, "input-file", "", "Path to an input file")

	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatEndpoint, "s3-compat-endpoint", "", "URL for S3-compatible storage.")
	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatAccessKey, "s3-compat-access-key", "", "Access key for S3-compatible storage.")
	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatSecretKey, "s3-compat-secret-key", "", "Secret key for S3-compatible storage.")
	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatToken, "s3-compat-token", "", "Security token for S3-compatible storage.")
	cmd.PersistentFlags().BoolVar(&flowCfg.S3CompatSecure, "s3-compat-secure", true, "Encrypt communication for S3-compatible storage.")
	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatBucketName, "s3-compat-bucket-name", "", "Bucket name for S3-compatible storage.")
	cmd.PersistentFlags().StringVar(&flowCfg.S3CompatObjectName, "s3-compat-object-name", "", "Object name for S3-compatible storage.")

	cmd.PersistentFlags().StringVar(&inputLine, "input", "", "Input description.")
	cmd.PersistentFlags().StringSliceVar(&outputLines, "output", nil, "Output description.")

	cmd.PersistentFlags().BoolVar(&flowCfg.BufferOutput, "buffer-output", false, "Buffer output until the end of the flow.")
	cmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Emit debug logging.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugPackets, "debug-packets", false, "Emit all packets as debug logs.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugHTTPMessages, "debug-http-messages", false, "Emit all parsed HTTP messages as debug logs.")

	cmd.PersistentFlags().StringVar(&cfg.ProfilePathCPU, "profile-path-cpu", "", "Output CPU profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfilePathMemory, "profile-path-memory", "", "Output memory profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfileHTTPAddr, "profile-http-address", "", "Expose profile data on this address.")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := root(cfg, flowCfg, inputLine, outputLines, defaultParser(), os.Stderr)
		if err != nil {
			fmt.Printf("failed to initialize vhs: %v", err)
		}
	}

	return cmd
}

func root(cfg *session.Config, flowCfg *session.FlowConfig, inputLine string, outputLines []string, parser *flow.Parser, logWriter io.Writer) error {
	var (
		errs = make(chan error, errBufSize)
		ctx  = session.NewContextsForWriter(cfg, flowCfg, errs, logWriter)
	)

	go func() {
		for err := range errs {
			if err != nil {
				ctx.Logger.Err(err).Msg("flow error")
			}
		}
	}()

	ctx.Logger.Debug().Msg("hello, vhs")

	m, err := startMiddleware(ctx)
	if err != nil {
		return fmt.Errorf("failed to start middleware: %v", err)
	}

	f, err := parser.Parse(ctx, inputLine, outputLines)
	if err != nil {
		return fmt.Errorf("failed to initialize: %v", err)
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
			return fmt.Errorf("failed to create %s: %v", cfg.ProfilePathCPU, err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("failed to start CPU profile: %v", err)
		}
		ctx.Logger.Debug().Str("path", cfg.ProfilePathCPU).Msg("CPU profiling enabled")
		defer pprof.StopCPUProfile()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx.Logger.Debug().Msg("shutdown requested")
		ctx.Cancel()
	}()

	go f.Run(ctx, m)

	<-ctx.StdContext.Done()

	if cfg.ProfilePathMemory != "" {
		f, err := os.Create(cfg.ProfilePathMemory)
		if err != nil {
			return fmt.Errorf("failed to create %s: %v", cfg.ProfilePathMemory, err)
		}
		defer f.Close()

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("failed to start write heap profile: %v", err)
		}

		ctx.Logger.Debug().Str("path", cfg.ProfilePathMemory).Msg("memory profile written")
	}

	return nil
}

func startMiddleware(ctx session.Context) (middleware.Middleware, error) {
	if ctx.FlowConfig.Middleware == "" {
		ctx.Logger.Debug().Msg("no middleware configured")
		return nil, nil
	}

	m, err := middleware.New(ctx, ctx.FlowConfig.Middleware)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware: %w", err)
	}

	ctx.Logger.Debug().Msg("middleware created")

	if err := m.Start(); err != nil {
		return nil, fmt.Errorf("failed to start middleware: %w", err)
	}

	ctx.Logger.Debug().Msg("middleware started")

	go func() {
		if err := m.Wait(); err != nil {
			ctx.Errors <- fmt.Errorf("middleware crashed: %w", err)
		}
	}()

	return m, nil
}

func defaultParser() *flow.Parser {
	return &flow.Parser{
		Sources: map[string]flow.SourceCtor{
			"tcp":      tcp.NewSource,
			"gcs":      gcs.NewSource,
			"file":     file.NewSource,
			"s3compat": s3compat.NewSource,
		},

		InputFormats: map[string]flow.InputFormatCtor{
			"http": httpx.NewInputFormat,
			"json": jsonx.NewInputFormat,
		},

		OutputFormats: map[string]flow.OutputFormatCtor{
			"har":  httpx.NewHAR,
			"json": jsonx.NewOutputFormat,
		},

		Sinks: map[string]flow.SinkCtor{
			"gcs":      gcs.NewSink,
			"s3compat": s3compat.NewSink,
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
