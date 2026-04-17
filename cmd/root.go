package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sporkops/cli/cmd/alertchannel"
	"github.com/sporkops/cli/cmd/apikey"
	"github.com/sporkops/cli/cmd/incident"
	"github.com/sporkops/cli/cmd/maintenance"
	"github.com/sporkops/cli/cmd/members"
	"github.com/sporkops/cli/cmd/monitor"
	"github.com/sporkops/cli/cmd/statuspage"
	"github.com/sporkops/cli/cmd/webhook"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	version     = "dev"
	jsonFlag    bool
	outputFlag  string
	debugFlag   bool
	noColorFlag bool
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

// DebugEnabled returns whether the user requested HTTP request/response
// tracing via --debug (or the SPORK_DEBUG environment variable). Consumed
// by internal/cmdutil.RequireAuth to wrap the SDK's http.Client in a
// logging transport.
func DebugEnabled() bool {
	return debugFlag
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
		case "table", "json", "yaml":
			// supported
		default:
			return fmt.Errorf("invalid --output %q: must be table, json, or yaml", outputFlag)
		}
		// Propagate --debug (or SPORK_DEBUG) to cmdutil so every client it
		// constructs wraps the HTTP transport in debughttp.Transport.
		cmdutil.Debug = debugFlag
		// --no-color forces colors off. The default (nil) honors NO_COLOR
		// and TTY detection in internal/output.
		if noColorFlag {
			off := false
			output.SetColor(&off)
		}
		// Emit a one-line deprecation warning when the legacy `spork ping`
		// alias is used. Detection is via os.Args because Cobra does not
		// propagate CalledAs() to parent commands in the execution path
		// (it is only set on the leaf command actually dispatched).
		if firstPositional(os.Args[1:]) == "ping" {
			fmt.Fprintln(os.Stderr, "warning: `spork ping` is deprecated and will be removed in a future release; use `spork monitor` instead")
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output as JSON (shorthand for --output json)")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "log HTTP requests and responses to stderr (tokens are redacted)")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable colored output (also honors NO_COLOR env)")
	// Honor SPORK_DEBUG as a convenience for CI; flag takes precedence.
	if v := os.Getenv("SPORK_DEBUG"); v == "1" || strings.EqualFold(v, "true") {
		debugFlag = true
	}
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(members.Cmd)
	rootCmd.AddCommand(monitor.Cmd)
	rootCmd.AddCommand(apikey.Cmd)
	rootCmd.AddCommand(alertchannel.Cmd)
	rootCmd.AddCommand(incident.Cmd)
	rootCmd.AddCommand(maintenance.Cmd)
	rootCmd.AddCommand(statuspage.Cmd)
	rootCmd.AddCommand(webhook.Cmd)
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

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, info)
		}

		fmt.Printf("spork %s\n", version)
		fmt.Printf("  SDK:     %s\n", spork.Version)
		fmt.Printf("  Go:      %s\n", runtime.Version())
		fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	},
}

// firstPositional returns the first non-flag argument from argv, or "" if
// every argument is a flag or a flag value. Handles both `--foo=bar` and
// `--foo bar` forms by conservatively skipping the token following any flag
// that does not contain `=`. Unknown flags are treated as potentially taking
// a value (same conservative behaviour Cobra uses internally).
//
// Used by PersistentPreRunE to detect whether the user typed the deprecated
// `spork ping` alias vs the canonical `spork monitor`, because Cobra does
// not expose CalledAs() on parent commands.
func firstPositional(argv []string) string {
	i := 0
	for i < len(argv) {
		a := argv[i]
		if !strings.HasPrefix(a, "-") {
			return a
		}
		if a == "--" || strings.Contains(a, "=") {
			i++
			continue
		}
		i += 2
	}
	return ""
}

// Execute runs the root command.
func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
