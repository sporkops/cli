package monitor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

// validMonitorTypes lists allowed monitor types for add and update commands.
// Note: "ping" here is the ICMP-style check type, which is distinct from the
// deprecated "ping" command alias below.
var validMonitorTypes = map[string]bool{"http": true, "ssl": true, "dns": true, "keyword": true, "tcp": true, "ping": true}

// validHTTPMethods lists allowed HTTP methods for add and update commands.
var validHTTPMethods = map[string]bool{"GET": true, "HEAD": true, "POST": true, "PUT": true}

// Cmd is the `spork monitor` parent command. The historical name was
// `spork ping`; it is retained as an alias for one release cycle and emits a
// deprecation warning when used.
var Cmd = &cobra.Command{
	Use:     "monitor",
	Aliases: []string{"ping"},
	Short:   "Manage uptime monitors",
	Long:    "Add, list, and manage uptime monitors for your sites and APIs.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Cobra does not expose CalledAs() on parent commands, so inspect the
		// raw argv for the first non-flag token. If the user typed `ping`,
		// emit a deprecation warning on stderr.
		if firstPositional(os.Args[1:]) == "ping" {
			fmt.Fprintln(os.Stderr, "warning: `spork ping` is deprecated and will be removed in a future release; use `spork monitor` instead")
		}
	},
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

// firstPositional returns the first non-flag argument from argv, or "" if
// every argument is a flag or a flag value. Handles both `--foo=bar` and
// `--foo bar` forms by conservatively skipping the token following any flag
// that does not contain `=`. Unknown flags are treated as potentially taking
// a value (same conservative behavior Cobra uses internally).
func firstPositional(argv []string) string {
	i := 0
	for i < len(argv) {
		a := argv[i]
		if !strings.HasPrefix(a, "-") {
			return a
		}
		// If the flag has an inline value (`--foo=bar`) or is the double-dash
		// terminator, don't skip the next token.
		if a == "--" || strings.Contains(a, "=") {
			i++
			continue
		}
		// Skip the flag and, conservatively, its next token (a potential value).
		i += 2
	}
	return ""
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
