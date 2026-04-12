package monitor

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
	listLimit        int
	listPage         int
	listPageSize     int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all monitors",
	Example: `  spork monitor list
  spork monitor list --status up
  spork monitor list --limit 10
  spork monitor list --page 2 --page-size 25
  spork monitor list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		monitors, err := fetchMonitorsForListing(client)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing monitors: %s\n", err)
			return err
		}

		monitors = filterMonitors(monitors, listFilterStatus, listFilterType)

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, monitors)
		}

		if len(monitors) == 0 {
			fmt.Println("No monitors yet. Add one:")
			fmt.Println("  spork monitor add <url>")
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
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "stop after N monitors (0 means no limit — auto-paginate all)")
	listCmd.Flags().IntVar(&listPage, "page", 0, "fetch only this 1-indexed page instead of auto-paginating")
	listCmd.Flags().IntVar(&listPageSize, "page-size", 0, "page size (1-100, default 100) when --page or --limit are set")
}

// fetchMonitorsForListing resolves the caller's pagination intent into the
// right SDK call. Three modes:
//
//   - --page N  → one explicit page (uses ListMonitorsPage).
//   - --limit N → auto-paginate but stop once we have N items.
//   - neither   → auto-paginate through every page (ListMonitors default).
//
// --page-size controls per_page on the server. It is only consulted when one
// of --page or --limit is set; otherwise we trust the SDK default.
func fetchMonitorsForListing(client *spork.Client) ([]spork.Monitor, error) {
	ctx := context.Background()

	// Explicit single page.
	if listPage > 0 {
		opts := spork.ListOptions{Page: listPage, PerPage: listPageSize}
		monitors, _, err := client.ListMonitorsPage(ctx, opts)
		return monitors, err
	}

	// No cap → use the plain auto-paginator.
	if listLimit <= 0 {
		return client.ListMonitors(ctx)
	}

	// Capped auto-pagination: loop ListMonitorsPage until we hit the limit or
	// run out of pages. Matches how Stripe / PagerDuty CLIs behave under
	// --limit.
	var collected []spork.Monitor
	opts := spork.ListOptions{Page: 1, PerPage: listPageSize}
	for {
		page, meta, err := client.ListMonitorsPage(ctx, opts)
		if err != nil {
			return nil, err
		}
		collected = append(collected, page...)
		if len(collected) >= listLimit {
			return collected[:listLimit], nil
		}
		if !meta.HasMore(len(page)) {
			return collected, nil
		}
		opts.Page = meta.Page + 1
	}
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
