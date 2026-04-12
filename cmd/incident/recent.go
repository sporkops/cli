package incident

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var recentLimit int

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "List recent incidents across all your status pages",
	Long: `Show the most recent incidents across every status page in your organization.

Newest-first. Useful for a quick at-a-glance operational status check.

Examples:
  spork incident recent
  spork incident recent --limit 50`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		incidents, err := client.ListRecentIncidents(context.Background(), recentLimit)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing recent incidents: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(incidents)
		}

		if len(incidents) == 0 {
			fmt.Println("No incidents yet.")
			return nil
		}

		headers := []string{"ID", "TITLE", "STATUS", "IMPACT", "CREATED"}
		rows := make([][]string, len(incidents))
		for i, inc := range incidents {
			title := inc.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			rows[i] = []string{
				inc.ID,
				title,
				inc.Status,
				inc.Impact,
				inc.CreatedAt,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	recentCmd.Flags().IntVar(&recentLimit, "limit", 20, "max number of incidents to show (server caps at 100)")
}
