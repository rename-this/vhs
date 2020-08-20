package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/pipe"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
	"github.com/gramLabs/vhs/tcp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	httpTimeout = 30 * time.Second
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record network traffic",
	Run:   record,
}

func record(cmd *cobra.Command, args []string) {
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

	var (
		m = mustStartMiddleware(ctx)
		f = &flow.Flow{
			Input: pipe.NewInput(httpx.NewInputFormat(), tcp.NewSource()),
			Outputs: pipe.Outputs{
				pipe.NewOutput(format.NewJSON(), os.Stdout),
			},
		}
	)

	if cfg.GCSBucketName != "" {
		gcs, err := sink.NewGCS(ctx)
		if err != nil {
			log.Fatalf("failed to initialize GCS sink: %v", err)
		}
		p := pipe.NewOutput(format.NewJSONBuffered(), gcs, &modifier.GzipWriteCloser{})
		f.Outputs = append(f.Outputs, p)
	}

	// Add the metrics pipe if the user has enabled Prometheus metrics.
	if cfg.PrometheusAddr != "" {
		f.Outputs = append(f.Outputs, httpx.NewMetricsPipe(httpTimeout))
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(http.ListenAndServe(cfg.PrometheusAddr, nil))
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("user shutdown requested")
		ctx.Cancel()
	}()

	f.Run(ctx, inputCtx, outputCtx, m)
}
