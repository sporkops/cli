package cmdutil

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/auth"
	"github.com/sporkops/cli/internal/debughttp"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

// Debug is set by the root command from --debug / SPORK_DEBUG. When true,
// RequireAuth attaches debughttp.Transport as an SDK HTTP middleware so
// every request and response is traced to stderr (with Authorization
// redacted).
//
// A package-level variable is used rather than threading a flag through
// every call site to keep the change surgical. The CLI already imports
// cmdutil from every command; there is no graceful way to pass the debug
// state through spork.NewClient without introducing a config struct that
// is a follow-up refactor.
var Debug bool

// OrgID is set by the root command from --org / SPORK_ORG_ID. When
// non-empty it is forwarded to the SDK via spork.WithOrganization so
// every org-scoped call lands on the named tenant. When empty the SDK
// auto-resolves on first use by listing /users/me/orgs — the right
// behaviour for API keys (which are bound to one org) and for users
// who only belong to one organization. Users in multiple orgs should
// pass --org explicitly to avoid the auto-resolve picking arbitrarily.
var OrgID string

// debugMiddleware returns a spork.HTTPMiddleware that wraps the SDK's
// transport in debughttp.Transport. We use middleware (v0.4.0+) rather
// than replacing the whole http.Client via spork.WithHTTPClient so the
// SDK's CheckRedirect logic — which strips Authorization on cross-host
// redirects — stays in place even when --debug is on.
func debugMiddleware() spork.HTTPMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return debughttp.NewTransport(next, nil)
	}
}

// StructuredOutput reports whether the user requested structured output
// (JSON or YAML) via --json or --output, and if so which format to emit.
// --json always wins over --output json|yaml for backwards compatibility
// with scripts that piped CLI output before --output existed.
func StructuredOutput(cmd *cobra.Command) (bool, string) {
	r := cmd.Root()
	if f := r.Flag("json"); f != nil && f.Changed {
		return true, "json"
	}
	if f := r.Flag("output"); f != nil {
		switch f.Value.String() {
		case "json":
			return true, "json"
		case "yaml":
			return true, "yaml"
		}
	}
	return false, ""
}

// PrintStructured prints v in the format requested by the current command.
// For commands that have already determined JSON-vs-YAML elsewhere, prefer
// output.PrintJSON / output.PrintYAML directly.
func PrintStructured(cmd *cobra.Command, v any) error {
	_, format := StructuredOutput(cmd)
	if format == "yaml" {
		return output.PrintYAML(v)
	}
	return output.PrintJSON(v)
}

// Structured reports whether structured (JSON or YAML) output was requested.
// Exists primarily to keep call sites terse: they can write
//
//	if cmdutil.Structured(cmd) { return cmdutil.PrintStructured(cmd, v) }
//
// without having to destructure the two-value StructuredOutput result.
//
// Also used as a "non-interactive" signal by destructive commands that
// default to a TTY confirmation prompt — if the user asked for JSON or
// YAML, they are scripting, not sitting at a prompt.
func Structured(cmd *cobra.Command) bool {
	ok, _ := StructuredOutput(cmd)
	return ok
}

// ParseKeyValue splits a "key=value" string and returns the key and value.
// It returns an error if the input is not in "key=value" format.
func ParseKeyValue(input string) (string, string, error) {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format %q, expected key=value", input)
	}
	return parts[0], parts[1], nil
}

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
	opts := []spork.Option{
		spork.WithAPIKey(token),
		spork.WithUserAgent("spork-cli/" + spork.Version),
	}
	if Debug {
		opts = append(opts, spork.WithHTTPMiddleware(debugMiddleware()))
	}
	if OrgID != "" {
		opts = append(opts, spork.WithOrganization(OrgID))
	}
	return spork.NewClient(opts...), nil
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

