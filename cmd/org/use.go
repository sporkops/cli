package org

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/config"
	"github.com/spf13/cobra"
)

var clearFlag bool

var useCmd = &cobra.Command{
	Use:   "use ORG_ID",
	Short: "Switch the active organization",
	Long: `Persist an organization ID as the CLI default so org-scoped
commands stop needing --org or SPORK_ORG_ID.

The chosen org must be one your account belongs to — ` + "`use`" + ` calls
` + "`/users/me/orgs`" + ` to verify membership before saving, so a typo
fails fast instead of silently configuring a non-existent tenant.

Run with --clear to remove the saved preference and fall back to
auto-resolve.

Examples:
  spork org use org_8f4c2a91
  spork org use --clear`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if clearFlag {
			if len(args) > 0 {
				return fmt.Errorf("--clear takes no arguments")
			}
			if err := config.ClearActiveOrganization(); err != nil {
				return fmt.Errorf("clear active org: %w", err)
			}
			fmt.Fprintln(os.Stderr, "Active organization cleared.")
			fmt.Fprintln(os.Stderr, "Org-scoped commands will fall back to SPORK_ORG_ID or auto-resolve.")
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify an organization ID (or pass --clear). Run \"spork org list\" to see candidates")
		}
		orgID := args[0]

		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		// Verify membership before saving. A typo here would mean every
		// subsequent command 404s or 403s with `not_a_member`; better to
		// catch it now with a clear message. ListMyOrgs is the cheapest
		// check available and stays consistent with `spork org list`.
		orgs, err := client.ListMyOrgs(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error verifying organization: %s\n", err)
			return err
		}
		var found bool
		var matchedName string
		for _, o := range orgs {
			if o.ID == orgID {
				found = true
				matchedName = o.Name
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Organization %q is not one of yours.\n", orgID)
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Run \"spork org list\" to see the IDs you can choose.")
			return fmt.Errorf("not a member of %s", orgID)
		}

		if err := config.SetActiveOrganization(orgID); err != nil {
			return fmt.Errorf("save active org: %w", err)
		}

		display := orgID
		if matchedName != "" {
			display = fmt.Sprintf("%s (%s)", matchedName, orgID)
		}
		fmt.Fprintf(os.Stderr, "Active organization set to %s.\n", display)
		return nil
	},
}

func init() {
	useCmd.Flags().BoolVar(&clearFlag, "clear", false, "remove the saved active organization preference")
}
