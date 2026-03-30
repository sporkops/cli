package ping

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
)

var getCmd = &cobra.Command{
	Use:   "get <id|url>",
	Short: "Show details of a monitor",
	Long: `Show full details of a single monitor by ID or URL.

Example:
  spork ping get abc123
  spork ping get https://example.com
  spork ping get abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id, _, err := resolveMonitorID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		monitor, err := client.GetMonitor(id)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching monitor: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(monitor)
		}

		// Detailed key-value output showing all fields.
		fmt.Printf("%-20s %s\n", "ID:", monitor.ID)
		fmt.Printf("%-20s %s\n", "Name:", monitor.Name)
		fmt.Printf("%-20s %s\n", "Target:", monitor.Target)
		fmt.Printf("%-20s %s\n", "Type:", monitor.Type)
		fmt.Printf("%-20s %s\n", "Status:", output.ColorStatus(monitor.Status))
		fmt.Printf("%-20s %s\n", "Method:", monitor.Method)
		fmt.Printf("%-20s %d\n", "Expected Status:", monitor.ExpectedStatus)
		fmt.Printf("%-20s %ds\n", "Interval:", monitor.Interval)
		fmt.Printf("%-20s %ds\n", "Timeout:", monitor.Timeout)

		paused := false
		if monitor.Paused != nil {
			paused = *monitor.Paused
		}
		fmt.Printf("%-20s %s\n", "Paused:", strconv.FormatBool(paused))

		if len(monitor.Regions) > 0 {
			fmt.Printf("%-20s %s\n", "Regions:", strings.Join(monitor.Regions, ", "))
		}
		if len(monitor.Tags) > 0 {
			fmt.Printf("%-20s %s\n", "Tags:", strings.Join(monitor.Tags, ", "))
		}
		if len(monitor.AlertChannelIDs) > 0 {
			fmt.Printf("%-20s %s\n", "Alert Channels:", strings.Join(monitor.AlertChannelIDs, ", "))
		}
		if monitor.Keyword != "" {
			fmt.Printf("%-20s %s\n", "Keyword:", monitor.Keyword)
			fmt.Printf("%-20s %s\n", "Keyword Type:", monitor.KeywordType)
		}
		if monitor.SSLWarnDays > 0 {
			fmt.Printf("%-20s %d days\n", "SSL Warn Days:", monitor.SSLWarnDays)
		}
		if len(monitor.Headers) > 0 {
			fmt.Printf("%-20s\n", "Headers:")
			for k, v := range monitor.Headers {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		if monitor.Body != "" {
			fmt.Printf("%-20s %s\n", "Body:", monitor.Body)
		}

		lastChecked := monitor.LastCheckedAt
		if lastChecked == "" {
			lastChecked = "-"
		}
		fmt.Printf("%-20s %s\n", "Last Checked:", lastChecked)
		fmt.Printf("%-20s %s\n", "Created:", monitor.CreatedAt)
		fmt.Printf("%-20s %s\n", "Updated:", monitor.UpdatedAt)

		return nil
	},
}
