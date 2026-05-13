package org

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/config"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations you belong to",
	Long: `List every organization your account is a member of, with the
currently active one marked.

The active org is whichever takes precedence at call time — --org,
SPORK_ORG_ID, or the value saved by ` + "`spork org use`" + `. When you're
authenticated with an API key, the list has exactly one row (keys are
bound to a single org).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		spinner := output.NewSpinner("Loading organizations...")
		if spinner != nil {
			spinner.Start()
		}
		orgs, err := client.ListMyOrgs(context.Background())
		if spinner != nil {
			spinner.Stop()
		}
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing organizations: %s\n", err)
			return err
		}

		active := activeOrgID()

		if cmdutil.Structured(cmd) {
			// In structured output, encode `active` as a per-row boolean
			// rather than a separate top-level field so each row is
			// self-describing for jq filters: `.[] | select(.active)`.
			type row struct {
				ID            string `json:"id"`
				Name          string `json:"name,omitempty"`
				Role          string `json:"role"`
				Active        bool   `json:"active"`
				MonitoringPlan string `json:"monitoring_plan,omitempty"`
			}
			out := make([]row, 0, len(orgs))
			for _, o := range orgs {
				out = append(out, row{
					ID: o.ID, Name: o.Name, Role: o.Role,
					Active:         o.ID == active,
					MonitoringPlan: monitoringPlan(o.Subscriptions),
				})
			}
			return cmdutil.PrintStructured(cmd, out)
		}

		if len(orgs) == 0 {
			fmt.Println("No organizations found.")
			return nil
		}

		headers := []string{"", "ID", "NAME", "ROLE", "PLAN"}
		rows := make([][]string, 0, len(orgs))
		for _, o := range orgs {
			marker := ""
			if o.ID == active {
				marker = "*"
			}
			rows = append(rows, []string{
				marker,
				o.ID,
				orDash(o.Name),
				o.Role,
				orDash(monitoringPlan(o.Subscriptions)),
			})
		}
		output.PrintTable(headers, rows)
		if active != "" {
			fmt.Printf("\n* active organization (use \"spork org use ORG_ID\" to switch)\n")
		} else {
			fmt.Printf("\nNo active org set. Use \"spork org use ORG_ID\" to pin one,\n")
			fmt.Printf("or set --org / SPORK_ORG_ID on individual commands.\n")
		}
		return nil
	},
}

// activeOrgID resolves which org is currently active using the same
// precedence the rest of the CLI applies (flag > env > config). Unlike
// the runtime path, "current" intentionally observes the persisted
// preference too so list/current/use stay self-consistent.
func activeOrgID() string {
	// --org flag / SPORK_ORG_ID are already merged into cmdutil.OrgID by
	// the root command's PersistentPreRunE, so we just need to fall back
	// to the persisted value when nothing more specific was supplied.
	if cmdutil.OrgID != "" {
		return cmdutil.OrgID
	}
	s, err := config.Load()
	if err != nil {
		// A malformed config is reported elsewhere — for the marker we
		// degrade gracefully to "no active org" rather than failing the
		// list command on a corrupt preferences file.
		return ""
	}
	return s.ActiveOrganizationID
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// monitoringPlan extracts the plan for the "monitoring" product from a
// subscription bundle, returning "" if absent. The CLI is primarily a
// monitoring tool, so showing the monitoring plan is more useful than
// the dim sum of "you have N subscriptions"; users who want the full
// breakdown can `--output json`.
func monitoringPlan(subs []spork.Subscription) string {
	for _, s := range subs {
		if s.Product == "monitoring" {
			return s.Plan
		}
	}
	return ""
}
