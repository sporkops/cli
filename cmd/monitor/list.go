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
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all monitors",
	Example: `  spork monitor list
  spork monitor list --status up
  spork monitor list --limit 10
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

		// Defense in depth: server already honored filters via
		// ListOptions.Filters (spork-go v0.4.1+). Re-apply locally so
		// older SDKs or upstream misconfig can't leak unfiltered rows
		// into structured output.
		if listFilterStatus != "" || listFilterType != "" {
			monitors = filterMonitors(monitors, listFilterStatus, listFilterType)
		}

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
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "max number of monitors to show (0 = no cap, auto-paginate through all)")
}

// fetchMonitorsForListing resolves the caller's pagination intent into the
// right SDK call. Two modes, layered on top of server-side filters:
//
//   - no flags  → auto-paginate through every page (ListMonitors).
//   - --limit N → auto-paginate but stop once we have N items.
//
// --status and --type are pushed to the server via ListOptions.Filters.
// Filtering happens server-side, so `--limit 5 --status down` returns
// "up to 5 down monitors" rather than "up to 5 monitors, of which some
// happen to be down".
func fetchMonitorsForListing(client *spork.Client) ([]spork.Monitor, error) {
	ctx := context.Background()
	filters := listFilters()

	// No cap, no filters → the zero-config path; plain auto-paginator.
	if listLimit <= 0 && len(filters) == 0 {
		return client.ListMonitors(ctx)
	}

	// Otherwise: hand-rolled auto-pagination so we can (a) cap at --limit
	// and/or (b) push filters to the server.
	var collected []spork.Monitor
	opts := spork.ListOptions{Filters: filters}
	if listLimit > 0 && listLimit < 100 {
		// Small --limit → request a small first page so we don't fetch 100
		// records just to discard 95 of them.
		opts.Limit = listLimit
	}
	for {
		page, info, err := client.ListMonitorsWithOptions(ctx, opts)
		if err != nil {
			return nil, err
		}
		collected = append(collected, page...)
		if listLimit > 0 && len(collected) >= listLimit {
			return collected[:listLimit], nil
		}
		if !info.HasMore {
			return collected, nil
		}
		opts.Cursor = info.NextCursor
	}
}

// listFilters collects the --status and --type flag values into the
// ListOptions.Filters map. Empty values are omitted; the SDK drops them
// anyway but keeping the map compact makes --debug output readable.
func listFilters() map[string]string {
	if listFilterStatus == "" && listFilterType == "" {
		return nil
	}
	f := make(map[string]string, 2)
	if listFilterStatus != "" {
		f["status"] = listFilterStatus
	}
	if listFilterType != "" {
		f["type"] = listFilterType
	}
	return f
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
