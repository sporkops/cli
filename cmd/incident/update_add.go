package incident

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	updateAddStatus  string
	updateAddMessage string
)

var updateAddCmd = &cobra.Command{
	Use:   "update-add <incident-id>",
	Short: "Add a timeline update to an incident",
	Long: `Add a timeline update to an existing incident.

At least one of --status or --message is required. If --status differs from the
current incident status, the incident status is automatically updated.

Examples:
  spork incident update-add inc_abc --status identified --message "Root cause found"
  spork incident update-add inc_abc --message "Continuing to monitor"
  spork incident update-add inc_abc --status resolved`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		hasStatus := cmd.Flags().Changed("status")
		hasMessage := cmd.Flags().Changed("message")

		if !hasStatus && !hasMessage {
			return fmt.Errorf("at least one of --status or --message is required")
		}

		upd := &spork.IncidentUpdate{}
		if hasStatus {
			if !validStatuses[updateAddStatus] {
				return fmt.Errorf("invalid --status %q: must be investigating, identified, monitoring, or resolved", updateAddStatus)
			}
			upd.Status = updateAddStatus
		}
		if hasMessage {
			if len(updateAddMessage) > 10000 {
				return fmt.Errorf("--message must be at most 10000 characters")
			}
			upd.Message = updateAddMessage
		}

		result, err := client.CreateIncidentUpdate(context.Background(), args[0], upd)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error adding incident update: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
		}

		fmt.Printf("✓ Update added to incident %s\n", args[0])
		if result.Status != "" {
			fmt.Printf("  Status: %s\n", result.Status)
		}
		return nil
	},
}

func init() {
	updateAddCmd.Flags().StringVar(&updateAddStatus, "status", "", "status: investigating, identified, monitoring, resolved")
	updateAddCmd.Flags().StringVar(&updateAddMessage, "message", "", "update message (max 10000 chars)")
}
