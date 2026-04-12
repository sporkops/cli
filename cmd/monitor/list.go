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
	listCmd.Flags().IntVar(&listPage, "page", 0, "fetch only this 1-indexed page instead of auto-paginating")
	listCmd.Flags().IntVar(&listPageSize, "page-size", 0, "page size (1-100, default 100) when --page or --limit are set")
}

// fetchMonitorsForListing resolves the caller's pagination intent into the
// right SDK call. Four modes, layered on top of server-side filters:
//
//   - --page N   → one explicit page (uses ListMonitorsPage).
//   - --limit N  → auto-paginate but stop once we have N items.
//   - neither    → auto-paginate through every page (ListMonitors).
//
// --page-size controls per_page on the server. It is only consulted when
// one of --page or --limit is set; otherwise the SDK default applies.
//
// --status and --type are pushed to the server via ListOptions.Filters.
// Filtering happens server-side, so `--limit 5 --status down` now returns
// "up to 5 down monitors" rather than "up to 5 monitors, of which some
// happen to be down".
func fetchMonitorsForListing(client *spork.Client) ([]spork.Monitor, error) {
	ctx := context.Background()
	filters := listFilters()

	// Explicit single page: one request, server-side filtered.
	if listPage > 0 {
		opts := spork.ListOptions{Page: listPage, PerPage: listPageSize, Filters: filters}
		monitors, _, err := client.ListMonitorsPage(ctx, opts)
		return monitors, err
	}

	// No cap, no filters → the zero-config path; plain auto-paginator.
	if listLimit <= 0 && len(filters) == 0 {
		return client.ListMonitors(ctx)
	}

	// Otherwise: hand-rolled auto-pagination so we can (a) cap at --limit
	// and/or (b) pass filters down. The loop's termination uses PageMeta
	// HasMore, which respects both the server-reported total and the
	// short-page heuristic for endpoints that don't populate Total.
	var collected []spork.Monitor
	opts := spork.ListOptions{Page: 1, PerPage: listPageSize, Filters: filters}
	if listLimit > 0 && (opts.PerPage == 0 || listLimit < opts.PerPage) {
		// Small --limit → request a small first page so we don't fetch 100
		// records just to discard 95 of them.
		opts.PerPage = listLimit
	}
	for {
		page, meta, err := client.ListMonitorsPage(ctx, opts)
		if err != nil {
			return nil, err
		}
		collected = append(collected, page...)
		if listLimit > 0 && len(collected) >= listLimit {
			return collected[:listLimit], nil
		}
		if !meta.HasMore(len(page)) {
			return collected, nil
		}
		opts.Page = meta.Page + 1
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
