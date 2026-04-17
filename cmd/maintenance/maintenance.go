// Package maintenance implements `spork maintenance` — CRUD for scheduled
// maintenance windows that suppress alerts (and optionally pause checks)
// for a set of monitors.
//
// The command layout and flag conventions mirror the monitor package:
// create / list / get / update / delete / cancel, each file holding one
// command. Resources are identified by ID; update and cancel take a
// single positional arg, list is paginated, create accepts either flat
// flags or --file for JSON/YAML.
package maintenance

import "github.com/spf13/cobra"

// Cmd is the `spork maintenance` parent command, registered from cmd/root.go.
var Cmd = &cobra.Command{
	Use:     "maintenance",
	Aliases: []string{"mw", "maintenance-window"},
	Short:   "Manage scheduled maintenance windows",
	Long: `Schedule maintenance windows that suppress alerts and optionally
pause checks for a set of monitors.

A window can target specific monitor IDs, match monitors by tag, or apply
to every monitor in the organization. Windows can be one-time or recurring
(daily / weekly / monthly). All times are stored in UTC; supply an IANA
timezone so recurrence is expanded correctly across DST boundaries.

See https://sporkops.com/docs/maintenance-windows for details.`,
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(cancelCmd)
}
