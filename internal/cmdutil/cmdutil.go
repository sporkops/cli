package cmdutil

import (
	"errors"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/auth"
	"github.com/sporkops/cli/pkg/spork"
)

// RequireAuth loads the stored token and returns an API client.
// If no token is found, it prints login instructions and returns an error.
func RequireAuth() (*spork.Client, error) {
	token, err := auth.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "Login required")
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
	return spork.NewClient(spork.WithAPIKey(token), spork.WithUserAgent("spork-cli/"+spork.Version)), nil
}

// HandleAPIError prints user-friendly messages for common API errors.
// Returns true if the error was handled (printed), false otherwise.
func HandleAPIError(err error) bool {
	if spork.IsUnauthorized(err) {
		fmt.Fprintln(os.Stderr, "Session expired")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Log in again:")
		fmt.Fprintln(os.Stderr, "    spork login")
		return true
	}
	if spork.IsPaymentRequired(err) {
		fmt.Fprintln(os.Stderr, "Subscription required")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Subscribe to a plan to get started:")
		fmt.Fprintln(os.Stderr, "    https://sporkops.com/billing?ref=cli")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  Plans start at $4/mo.")
		return true
	}
	if spork.IsForbidden(err) {
		fmt.Fprintln(os.Stderr, "Access denied.")
		fmt.Fprintln(os.Stderr)
		var apiErr *spork.APIError
		if errors.As(err, &apiErr) && apiErr.Message != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", apiErr.Message)
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintln(os.Stderr, "  https://sporkops.com/billing")
		return true
	}
	return false
}

