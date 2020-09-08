package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/gcs"
	"github.com/gramLabs/vhs/gzipx"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/jsonx"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/tcp"
	"go.uber.org/multierr"

	"github.com/go-errors/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
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
	cmd.PersistentFlags().DurationVar(&cfg.InputDrainDuration, "input-drain-duration", 10*time.Second, "A grace period to allow for a inputs to drain.")
	cmd.PersistentFlags().DurationVar(&cfg.ShutdownDuration, "shutdown-duration", 10*time.Second, "A grace period to allow for a clean shutdown.")
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

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return root(cfg, inputLine, outputLines, defaultParser())
	}

	return cmd
}

func root(cfg *session.Config, inputLine string, outputLines []string, parser *flow.Parser) error {
	var (
		errs                     = make(chan error)
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

	m, err := startMiddleware(ctx)
	if err != nil {
		return errors.Errorf("failed to start middleware: %v", err)
	}

	f, err := parser.Parse(ctx, inputLine, outputLines)
	if err != nil {
		return errors.Errorf("failed to initialize: %v", err)
	}

	// Add the metrics pipe if the user has enabled Prometheus metrics.
	if cfg.PrometheusAddr != "" {
		f.Outputs = append(f.Outputs, httpx.NewMetricsOutput())
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(http.ListenAndServe(cfg.PrometheusAddr, nil))
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\ruser shutdown requested...")
		ctx.Cancel()
	}()

	f.Run(ctx, inputCtx, outputCtx, m)

	return allErrs
}

func startMiddleware(ctx *session.Context) (middleware.Middleware, error) {
	if ctx.Config.Middleware == "" {
		return nil, nil
	}

	m, err := middleware.New(ctx, ctx.Config.Middleware, os.Stderr)
	if err != nil {
		return nil, errors.Errorf("failed to create middleware: %w", err)
	}

	if err := m.Start(); err != nil {
		return nil, errors.Errorf("failed to start middleware: %w", err)
	}

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
			"stdout": func(_ *session.Context) (flow.Sink, error) {
				return os.Stdout, nil
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
