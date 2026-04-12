// Package webhook hosts the `spork webhook ...` command group. Today that
// is a single subcommand — `trigger`, which fires a synthetic signed event
// at a webhook alert channel for integration testing. The group exists so
// a future `listen` subcommand (tunneling real events to a local endpoint)
// can slot in without a rename.
package webhook

import "github.com/spf13/cobra"

// Cmd is the `spork webhook` parent command.
var Cmd = &cobra.Command{
	Use:   "webhook",
	Short: "Test and verify outbound webhook deliveries",
	Long: `Tools for working with Spork's outbound webhooks.

  - ` + "`trigger`" + ` fires a synthetic, signed event at a webhook alert
    channel so you can exercise your receiver end-to-end without waiting
    for a real monitor to flap.`,
}

func init() {
	Cmd.AddCommand(triggerCmd)
}
