package incident

import (
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/output"
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
		client, err := requireAuth()
		if err != nil {
			return err
		}

		updates, err := client.ListIncidentUpdates(args[0])
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing incident updates: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(updates)
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
