package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	spork "github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current organization info",
	Long:  "Display your organization details: name, email, role, subscriptions, and entitlements.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use the shared RequireAuth so --org / SPORK_ORG_ID,
		// --debug HTTP tracing, and the spork-cli user-agent all
		// flow through. Hand-rolling spork.NewClient here used to
		// silently ignore those flags — a multi-org user passing
		// --org would see the wrong org, and --debug was inert.
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		org, err := client.GetOrganization(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, org)
		}

		name := org.Name
		if name == "" {
			name = org.ID
		}
		fmt.Printf("%-17s%s (%s)\n", "Organization:", org.ID, name)
		if org.User != nil {
			fmt.Printf("%-17s%s\n", "Email:", org.User.Email)
			fmt.Printf("%-17s%s\n", "Role:", org.User.Role)
		}
		fmt.Printf("%-17s%s\n", "Member Since:", org.CreatedAt.Format("2006-01-02"))

		if len(org.Subscriptions) > 0 {
			fmt.Println()
			fmt.Println("Subscriptions:")
			for _, sub := range org.Subscriptions {
				summary := formatEntitlementsSummary(sub)
				fmt.Printf("  %-14s%-11s%s\n", sub.Product, sub.Plan, summary)
			}
		}
		return nil
	},
}

// formatEntitlementsSummary formats key entitlements as a human-readable string.
func formatEntitlementsSummary(sub spork.Subscription) string {
	var parts []string
	if v := sub.EntitlementInt("monitor_limit"); v > 0 {
		parts = append(parts, fmt.Sprintf("%d monitors", v))
	}
	if v := sub.EntitlementInt("check_interval_s"); v > 0 {
		parts = append(parts, fmt.Sprintf("%ds checks", v))
	}
	if v := sub.EntitlementInt("member_limit"); v > 0 {
		parts = append(parts, fmt.Sprintf("%d members", v))
	}
	return strings.Join(parts, ", ")
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
