package maintenance

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	listStateFilter string
	listLimit       int
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List maintenance windows",
	Example: `  spork maintenance list
  spork maintenance list --state scheduled
  spork maintenance list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		windows, err := client.ListMaintenanceWindows(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing maintenance windows: %s\n", err)
			return err
		}

		if listStateFilter != "" {
			filtered := windows[:0]
			for _, w := range windows {
				if w.State == listStateFilter {
					filtered = append(filtered, w)
				}
			}
			windows = filtered
		}
		if listLimit > 0 && len(windows) > listLimit {
			windows = windows[:listLimit]
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, windows)
		}

		if len(windows) == 0 {
			fmt.Println("No maintenance windows yet. Create one:")
			fmt.Println("  spork maintenance create --name <name> --all-monitors \\")
			fmt.Println("    --timezone <IANA> --start <RFC3339> --end <RFC3339>")
			return nil
		}

		headers := []string{"ID", "NAME", "STATE", "START", "END", "RECURRENCE", "TZ"}
		rows := make([][]string, len(windows))
		for i, w := range windows {
			recurrence := w.RecurrenceType
			if recurrence == "" {
				recurrence = "once"
			}
			rows[i] = []string{
				w.ID,
				w.Name,
				output.ColorStatus(w.State),
				w.StartAt,
				w.EndAt,
				recurrence,
				w.Timezone,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStateFilter, "state", "", "filter by state (scheduled, in_progress, completed, cancelled)")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "max number of windows to show (0 = no cap)")
}
