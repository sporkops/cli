package incident

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	createStatusPage   string
	createTitle        string
	createMessage      string
	createType         string
	createStatus       string
	createImpact       string
	createComponentIDs []string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new incident",
	Long: `Create a new incident on a status page.

Examples:
  spork incident create --status-page sp_abc --title "API latency detected"
  spork incident create --status-page sp_abc --title "Scheduled maintenance" \
    --type maintenance --status monitoring --impact minor
  spork incident create --status-page sp_abc --title "Database outage" \
    --message "Investigating elevated error rates" --impact critical`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		title := strings.TrimSpace(createTitle)
		if len(title) == 0 || len(title) > 200 {
			return fmt.Errorf("--title must be 1-200 characters")
		}
		if len(createMessage) > 10000 {
			return fmt.Errorf("--message must be at most 10000 characters")
		}
		if !validTypes[createType] {
			return fmt.Errorf("invalid --type %q: must be incident or maintenance", createType)
		}
		if !validStatuses[createStatus] {
			return fmt.Errorf("invalid --status %q: must be investigating, identified, monitoring, or resolved", createStatus)
		}
		if !validImpacts[createImpact] {
			return fmt.Errorf("invalid --impact %q: must be none, minor, major, or critical", createImpact)
		}

		inc := &spork.Incident{
			Title:        title,
			Message:      createMessage,
			Type:         createType,
			Status:       createStatus,
			Impact:       createImpact,
			ComponentIDs: createComponentIDs,
		}

		result, err := client.CreateIncident(context.Background(), createStatusPage, inc)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating incident: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
		}

		fmt.Printf("✓ Incident created: %s\n", result.Title)
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Status: %s\n", result.Status)
		if result.Impact != "" && result.Impact != "none" {
			fmt.Printf("  Impact: %s\n", result.Impact)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createStatusPage, "status-page", "", "status page ID (required)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "incident title (required, 1-200 chars)")
	createCmd.Flags().StringVar(&createMessage, "message", "", "incident message (max 10000 chars)")
	createCmd.Flags().StringVar(&createType, "type", "incident", "incident type: incident, maintenance")
	createCmd.Flags().StringVar(&createStatus, "status", "investigating", "status: investigating, identified, monitoring, resolved")
	createCmd.Flags().StringVar(&createImpact, "impact", "none", "impact level: none, minor, major, critical")
	createCmd.Flags().StringSliceVar(&createComponentIDs, "component-ids", nil, "affected component IDs (comma-separated)")
	createCmd.MarkFlagRequired("status-page")
	createCmd.MarkFlagRequired("title")
}
