package apikey

import (
	"errors"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/auth"
	"github.com/spf13/cobra"
)

// Cmd is the `spork api-key` parent command.
var Cmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API keys",
	Long:  "Create, list, and delete API keys for programmatic access.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deleteCmd)
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
		fmt.Fprintln(os.Stderr, "💳 Payment method required")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Add a payment method: https://sporkops.com/billing?ref=cli")
	case 403:
		fmt.Fprintln(os.Stderr, "Access denied.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Contact support: https://sporkops.com/support")
	default:
		return false
	}
	return true
}
