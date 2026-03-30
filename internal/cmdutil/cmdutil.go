package cmdutil

import (
	"errors"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/auth"
)

// RequireAuth loads the stored token and returns an API client.
// If no token is found, it prints login instructions and returns an error.
func RequireAuth() (*api.Client, error) {
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

// HandleAPIError prints user-friendly messages for common API errors.
// Returns true if the error was handled (printed), false otherwise.
func HandleAPIError(err error) bool {
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
