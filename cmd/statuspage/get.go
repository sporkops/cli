package statuspage

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id|name|slug>",
	Short: "Show details of a status page",
	Long: `Show full details of a single status page by ID, name, or slug.

Example:
  spork status-page get sp_abc123
  spork status-page get acme-status
  spork status-page get "Acme Status" --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id, _, err := resolveStatusPageID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		sp, err := client.GetStatusPage(id)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching status page: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(sp)
		}

		fmt.Printf("%-20s %s\n", "ID:", sp.ID)
		fmt.Printf("%-20s %s\n", "Name:", sp.Name)
		fmt.Printf("%-20s %s\n", "Slug:", sp.Slug)
		fmt.Printf("%-20s https://%s.status.sporkops.com\n", "URL:", sp.Slug)
		fmt.Printf("%-20s %s\n", "Theme:", sp.Theme)
		fmt.Printf("%-20s %s\n", "Public:", strconv.FormatBool(sp.IsPublic))

		if sp.AccentColor != "" {
			fmt.Printf("%-20s %s\n", "Accent Color:", sp.AccentColor)
		}
		if sp.FontFamily != "" {
			fmt.Printf("%-20s %s\n", "Font Family:", sp.FontFamily)
		}
		if sp.HeaderStyle != "" {
			fmt.Printf("%-20s %s\n", "Header Style:", sp.HeaderStyle)
		}
		if sp.LogoURL != "" {
			fmt.Printf("%-20s %s\n", "Logo URL:", sp.LogoURL)
		}
		if sp.WebhookURL != "" {
			fmt.Printf("%-20s %s\n", "Webhook URL:", sp.WebhookURL)
		}
		if sp.CustomDomain != "" {
			fmt.Printf("%-20s %s (%s)\n", "Custom Domain:", sp.CustomDomain, sp.DomainStatus)
		}
		fmt.Printf("%-20s %s\n", "Email Subscribers:", strconv.FormatBool(sp.EmailSubscribersEnabled))

		if len(sp.ComponentGroups) > 0 {
			fmt.Println()
			fmt.Println("Component Groups:")
			groupHeaders := []string{"ID", "NAME", "ORDER"}
			groupRows := make([][]string, len(sp.ComponentGroups))
			for i, g := range sp.ComponentGroups {
				groupRows[i] = []string{g.ID, g.Name, strconv.Itoa(g.Order)}
			}
			output.PrintTable(groupHeaders, groupRows)
		}

		if len(sp.Components) > 0 {
			fmt.Println()
			fmt.Println("Components:")
			headers := []string{"ORDER", "DISPLAY NAME", "MONITOR ID", "GROUP ID", "DESCRIPTION"}
			rows := make([][]string, len(sp.Components))
			for i, c := range sp.Components {
				desc := "-"
				if c.Description != "" {
					desc = c.Description
				}
				groupID := "-"
				if c.GroupID != "" {
					groupID = c.GroupID
				}
				rows[i] = []string{
					strconv.Itoa(c.Order),
					c.DisplayName,
					c.MonitorID,
					groupID,
					desc,
				}
			}
			output.PrintTable(headers, rows)
		}

		fmt.Println()
		fmt.Printf("%-20s %s\n", "Created:", sp.CreatedAt)
		fmt.Printf("%-20s %s\n", "Updated:", sp.UpdatedAt)

		return nil
	},
}
