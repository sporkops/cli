package maintenance

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a maintenance window",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		result, err := client.GetMaintenanceWindow(context.Background(), args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching maintenance window: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}

		fmt.Printf("ID:           %s\n", result.ID)
		fmt.Printf("Name:         %s\n", result.Name)
		if result.Description != "" {
			fmt.Printf("Description:  %s\n", result.Description)
		}
		fmt.Printf("State:        %s\n", result.State)
		fmt.Printf("Timezone:     %s\n", result.Timezone)
		fmt.Printf("Start:        %s\n", result.StartAt)
		fmt.Printf("End:          %s\n", result.EndAt)
		if result.RecurrenceType != "" {
			fmt.Printf("Recurrence:   %s", result.RecurrenceType)
			if len(result.RecurrenceDays) > 0 {
				days := make([]string, len(result.RecurrenceDays))
				for i, d := range result.RecurrenceDays {
					days[i] = fmt.Sprintf("%d", d)
				}
				fmt.Printf(" (days: %s)", strings.Join(days, ","))
			}
			if result.RecurrenceUntil != "" {
				fmt.Printf(" until %s", result.RecurrenceUntil)
			}
			fmt.Println()
		} else {
			fmt.Println("Recurrence:   once")
		}
		switch {
		case result.AllMonitors != nil && *result.AllMonitors:
			fmt.Println("Targets:      all monitors")
		case len(result.TagSelectors) > 0:
			fmt.Printf("Targets:      tags %s\n", strings.Join(result.TagSelectors, ","))
		case len(result.MonitorIDs) > 0:
			fmt.Printf("Targets:      monitors %s\n", strings.Join(result.MonitorIDs, ","))
		}
		if result.SuppressAlerts != nil {
			fmt.Printf("Suppress:     %v\n", *result.SuppressAlerts)
		}
		if result.ExcludeFromUptime != nil {
			fmt.Printf("ExcludeUptime: %v\n", *result.ExcludeFromUptime)
		}
		if result.PauseChecks != nil {
			fmt.Printf("PauseChecks:  %v\n", *result.PauseChecks)
		}
		if result.CancelledAt != "" {
			fmt.Printf("Cancelled:    %s\n", result.CancelledAt)
		}
		fmt.Printf("Created:      %s\n", result.CreatedAt)
		fmt.Printf("Updated:      %s\n", result.UpdatedAt)
		return nil
	},
}
