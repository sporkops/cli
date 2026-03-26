package ping

import (
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status of all monitors",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		monitors, err := client.ListMonitors()
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching status: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(monitors)
		}

		if len(monitors) == 0 {
			fmt.Println("No monitors yet. Add one:")
			fmt.Println("  spork ping add <url>")
			return nil
		}

		headers := []string{"STATUS", "NAME", "TARGET", "LAST CHECKED"}
		rows := make([][]string, len(monitors))
		for i, m := range monitors {
			lastChecked := m.LastCheckedAt
			if lastChecked == "" {
				lastChecked = "-"
			}
			rows[i] = []string{
				output.ColorStatus(m.Status),
				m.Name,
				m.Target,
				lastChecked,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
