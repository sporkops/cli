package monitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

// validMonitorTypes lists allowed monitor types for add and update commands.
// Note: "ping" here is the ICMP-style check type, which is distinct from the
// deprecated "ping" command alias (the command alias lives on Cmd.Aliases;
// the deprecation warning is emitted from the root command's
// PersistentPreRunE so it fires for every subcommand, not just the ones
// that inherit monitor.Cmd's pre-run hook).
var validMonitorTypes = map[string]bool{"http": true, "ssl": true, "dns": true, "keyword": true, "tcp": true, "ping": true}

// validHTTPMethods lists allowed HTTP methods for add and update commands.
var validHTTPMethods = map[string]bool{"GET": true, "HEAD": true, "POST": true, "PUT": true}

// Cmd is the `spork monitor` parent command. The historical name was
// `spork ping`; it is retained as an alias for one release cycle. The
// deprecation warning is emitted by the root command's PersistentPreRunE
// (see cmd/root.go) — attaching it here would shadow the root-level
// PersistentPreRunE because Cobra fires only the nearest pre-run hook in
// the command tree.
var Cmd = &cobra.Command{
	Use:     "monitor",
	Aliases: []string{"ping"},
	Short:   "Manage uptime monitors",
	Long:    "Create, list, and manage uptime monitors for your sites and APIs.",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(historyCmd)
	Cmd.AddCommand(pauseCmd)
	Cmd.AddCommand(unpauseCmd)
}

// resolveMonitorID resolves an ID-or-URL argument to a monitor ID.
// If the argument looks like a URL, it fetches all monitors and finds the match.
func resolveMonitorID(client *spork.Client, idOrURL string) (string, string, error) {
	if strings.Contains(idOrURL, "://") {
		// Looks like a URL — resolve via list
		monitors, err := client.ListMonitors(context.Background())
		if err != nil {
			return "", "", err
		}
		for _, m := range monitors {
			if m.Target == idOrURL {
				return m.ID, m.Name, nil
			}
		}
		return "", "", fmt.Errorf("no monitor found for URL: %s", idOrURL)
	}
	return idOrURL, "", nil
}
