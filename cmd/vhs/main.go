package main

import (
	"log"
	"time"

	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vhs",
		Short: "A tool for capturing and recording network traffic.",
	}

	cfg = &config.Config{}
)

func main() {
	rootCmd.PersistentFlags().DurationVar(&cfg.FlowDuration, "flow-duration", 0, "The length of the running command.")
	rootCmd.PersistentFlags().DurationVar(&cfg.InputDrainDuration, "input-drain-duration", 10*time.Second, "A grace period to allow for a inputs to drain.")
	rootCmd.PersistentFlags().DurationVar(&cfg.ShutdownDuration, "shutdown-duration", 10*time.Second, "A grace period to allow for a clean shutdown.")

	rootCmd.PersistentFlags().StringVar(&cfg.Addr, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	rootCmd.PersistentFlags().BoolVar(&cfg.CaptureResponse, "capture-response", false, "Capture the responses.")
	rootCmd.PersistentFlags().StringVar(&cfg.Middleware, "middleware", "", "A path to an executable that VHS will use as middleware.")
	rootCmd.PersistentFlags().DurationVar(&cfg.TCPTimeout, "tcp-timeout", 5*time.Minute, "A length of time after which unused TCP connections are closed.")

	// Metrics are only relevant when recording, so binding prometheus flag to record command.
	recordCmd.PersistentFlags().StringVar(&cfg.PrometheusAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")

	rootCmd.PersistentFlags().StringVar(&cfg.GCSBucketName, "gcs-bucket-name", "", "Bucket name for Google Cloud Storage")

	rootCmd.AddCommand(recordCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
