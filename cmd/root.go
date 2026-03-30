package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sporkops/cli/cmd/alertchannel"
	"github.com/sporkops/cli/cmd/apikey"
	"github.com/sporkops/cli/cmd/incident"
	"github.com/sporkops/cli/cmd/ping"
	"github.com/sporkops/cli/cmd/statuspage"
	"github.com/sporkops/cli/pkg/spork"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	jsonFlag   bool
	outputFlag string
)

// SetVersion sets the CLI version string (injected via ldflags).
func SetVersion(v string) {
	version = v
}

// JSONOutput returns whether JSON output was requested.
func JSONOutput() bool {
	return jsonFlag
}

// OutputFormat returns the configured output format (table, json, or yaml).
func OutputFormat() string {
	return outputFlag
}

var rootCmd = &cobra.Command{
	Use:     "spork",
	Short:   "Spork — uptime monitoring from your terminal",
	Long:    "Manage your uptime monitors from the terminal.\n\nDocs: https://sporkops.com/docs",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// --json is a backwards-compatible alias for --output json
		if jsonFlag {
			outputFlag = "json"
		}
		// Validate output flag
		switch outputFlag {
		case "table", "json":
			// supported
		case "yaml":
			fmt.Fprintln(os.Stderr, "Warning: YAML output is not yet implemented, falling back to JSON")
			outputFlag = "json"
			jsonFlag = true
		default:
			return fmt.Errorf("invalid --output %q: must be table, json, or yaml", outputFlag)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output as JSON (shorthand for --output json)")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "table", "output format: table, json, yaml")
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
	Example: `  spork completion bash
  spork completion zsh`,
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
	Example: `  spork version
  spork version --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]string{
			"cli_version": version,
			"sdk_version": spork.Version,
			"go_version":  runtime.Version(),
			"os":          runtime.GOOS,
			"arch":        runtime.GOARCH,
		}

		if cmd.Root().Flag("json").Changed {
			enc := fmt.Sprintf(
				`{"cli_version":%q,"sdk_version":%q,"go_version":%q,"os":%q,"arch":%q}`+"\n",
				info["cli_version"], info["sdk_version"], info["go_version"], info["os"], info["arch"],
			)
			fmt.Print(enc)
			return nil
		}

		fmt.Printf("spork %s\n", version)
		fmt.Printf("  SDK:     %s\n", spork.Version)
		fmt.Printf("  Go:      %s\n", runtime.Version())
		fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
