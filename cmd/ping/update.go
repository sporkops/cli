package ping

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	updateName           string
	updateTarget         string
	updateMethod         string
	updateInterval       int
	updatePaused         bool
	updateType           string
	updateExpectedStatus int
	updateTimeout        int
	updateRegions        []string
	updateHeaders        []string
	updateBody           string
	updateKeyword        string
	updateKeywordType    string
	updateSSLWarnDays    int
	updateAlertChannels  []string
	updateTags           []string
)

var updateCmd = &cobra.Command{
	Use:   "update <id|url>",
	Short: "Update an existing monitor",
	Long:  "Update an existing monitor's settings.\n\nExample:\n  spork ping update https://example.com --name \"New Name\"\n  spork ping update abc123 --interval 300\n\nTip: Use 'spork ping pause' and 'spork ping unpause' to pause/resume monitors.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id, name, err := resolveMonitorID(client, args[0])
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		// Build partial update — only include fields that were explicitly set.
		update := &api.Monitor{}
		hasChanges := false

		if cmd.Flags().Changed("name") {
			update.Name = updateName
			hasChanges = true
		}
		if cmd.Flags().Changed("target") {
			parsed, err := url.Parse(updateTarget)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				fmt.Fprintln(os.Stderr, "Error: URL must start with http:// or https://")
				return fmt.Errorf("invalid URL: %s", updateTarget)
			}
			update.Target = updateTarget
			hasChanges = true
		}
		if cmd.Flags().Changed("type") {
			update.Type = updateType
			hasChanges = true
		}
		if cmd.Flags().Changed("method") {
			update.Method = updateMethod
			hasChanges = true
		}
		if cmd.Flags().Changed("expected-status") {
			update.ExpectedStatus = updateExpectedStatus
			hasChanges = true
		}
		if cmd.Flags().Changed("interval") {
			update.Interval = updateInterval
			hasChanges = true
		}
		if cmd.Flags().Changed("timeout") {
			update.Timeout = updateTimeout
			hasChanges = true
		}
		if cmd.Flags().Changed("regions") {
			update.Regions = updateRegions
			hasChanges = true
		}
		if cmd.Flags().Changed("header") {
			headers := make(map[string]string)
			for _, kv := range updateHeaders {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					fmt.Fprintf(os.Stderr, "Error: invalid header format %q, expected key=value\n", kv)
					return fmt.Errorf("invalid header: %s", kv)
				}
				headers[parts[0]] = parts[1]
			}
			update.Headers = headers
			hasChanges = true
		}
		if cmd.Flags().Changed("body") {
			update.Body = updateBody
			hasChanges = true
		}
		if cmd.Flags().Changed("keyword") {
			update.Keyword = updateKeyword
			hasChanges = true
		}
		if cmd.Flags().Changed("keyword-type") {
			update.KeywordType = updateKeywordType
			hasChanges = true
		}
		if cmd.Flags().Changed("ssl-warn-days") {
			update.SSLWarnDays = updateSSLWarnDays
			hasChanges = true
		}
		if cmd.Flags().Changed("alert-channels") {
			update.AlertChannelIDs = updateAlertChannels
			hasChanges = true
		}
		if cmd.Flags().Changed("tags") {
			update.Tags = updateTags
			hasChanges = true
		}
		if cmd.Flags().Changed("paused") {
			update.Paused = &updatePaused
			hasChanges = true
		}

		if !hasChanges {
			fmt.Fprintln(os.Stderr, "Nothing to update. Specify at least one flag:")
			fmt.Fprintln(os.Stderr, "  --name, --target, --type, --method, --interval, --timeout, --paused, etc.")
			return fmt.Errorf("no changes specified")
		}

		result, err := client.UpdateMonitor(id, update)
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error updating monitor: %s\n", err)
			return err
		}

		label := result.Name
		if label == "" {
			label = name
		}
		fmt.Printf("✓ Monitor updated: %s (%s)\n", label, result.Target)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "new monitor name")
	updateCmd.Flags().StringVar(&updateTarget, "target", "", "new target URL to monitor")
	updateCmd.Flags().StringVar(&updateType, "type", "", "monitor type: http, ssl, dns, keyword, tcp, ping")
	updateCmd.Flags().StringVar(&updateMethod, "method", "", "HTTP method (GET, HEAD, POST, PUT)")
	updateCmd.Flags().IntVar(&updateExpectedStatus, "expected-status", 0, "expected HTTP status code (100-599)")
	updateCmd.Flags().IntVar(&updateInterval, "interval", 0, "check interval in seconds (60-86400, multiple of 60)")
	updateCmd.Flags().IntVar(&updateTimeout, "timeout", 0, "timeout per check in seconds (5-120)")
	updateCmd.Flags().StringSliceVar(&updateRegions, "regions", nil, "check regions (us-central1, europe-west1)")
	updateCmd.Flags().StringArrayVar(&updateHeaders, "header", nil, "custom HTTP header as key=value (repeatable)")
	updateCmd.Flags().StringVar(&updateBody, "body", "", "HTTP request body for POST/PUT")
	updateCmd.Flags().StringVar(&updateKeyword, "keyword", "", "keyword to search in response")
	updateCmd.Flags().StringVar(&updateKeywordType, "keyword-type", "", "keyword match type: exists, not_exists")
	updateCmd.Flags().IntVar(&updateSSLWarnDays, "ssl-warn-days", 0, "days before SSL expiry to warn")
	updateCmd.Flags().StringSliceVar(&updateAlertChannels, "alert-channels", nil, "alert channel IDs to notify")
	updateCmd.Flags().StringSliceVar(&updateTags, "tags", nil, "organization tags")
	updateCmd.Flags().BoolVar(&updatePaused, "paused", false, "pause or unpause the monitor")
}
