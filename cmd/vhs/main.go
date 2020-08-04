package main

import (
	"log"
	"os"

	"github.com/gramLabs/vhs/capture"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vhs",
		Short: "A tool for capturing and recording network traffic.",
	}

	captureResponse bool

	address    string
	middleware string
	protocol   string

	//promMetrics bool
	promAddr string

	gcsProjectID  string
	gcsBucketName string
)

func main() {
	defer os.Exit(0)

	rootCmd.PersistentFlags().BoolVar(&captureResponse, "capture-response", false, "Capture the responses.")

	rootCmd.PersistentFlags().StringVar(&address, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	rootCmd.PersistentFlags().StringVar(&middleware, "middleware", "", "A path to an executable that VHS will use as middleware.")
	rootCmd.PersistentFlags().StringVar(&protocol, "protocol", "http", "Protocol to be used when assembling packets.")

	rootCmd.PersistentFlags().StringVar(&gcsProjectID, "gcs-project-id", "", "Project ID for Google Cloud Storage")
	rootCmd.PersistentFlags().StringVar(&gcsBucketName, "gcs-bucket-name", "", "Bucket name for Google Cloud Storage")

	// Metrics are only relevant when recording, so binding prometheus flag to record command.
	recordCmd.PersistentFlags().StringVar(&promAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")
	rootCmd.AddCommand(recordCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
