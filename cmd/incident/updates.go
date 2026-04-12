package incident

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var updatesCmd = &cobra.Command{
	Use:   "updates <incident-id>",
	Short: "List timeline updates for an incident",
	Long: `Show all timeline updates for an incident.

Example:
  spork incident updates inc_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		updates, err := client.ListIncidentUpdates(context.Background(), args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing incident updates: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, updates)
		}

		if len(updates) == 0 {
			fmt.Println("No updates yet.")
			return nil
		}

		headers := []string{"ID", "STATUS", "MESSAGE", "CREATED"}
		rows := make([][]string, len(updates))
		for i, u := range updates {
			msg := u.Message
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			rows[i] = []string{
				u.ID,
				u.Status,
				msg,
				u.CreatedAt,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
