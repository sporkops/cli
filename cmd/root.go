package cmd

import (
	"github.com/sporkops/cli/cmd/apikey"
	"github.com/sporkops/cli/cmd/ping"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	jsonFlag bool
)

// SetVersion sets the CLI version string (injected via ldflags).
func SetVersion(v string) {
	version = v
}

// JSONOutput returns whether JSON output was requested.
func JSONOutput() bool {
	return jsonFlag
}

var rootCmd = &cobra.Command{
	Use:     "spork",
	Short:   "CLI for Ping by Spork — uptime monitoring",
	Long:    "Manage your uptime monitors from the terminal.\n\nDocs: https://sporkops.com/docs",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output as JSON")
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(ping.Cmd)
	rootCmd.AddCommand(apikey.Cmd)
}

// Execute runs the root command.
func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
