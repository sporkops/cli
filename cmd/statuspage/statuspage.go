package statuspage

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/auth"
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

func requireAuth() (*api.Client, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "⚡ Login required")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Sign up or log in to get started:")
		fmt.Fprintln(os.Stderr, "    spork login")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  New to Spork? Sign up free:")
		fmt.Fprintln(os.Stderr, "    https://sporkops.com/signup?ref=cli")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Docs: https://sporkops.com/docs")
		return nil, fmt.Errorf("not logged in")
	}
	return api.NewClient(token), nil
}

func handleAPIError(err error) bool {
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.StatusCode {
	case 401:
		fmt.Fprintln(os.Stderr, "⚡ Session expired")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Log in again:")
		fmt.Fprintln(os.Stderr, "    spork login")
	case 402:
		fmt.Fprintln(os.Stderr, "💳 Subscription required")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Subscribe to a plan to get started:")
		fmt.Fprintln(os.Stderr, "    https://sporkops.com/billing?ref=cli")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Plans start at $4/mo.")
	case 403:
		fmt.Fprintln(os.Stderr, "Access denied.")
		fmt.Fprintln(os.Stderr)
		if apiErr.Message != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", apiErr.Message)
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintln(os.Stderr, "  https://sporkops.com/billing")
	default:
		return false
	}
	return true
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
