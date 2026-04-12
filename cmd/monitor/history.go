package monitor

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history <id|url>",
	Short: "Show recent check results for a monitor",
	Long:  "Show recent uptime check results for a monitor.\n\nExample:\n  spork monitor history https://example.com\n  spork monitor history abc123 --limit 50",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id, _, err := resolveMonitorID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		results, err := client.GetMonitorResults(context.Background(), id, historyLimit)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching results: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(results)
		}

		if len(results) == 0 {
			fmt.Println("No check results yet.")
			return nil
		}

		headers := []string{"TIME", "STATUS", "RESPONSE TIME", "STATUS CODE", "REGION"}
		rows := make([][]string, len(results))
		for i, r := range results {
			rows[i] = []string{
				r.CheckedAt,
				output.ColorStatus(r.Status),
				strconv.FormatInt(r.ResponseTimeMs, 10) + "ms",
				strconv.Itoa(r.StatusCode),
				r.Region,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	historyCmd.Flags().IntVar(&historyLimit, "limit", 20, "number of results to show")
}
