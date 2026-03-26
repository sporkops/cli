package ping

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all monitors",
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
			fmt.Fprintf(os.Stderr, "Error listing monitors: %s\n", err)
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

		headers := []string{"ID", "NAME", "TARGET", "INTERVAL", "STATUS"}
		rows := make([][]string, len(monitors))
		for i, m := range monitors {
			rows[i] = []string{
				m.ID,
				m.Name,
				m.Target,
				strconv.Itoa(m.Interval) + "s",
				output.ColorStatus(m.Status),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
