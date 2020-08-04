package main

import "github.com/spf13/cobra"

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay network traffic",
	Run:   replay,
}

func replay(cmd *cobra.Command, args []string) {

}
