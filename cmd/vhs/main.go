package main

import (
	"log"

	"github.com/gramLabs/vhs/capture"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vhs",
		Short: "A tool for capturing and recording network traffic.",
	}

	address        string
	middlewarePath string
)

func main() {
	rootCmd.PersistentFlags().StringVar(&address, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	rootCmd.PersistentFlags().StringVar(&middlewarePath, "middleware", "", "A path to an executable that VHS will use as middleware.")

	rootCmd.AddCommand(recordCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
