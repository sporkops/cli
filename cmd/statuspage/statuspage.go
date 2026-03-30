package statuspage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/api"
)

// Cmd is the `spork status-page` parent command.
var Cmd = &cobra.Command{
	Use:     "status-page",
	Aliases: []string{"sp"},
	Short:   "Manage status pages",
	Long:    "Create, list, and manage public status pages for your monitors.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
}

func resolveStatusPageID(client *api.Client, nameOrID string) (string, string, error) {
	if !strings.Contains(nameOrID, " ") {
		sp, err := client.GetStatusPage(nameOrID)
		if err == nil {
			return sp.ID, sp.Name, nil
		}
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// Fall through to name search
		} else {
			return "", "", err
		}
	}
	pages, err := client.ListStatusPages()
	if err != nil {
		return "", "", err
	}
	for _, p := range pages {
		if p.Name == nameOrID || p.Slug == nameOrID {
			return p.ID, p.Name, nil
		}
	}
	return "", "", fmt.Errorf("no status page found for: %s", nameOrID)
}
