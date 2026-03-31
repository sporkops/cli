package statuspage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
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

func resolveStatusPageID(client *spork.Client, nameOrID string) (string, string, error) {
	if !strings.Contains(nameOrID, " ") {
		sp, err := client.GetStatusPage(context.Background(), nameOrID)
		if err == nil {
			return sp.ID, sp.Name, nil
		}
		var apiErr *spork.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// Fall through to name search
		} else {
			return "", "", err
		}
	}
	pages, err := client.ListStatusPages(context.Background())
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
