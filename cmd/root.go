package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/cmd/alertchannel"
	"github.com/sporkops/cli/cmd/apikey"
	"github.com/sporkops/cli/cmd/incident"
	"github.com/sporkops/cli/cmd/ping"
	"github.com/sporkops/cli/cmd/statuspage"
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
	Short:   "Spork — uptime monitoring from your terminal",
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
	rootCmd.AddCommand(alertchannel.Cmd)
	rootCmd.AddCommand(incident.Cmd)
	rootCmd.AddCommand(statuspage.Cmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for your shell.

To load completions:

Bash:
  source <(spork completion bash)

Zsh:
  source <(spork completion zsh)

Fish:
  spork completion fish | source`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("spork %s\n", version)
	},
}

// Execute runs the root command.
func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
