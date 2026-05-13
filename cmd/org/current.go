package org

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/config"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:     "current",
	Aliases: []string{"whoami"},
	Short:   "Print the currently active organization",
	Long: `Print the organization ID the CLI will target by default.

Resolves the same precedence chain as every org-scoped command:
  --org flag  >  SPORK_ORG_ID env  >  persisted active org  >  auto-resolve

When no preference is set (everything empty) and the caller is an API
key, the SDK auto-resolves to the key's home org and that is printed
along with the source. For multi-org Firebase users with no preference,
this command tells you to pick one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, id := resolveActiveOrg()
		if id == "" {
			// No explicit preference and no API-key home org to resolve.
			// Tell the user how to fix it rather than print a blank line.
			fmt.Fprintln(os.Stderr, "No active organization.")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "  List candidates:  spork org list")
			fmt.Fprintln(os.Stderr, "  Pick one:         spork org use ORG_ID")
			fmt.Fprintln(os.Stderr, "  Or per-call:      --org ORG_ID  /  SPORK_ORG_ID=ORG_ID")
			return fmt.Errorf("no active organization")
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, map[string]string{
				"organization_id": id,
				"source":          source,
			})
		}
		fmt.Printf("%s\n", id)
		// Stderr keeps the source out of `spork org current | xargs ...`
		// pipelines — stdout stays a single ID for scriptability.
		fmt.Fprintf(os.Stderr, "(source: %s)\n", source)
		return nil
	},
}

// resolveActiveOrg returns the active org and a label describing where
// the value came from. The label is for the human-readable trailer; it's
// also the `source` field in --output json so scripts can branch on the
// origin if they want to.
func resolveActiveOrg() (source, id string) {
	if cmdutil.OrgID != "" {
		// cmdutil.OrgID merges --org and SPORK_ORG_ID at root-cmd
		// PreRun time, so we can't cheaply tell which one supplied
		// the value here. Both have the same precedence for users so
		// labelling as "flag-or-env" is honest.
		return "flag or env", cmdutil.OrgID
	}
	s, err := config.Load()
	if err == nil && s.ActiveOrganizationID != "" {
		return "config", s.ActiveOrganizationID
	}
	// Fall back to SDK auto-resolve (API-key home org). This costs one
	// HTTP request; users who don't want it should pin --org or set the
	// env var. We do not call ListMyOrgs ourselves — the SDK already
	// has the same logic with caching.
	client, err := cmdutil.RequireAuth()
	if err != nil {
		return "", ""
	}
	resolved, err := client.OrganizationID(context.Background())
	if err != nil || resolved == "" {
		return "", ""
	}
	return "auto-resolved", resolved
}
