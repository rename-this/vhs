package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rename-this/vhs/session"
)

func SetConfigFlags(cmd *cobra.Command, cfg *session.Config) {
	cmd.PersistentFlags().StringVar(&cfg.PrometheusAddr, "prometheus-address", "", "Address for Prometheus metrics HTTP endpoint.")
	cmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Emit debug logging.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugPackets, "debug-packets", false, "Emit all packets as debug logs.")
	cmd.PersistentFlags().BoolVar(&cfg.DebugHTTPMessages, "debug-http-messages", false, "Emit all parsed HTTP messages as debug logs.")

	cmd.PersistentFlags().StringVar(&cfg.ProfilePathCPU, "profile-path-cpu", "", "Output CPU profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfilePathMemory, "profile-path-memory", "", "Output memory profile to this path.")
	cmd.PersistentFlags().StringVar(&cfg.ProfileHTTPAddr, "profile-http-address", "", "Expose profile data on this address.")
}
