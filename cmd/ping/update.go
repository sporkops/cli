package ping

import (
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/api"
	"github.com/spf13/cobra"
)

var (
	updateName     string
	updateTarget   string
	updateMethod   string
	updateInterval int
	updatePaused   string
)

var updateCmd = &cobra.Command{
	Use:   "update <id|url>",
	Short: "Update an existing monitor",
	Long:  "Update an existing monitor's settings.\n\nExample:\n  spork ping update https://example.com --name \"New Name\"\n  spork ping update abc123 --interval 300 --paused true",
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
			update.Target = updateTarget
			hasChanges = true
		}
		if cmd.Flags().Changed("method") {
			update.Method = updateMethod
			hasChanges = true
		}
		if cmd.Flags().Changed("interval") {
			update.Interval = updateInterval
			hasChanges = true
		}
		if cmd.Flags().Changed("paused") {
			update.Paused = updatePaused == "true"
			hasChanges = true
		}

		if !hasChanges {
			fmt.Fprintln(os.Stderr, "Nothing to update. Specify at least one flag:")
			fmt.Fprintln(os.Stderr, "  --name, --target, --method, --interval, --paused")
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
	updateCmd.Flags().StringVar(&updateMethod, "method", "", "HTTP method")
	updateCmd.Flags().IntVar(&updateInterval, "interval", 0, "check interval in seconds (60-3600)")
	updateCmd.Flags().StringVar(&updatePaused, "paused", "", "pause or unpause (true/false)")
}
