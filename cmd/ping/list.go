package ping

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	listFilterStatus string
	listFilterType   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all monitors",
	Example: `  spork ping list
  spork ping list --status up
  spork ping list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		monitors, err := client.ListMonitors(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing monitors: %s\n", err)
			return err
		}

		monitors = filterMonitors(monitors, listFilterStatus, listFilterType)

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(monitors)
		}

		if len(monitors) == 0 {
			fmt.Println("No monitors yet. Add one:")
			fmt.Println("  spork ping add <url>")
			return nil
		}

		headers := []string{"ID", "NAME", "TYPE", "TARGET", "INTERVAL", "STATUS"}
		rows := make([][]string, len(monitors))
		for i, m := range monitors {
			rows[i] = []string{
				m.ID,
				m.Name,
				m.Type,
				m.Target,
				strconv.Itoa(m.Interval) + "s",
				output.ColorStatus(m.Status),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listFilterStatus, "status", "", "filter by status (up, down, degraded, paused, pending)")
	listCmd.Flags().StringVar(&listFilterType, "type", "", "filter by monitor type")
}

// filterMonitors applies client-side status and type filters.
func filterMonitors(monitors []spork.Monitor, status, monitorType string) []spork.Monitor {
	if status == "" && monitorType == "" {
		return monitors
	}
	filtered := make([]spork.Monitor, 0, len(monitors))
	for _, m := range monitors {
		if status != "" && m.Status != status {
			continue
		}
		if monitorType != "" && m.Type != monitorType {
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}
