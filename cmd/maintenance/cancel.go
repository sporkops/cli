package maintenance

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// cancelCmd transitions a scheduled or in-progress window to the
// "cancelled" state. The window stays visible for audit but stops
// suppressing alerts and pausing checks immediately.
//
// Prefer cancel over delete when you want a record of the aborted
// maintenance (which you usually do — the audit trail is what
// on-call rotations look at when troubleshooting "why did this alert
// not fire?").
var cancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a maintenance window",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		result, err := client.CancelMaintenanceWindow(context.Background(), args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error cancelling maintenance window: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}
		fmt.Printf("✓ Maintenance window cancelled: %s (%s)\n", result.Name, result.ID)
		return nil
	},
}
