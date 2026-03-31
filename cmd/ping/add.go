package ping

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	addName           string
	addMethod         string
	addInterval       int
	addType           string
	addExpectedStatus int
	addTimeout        int
	addRegions        []string
	addHeaders        []string
	addBody           string
	addKeyword        string
	addKeywordType    string
	addSSLWarnDays    int
	addAlertChannels  []string
	addTags           []string
)

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a new uptime monitor",
	Long: `Add a new uptime monitor for the given URL.

Example:
  spork ping add https://example.com
  spork ping add https://api.example.com/health --name "API Health" --interval 300
  spork ping add https://example.com --type keyword --keyword "OK"
  spork ping add https://example.com --type ssl --ssl-warn-days 30`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		rawURL := args[0]
		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			fmt.Fprintln(os.Stderr, "Error: URL must start with http:// or https://")
			return fmt.Errorf("invalid URL: %s", rawURL)
		}

		// Validate flags to catch errors before hitting the API.
		validTypes := map[string]bool{"http": true, "ssl": true, "dns": true, "keyword": true, "tcp": true, "ping": true}
		if !validTypes[addType] {
			return fmt.Errorf("invalid --type %q: must be one of http, ssl, dns, keyword, tcp, ping", addType)
		}
		validMethods := map[string]bool{"GET": true, "HEAD": true, "POST": true, "PUT": true}
		if !validMethods[addMethod] {
			return fmt.Errorf("invalid --method %q: must be one of GET, HEAD, POST, PUT", addMethod)
		}
		if addInterval < 60 || addInterval > 86400 {
			return fmt.Errorf("invalid --interval %d: must be between 60 and 86400", addInterval)
		}
		if addInterval%60 != 0 {
			return fmt.Errorf("invalid --interval %d: must be a multiple of 60", addInterval)
		}
		if addExpectedStatus < 100 || addExpectedStatus > 599 {
			return fmt.Errorf("invalid --expected-status %d: must be between 100 and 599", addExpectedStatus)
		}
		if addTimeout < 5 || addTimeout > 120 {
			return fmt.Errorf("invalid --timeout %d: must be between 5 and 120", addTimeout)
		}
		if addKeywordType != "" && addKeywordType != "exists" && addKeywordType != "not_exists" {
			return fmt.Errorf("invalid --keyword-type %q: must be exists or not_exists", addKeywordType)
		}
		if addType == "keyword" && addKeyword == "" {
			return fmt.Errorf("--keyword is required when --type is keyword")
		}

		name := addName
		if name == "" {
			name = parsed.Hostname()
		}

		headers := make(map[string]string)
		for _, kv := range addHeaders {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Error: invalid header format %q, expected key=value\n", kv)
				return fmt.Errorf("invalid header: %s", kv)
			}
			headers[parts[0]] = parts[1]
		}

		monitor := &spork.Monitor{
			Target:          rawURL,
			Name:            name,
			Type:            addType,
			Method:          addMethod,
			ExpectedStatus:  addExpectedStatus,
			Interval:        addInterval,
			Timeout:         addTimeout,
			Regions:         addRegions,
			Headers:         headers,
			Body:            addBody,
			Keyword:         addKeyword,
			KeywordType:     addKeywordType,
			SSLWarnDays:     addSSLWarnDays,
			AlertChannelIDs: addAlertChannels,
			Tags:            addTags,
		}

		result, err := client.CreateMonitor(context.Background(), monitor)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating monitor: %s\n", err)
			return err
		}

		fmt.Printf("✓ Monitor added: %s (%s)\n", result.Name, result.Target)
		fmt.Printf("  Checking every %ds\n", result.Interval)
		fmt.Println()
		fmt.Println("  View dashboard: https://sporkops.com")
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "human-readable name (defaults to hostname)")
	addCmd.Flags().StringVar(&addType, "type", "http", "monitor type: http, ssl, dns, keyword, tcp, ping")
	addCmd.Flags().StringVar(&addMethod, "method", "GET", "HTTP method (GET, HEAD, POST, PUT)")
	addCmd.Flags().IntVar(&addExpectedStatus, "expected-status", 200, "expected HTTP status code (100-599)")
	addCmd.Flags().IntVar(&addInterval, "interval", 60, "check interval in seconds (60-86400, multiple of 60)")
	addCmd.Flags().IntVar(&addTimeout, "timeout", 30, "timeout per check in seconds (5-120)")
	addCmd.Flags().StringSliceVar(&addRegions, "regions", nil, "check regions (us-central1, europe-west1)")
	addCmd.Flags().StringArrayVar(&addHeaders, "header", nil, "custom HTTP header as key=value (repeatable)")
	addCmd.Flags().StringVar(&addBody, "body", "", "HTTP request body for POST/PUT")
	addCmd.Flags().StringVar(&addKeyword, "keyword", "", "keyword to search in response (required for keyword type)")
	addCmd.Flags().StringVar(&addKeywordType, "keyword-type", "", "keyword match type: exists, not_exists")
	addCmd.Flags().IntVar(&addSSLWarnDays, "ssl-warn-days", 0, "days before SSL expiry to warn (default 14)")
	addCmd.Flags().StringSliceVar(&addAlertChannels, "alert-channels", nil, "alert channel IDs to notify")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "organization tags")
}
