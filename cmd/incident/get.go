package incident

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show details of an incident",
	Long: `Show full details of a single incident by ID.

Example:
  spork incident get inc_abc123
  spork incident get inc_abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		inc, err := client.GetIncident(args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching incident: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(inc)
		}

		fmt.Printf("%-20s %s\n", "ID:", inc.ID)
		fmt.Printf("%-20s %s\n", "Title:", inc.Title)
		fmt.Printf("%-20s %s\n", "Type:", inc.Type)
		fmt.Printf("%-20s %s\n", "Status:", inc.Status)
		fmt.Printf("%-20s %s\n", "Impact:", inc.Impact)
		fmt.Printf("%-20s %s\n", "Status Page:", inc.StatusPageID)

		if len(inc.ComponentIDs) > 0 {
			fmt.Printf("%-20s %s\n", "Components:", strings.Join(inc.ComponentIDs, ", "))
		}
		if inc.Message != "" {
			fmt.Printf("%-20s %s\n", "Message:", inc.Message)
		}
		if inc.StartedAt != "" {
			fmt.Printf("%-20s %s\n", "Started:", inc.StartedAt)
		}
		if inc.ResolvedAt != "" {
			fmt.Printf("%-20s %s\n", "Resolved:", inc.ResolvedAt)
		}
		if inc.ScheduledStart != "" {
			fmt.Printf("%-20s %s\n", "Scheduled Start:", inc.ScheduledStart)
		}
		if inc.ScheduledEnd != "" {
			fmt.Printf("%-20s %s\n", "Scheduled End:", inc.ScheduledEnd)
		}
		fmt.Printf("%-20s %s\n", "Created:", inc.CreatedAt)
		fmt.Printf("%-20s %s\n", "Updated:", inc.UpdatedAt)

		return nil
	},
}
