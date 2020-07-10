package cmd

import (
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

func init() {
	rootCmd.PersistentFlags().StringVar(&address, "address", capture.DefaultAddr, "Address VHS will use to capture traffic.")
	rootCmd.PersistentFlags().StringVar(&middlewarePath, "middleware", "", "A path to an executable that VHS will use as middleware.")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
