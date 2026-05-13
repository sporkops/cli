// Package org implements the `spork org` subcommand tree.
//
// These commands manage which organization the CLI targets when --org
// and SPORK_ORG_ID are unset. Multi-org users (Agency-plan owners,
// contractors who belong to several customer orgs) want to switch
// without re-typing the ID on every invocation; this package gives
// them `spork org use <id>` to persist the choice, `spork org list`
// to discover candidates, and `spork org current` to see which one
// is active.
package org

import "github.com/spf13/cobra"

// Cmd is the parent command for organization management. Singular Use
// matches the rest of the CLI ("monitor", "member", "alert-channel");
// "orgs" is kept as an alias because `spork orgs` reads more naturally
// for the list-all case.
var Cmd = &cobra.Command{
	Use:     "org",
	Aliases: []string{"orgs"},
	Short:   "Switch between organizations and view membership",
	Long: `Manage which organization the CLI targets.

Spork is multi-tenant: an account can own or belong to several
organizations. Use these commands to pick an active org once
("spork org use ORG_ID") instead of repeating --org / SPORK_ORG_ID
on every invocation.

Precedence (highest wins):
  --org flag  >  SPORK_ORG_ID env  >  persisted active org  >  auto-resolve

The active org is stored in plaintext (it isn't a secret) at the
platform's config directory — typically ~/.config/spork/config.json
on Linux and macOS. Set SPORK_CONFIG_DIR to relocate.`,
}

func init() {
	Cmd.AddCommand(listCmd, currentCmd, useCmd)
}
