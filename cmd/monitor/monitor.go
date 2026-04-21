package monitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

// validMonitorTypes lists allowed monitor types for add and update commands.
// "ping" here is the ICMP-style check type, not a command name.
var validMonitorTypes = map[string]bool{"http": true, "ssl": true, "dns": true, "keyword": true, "tcp": true, "ping": true}

// validHTTPMethods lists allowed HTTP methods for add and update commands.
var validHTTPMethods = map[string]bool{"GET": true, "HEAD": true, "POST": true, "PUT": true}

// Cmd is the `spork monitor` parent command.
var Cmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage uptime monitors",
	Long:  "Create, list, and manage uptime monitors for your sites and APIs.",
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
