package ping

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/api"
)

// Cmd is the `spork ping` parent command.
var Cmd = &cobra.Command{
	Use:   "ping",
	Short: "Manage uptime monitors",
	Long:  "Add, list, and manage uptime monitors for your sites and APIs.",
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
func resolveMonitorID(client *api.Client, idOrURL string) (string, string, error) {
	if strings.Contains(idOrURL, "://") {
		// Looks like a URL — resolve via list
		monitors, err := client.ListMonitors()
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
