package ping

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/auth"
	"github.com/spf13/cobra"
)

// Cmd is the `spork ping` parent command.
var Cmd = &cobra.Command{
	Use:   "ping",
	Short: "Manage uptime monitors",
	Long:  "Add, list, and manage uptime monitors for your sites and APIs.",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(historyCmd)
	Cmd.AddCommand(pauseCmd)
	Cmd.AddCommand(unpauseCmd)
}

// requireAuth loads the stored token and returns an API client.
// If no token is found, it prints login instructions and returns an error.
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

// handleAPIError prints user-friendly messages for common API errors.
// Returns true if the error was handled (printed), false otherwise.
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
