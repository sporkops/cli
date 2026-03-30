package incident

import (
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var listStatusPage string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List incidents for a status page",
	Long: `List all incidents for a status page.

Example:
  spork incident list --status-page sp_abc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		incidents, err := client.ListIncidents(listStatusPage)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing incidents: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(incidents)
		}

		if len(incidents) == 0 {
			fmt.Println("No incidents yet.")
			return nil
		}

		headers := []string{"ID", "TITLE", "TYPE", "STATUS", "IMPACT", "CREATED"}
		rows := make([][]string, len(incidents))
		for i, inc := range incidents {
			rows[i] = []string{
				inc.ID,
				inc.Title,
				inc.Type,
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
	listCmd.Flags().StringVar(&listStatusPage, "status-page", "", "status page ID (required)")
	listCmd.MarkFlagRequired("status-page")
}
