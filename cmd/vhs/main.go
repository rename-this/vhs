package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/config"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/gcs"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
	"github.com/gramLabs/vhs/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	parser = &flow.Parser{
		Sources: map[string]flow.SourceCtor{
			"tcp": tcp.NewSource,
			"gcs": gcs.NewSource,
		},

		InputFormats: map[string]flow.InputFormatCtor{
			"http": httpx.NewInputFormat,
		},

		OutputFormats: map[string]flow.OutputFormatCtor{
			"har":     httpx.NewHAR,
			"json":    format.NewJSON,
			"jsonbuf": format.NewJSONBuffered,
		},

		Sinks: map[string]flow.SinkCtor{
			"gcs": gcs.NewSink,
			"stdout": func(_ *session.Context) (sink.Sink, error) {
				return os.Stdout, nil
			},
		},

		InputModifiers: map[string]flow.InputModifierCtor{
			"gzip": modifier.NewGzipInput,
		},

		OutputModifiers: map[string]flow.OutputModifierCtor{
			"gzip": modifier.NewGzipOutput,
		},
	}

	rootCmd = &cobra.Command{
		Use:   "vhs",
		Short: "A tool for capturing and recording network traffic.",
		Run:   runRoot,
	}

	cfg = &config.Config{}

	inputLine   string
	outputLines []string
)

func main() {
	rootCmd.PersistentFlags().DurationVar(&cfg.FlowDuration, "flow-duration", 10*time.Second, "The length of the running command.")
	rootCmd.PersistentFlags().DurationVar(&cfg.InputDrainDuration, "input-drain-duration", 10*time.Second, "A grace period to allow for a inputs to drain.")
	rootCmd.PersistentFlags().DurationVar(&cfg.ShutdownDuration, "shutdown-duration", 10*time.Second, "A grace period to allow for a clean shutdown.")
	rootCmd.PersistentFlags().StringVar(&cfg.Addr, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	rootCmd.PersistentFlags().BoolVar(&cfg.CaptureResponse, "capture-response", false, "Capture the responses.")
	rootCmd.PersistentFlags().StringVar(&cfg.Middleware, "middleware", "", "A path to an executable that VHS will use as middleware.")
	rootCmd.PersistentFlags().DurationVar(&cfg.TCPTimeout, "tcp-timeout", 5*time.Minute, "A length of time after which unused TCP connections are closed.")
	rootCmd.PersistentFlags().DurationVar(&cfg.HTTPTimeout, "http-timeout", 30*time.Second, "A length of time after which an HTTP request is considered to have timed out.")
	rootCmd.PersistentFlags().StringVar(&cfg.PrometheusAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")
	rootCmd.PersistentFlags().StringVar(&cfg.GCSBucketName, "gcs-bucket-name", "", "Bucket name for Google Cloud Storage")
	rootCmd.PersistentFlags().StringVar(&inputLine, "input", "", "Input description.")
	rootCmd.PersistentFlags().StringSliceVar(&outputLines, "output", nil, "Output description.")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	var (
		errs                     = make(chan error)
		ctx, inputCtx, outputCtx = session.NewContexts(cfg, errs)
	)

	go func() {
		for err := range errs {
			if err != nil {
				log.Printf("ERR: %v\n", (err.(*errors.Error).ErrorStack()))
			}
		}
	}()

	m := mustStartMiddleware(ctx)

	f, err := parser.Parse(ctx, inputLine, outputLines)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}

	// Add the metrics pipe if the user has enabled Prometheus metrics.
	if cfg.PrometheusAddr != "" {
		f.Outputs = append(f.Outputs, httpx.NewMetricsPipe(cfg.HTTPTimeout))
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
}

func mustStartMiddleware(ctx *session.Context) *middleware.Middleware {
	if ctx.Config.Middleware == "" {
		return nil
	}

	m, err := middleware.New(ctx, ctx.Config.Middleware, os.Stderr)
	if err != nil {
		log.Fatalf("failed to create middleware: %v\n", err)
	}

	go func() {
		if err := m.Start(); err != nil {
			log.Fatalf("failed to start middleware: %v\n", err)
		}
	}()

	return m
}
