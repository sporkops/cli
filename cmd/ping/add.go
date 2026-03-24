package ping

import (
	"fmt"
	"net/url"
	"os"

	"github.com/sporkops/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	addName     string
	addMethod   string
	addInterval int
)

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a new uptime monitor",
	Long:  "Add a new uptime monitor for the given URL.\n\nExample:\n  spork ping add https://example.com\n  spork ping add https://api.example.com/health --name \"API Health\" --interval 30",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		rawURL := args[0]
		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			fmt.Fprintln(os.Stderr, "Error: URL must start with http:// or https://")
			return fmt.Errorf("invalid URL: %s", rawURL)
		}

		name := addName
		if name == "" {
			name = parsed.Hostname()
		}

		monitor := &api.Monitor{
			Target:   rawURL,
			Name:     name,
			Type:     "http",
			Method:   addMethod,
			Interval: addInterval,
		}

		result, err := client.CreateMonitor(monitor)
		if err != nil {
			if handleAPIError(err) {
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
	addCmd.Flags().StringVar(&addMethod, "method", "GET", "HTTP method")
	addCmd.Flags().IntVar(&addInterval, "interval", 60, "check interval in seconds (60 or 30)")
}
