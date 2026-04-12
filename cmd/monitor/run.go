package monitor

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	runType           string
	runMethod         string
	runExpectedStatus int
	runTimeout        int
	runHeaders        []string
	runBody           string
	runKeyword        string
	runKeywordType    string
	runSSLWarnDays    int
)

var runCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run an ephemeral check without creating a monitor",
	Long: `Execute a single check against target with the supplied configuration
and print the result. No monitor is created, no history row is written —
the probe is discarded after the result is shown.

Use this to iterate on monitor configuration (keyword choice, expected
status code, custom headers) without polluting a real monitor's history.`,
	Example: `  spork monitor run https://example.com
  spork monitor run https://api.example.com/health --expected-status 204
  spork monitor run example.com --type dns
  spork monitor run https://shop.example.com --type keyword --keyword "Add to cart"
  spork monitor run https://api.example.com --header "X-Token=abc" --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		headers, err := parseHeaders(runHeaders)
		if err != nil {
			return err
		}

		input := &spork.RunMonitorInput{
			Target:         args[0],
			Type:           runType,
			Method:         runMethod,
			ExpectedStatus: runExpectedStatus,
			Timeout:        runTimeout,
			Headers:        headers,
			Body:           runBody,
			Keyword:        runKeyword,
			KeywordType:    runKeywordType,
			SSLWarnDays:    runSSLWarnDays,
		}

		result, err := client.RunMonitor(context.Background(), input)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error running probe: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}

		renderRunResult(os.Stdout, result)
		if result.Status != "up" {
			// Non-zero exit so `spork monitor run ... && deploy` gates
			// on the probe outcome in CI scripts.
			return fmt.Errorf("probe status: %s", result.Status)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&runType, "type", "http", "monitor type: http, ssl, dns, keyword, tcp, ping")
	runCmd.Flags().StringVar(&runMethod, "method", "GET", "HTTP method (http/keyword types)")
	runCmd.Flags().IntVar(&runExpectedStatus, "expected-status", 200, "expected HTTP status code")
	runCmd.Flags().IntVar(&runTimeout, "timeout", 30, "probe timeout in seconds (5-120)")
	runCmd.Flags().StringArrayVar(&runHeaders, "header", nil, "extra HTTP header, repeat for multiple: --header \"X-Token=abc\"")
	runCmd.Flags().StringVar(&runBody, "body", "", "HTTP request body")
	runCmd.Flags().StringVar(&runKeyword, "keyword", "", "keyword to match (keyword type)")
	runCmd.Flags().StringVar(&runKeywordType, "keyword-type", "", "\"exists\" or \"not_exists\" (keyword type)")
	runCmd.Flags().IntVar(&runSSLWarnDays, "ssl-warn-days", 0, "days before SSL expiry to flag (ssl type)")
	Cmd.AddCommand(runCmd)
}

// parseHeaders turns a repeated --header "key=value" list into a map.
// Returns an error if any entry lacks the `=` separator so users get a
// usage error instead of a silent no-op.
func parseHeaders(hs []string) (map[string]string, error) {
	if len(hs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(hs))
	for _, raw := range hs {
		k, v, err := cmdutil.ParseKeyValue(raw)
		if err != nil {
			return nil, fmt.Errorf("--header %q: %w", raw, err)
		}
		out[k] = v
	}
	return out, nil
}

// renderRunResult writes a human-readable summary of the probe result.
// Structured output (--json, --output yaml) bypasses this and emits the
// typed struct.
func renderRunResult(w *os.File, r *spork.RunMonitorResult) {
	statusLabel := "DOWN"
	if r.Status == "up" {
		statusLabel = "UP"
	}
	colored := output.ColorStatus(r.Status) + " " + statusLabel
	fmt.Fprintf(w, "\n  Target:       %s\n", r.Target)
	fmt.Fprintf(w, "  Type:         %s\n", r.Type)
	fmt.Fprintf(w, "  Region:       %s\n", r.Region)
	fmt.Fprintf(w, "  Status:       %s\n", colored)
	if r.HTTPCode != 0 {
		fmt.Fprintf(w, "  HTTP code:    %d\n", r.HTTPCode)
	}
	fmt.Fprintf(w, "  Response:     %dms\n", r.ResponseTimeMs)
	fmt.Fprintf(w, "  Checked at:   %s\n", r.CheckedAt.Format("2006-01-02 15:04:05 MST"))
	if r.Error != "" {
		fmt.Fprintf(w, "  Error:        %s\n", r.Error)
	}
	fmt.Fprintln(w)
}
