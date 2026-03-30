package incident

import (
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	updateTitle        string
	updateMessage      string
	updateStatus       string
	updateImpact       string
	updateComponentIDs []string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an incident",
	Long: `Update an existing incident's fields.

Example:
  spork incident update inc_abc123 --status identified --message "Root cause found"
  spork incident update inc_abc123 --status resolved
  spork incident update inc_abc123 --impact critical --title "Major outage"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inc := &api.Incident{}
		hasChanges := false

		if cmd.Flags().Changed("title") {
			title := strings.TrimSpace(updateTitle)
			if len(title) == 0 || len(title) > 200 {
				return fmt.Errorf("--title must be 1-200 characters")
			}
			inc.Title = title
			hasChanges = true
		}
		if cmd.Flags().Changed("message") {
			if len(updateMessage) > 10000 {
				return fmt.Errorf("--message must be at most 10000 characters")
			}
			inc.Message = updateMessage
			hasChanges = true
		}
		if cmd.Flags().Changed("status") {
			if !validStatuses[updateStatus] {
				return fmt.Errorf("invalid --status %q: must be investigating, identified, monitoring, or resolved", updateStatus)
			}
			inc.Status = updateStatus
			hasChanges = true
		}
		if cmd.Flags().Changed("impact") {
			if !validImpacts[updateImpact] {
				return fmt.Errorf("invalid --impact %q: must be none, minor, major, or critical", updateImpact)
			}
			inc.Impact = updateImpact
			hasChanges = true
		}
		if cmd.Flags().Changed("component-ids") {
			inc.ComponentIDs = updateComponentIDs
			hasChanges = true
		}

		if !hasChanges {
			fmt.Fprintln(os.Stderr, "Nothing to update. Specify at least one flag:")
			fmt.Fprintln(os.Stderr, "  --title, --message, --status, --impact, --component-ids")
			return fmt.Errorf("no changes specified")
		}

		result, err := client.UpdateIncident(args[0], inc)
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error updating incident: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
		}

		fmt.Printf("✓ Incident updated: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "new incident title (1-200 chars)")
	updateCmd.Flags().StringVar(&updateMessage, "message", "", "new incident message (max 10000 chars)")
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "new status: investigating, identified, monitoring, resolved")
	updateCmd.Flags().StringVar(&updateImpact, "impact", "", "new impact: none, minor, major, critical")
	updateCmd.Flags().StringSliceVar(&updateComponentIDs, "component-ids", nil, "affected component IDs (comma-separated)")
}
