package ping

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
)

var (
	statusFilterStatus string
	statusFilterType   string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status of all monitors",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		monitors, err := client.ListMonitors()
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching status: %s\n", err)
			return err
		}

		monitors = filterMonitors(monitors, statusFilterStatus, statusFilterType)

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(monitors)
		}

		if len(monitors) == 0 {
			fmt.Println("No monitors yet. Add one:")
			fmt.Println("  spork ping add <url>")
			return nil
		}

		headers := []string{"STATUS", "NAME", "TYPE", "TARGET", "INTERVAL", "LAST CHECKED"}
		rows := make([][]string, len(monitors))
		for i, m := range monitors {
			lastChecked := m.LastCheckedAt
			if lastChecked == "" {
				lastChecked = "-"
			}
			rows[i] = []string{
				output.ColorStatus(m.Status),
				m.Name,
				m.Type,
				m.Target,
				strconv.Itoa(m.Interval) + "s",
				lastChecked,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusFilterStatus, "status", "", "filter by status (up, down, degraded, paused, pending)")
	statusCmd.Flags().StringVar(&statusFilterType, "type", "", "filter by monitor type")
}
