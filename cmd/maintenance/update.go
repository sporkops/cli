package maintenance

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	updateName              string
	updateDescription       string
	updateStart             string
	updateEnd               string
	updateTimezone          string
	updateRecurrence        string
	updateRecurrenceDays    []int
	updateRecurrenceUntil   string
	updateSuppressAlerts    bool
	updateNoSuppressAlerts  bool
	updateExcludeUptime     bool
	updateNoExcludeUptime   bool
	updatePauseChecks       bool
	updateNoPauseChecks     bool
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a maintenance window (partial)",
	Long: `Update a maintenance window in-place. Only flags you pass are
changed — everything else is left alone.

To retarget monitors use a fresh create; the API does not replace
targeting arrays via PATCH to avoid accidental broadening.

All boolean flags default to false so "unset" and "explicitly false" are
distinguishable. To re-enable a previously disabled toggle, pass --X=true
(e.g., --suppress-alerts=true).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		patch := &spork.MaintenanceWindow{}
		applyUpdateFlags(cmd, patch)

		result, err := client.UpdateMaintenanceWindow(context.Background(), args[0], patch)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error updating maintenance window: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}
		fmt.Printf("✓ Maintenance window updated: %s (%s)\n", result.Name, result.ID)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "new display name")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "new description")
	updateCmd.Flags().StringVar(&updateStart, "start", "", "new RFC3339 UTC start")
	updateCmd.Flags().StringVar(&updateEnd, "end", "", "new RFC3339 UTC end")
	updateCmd.Flags().StringVar(&updateTimezone, "timezone", "", "new IANA timezone")
	updateCmd.Flags().StringVar(&updateRecurrence, "recurrence", "", "new recurrence type")
	updateCmd.Flags().IntSliceVar(&updateRecurrenceDays, "recurrence-days", nil, "new recurrence days")
	updateCmd.Flags().StringVar(&updateRecurrenceUntil, "recurrence-until", "", "new recurrence cap")
	// All bool flags default to false so Changed() cleanly detects explicit
	// user intent. Leaving a flag unset leaves the corresponding server
	// field untouched.
	updateCmd.Flags().BoolVar(&updateSuppressAlerts, "suppress-alerts", false, "re-enable alert suppression (set the field to true)")
	updateCmd.Flags().BoolVar(&updateNoSuppressAlerts, "no-suppress-alerts", false, "disable alert suppression (set the field to false)")
	updateCmd.Flags().BoolVar(&updateExcludeUptime, "exclude-from-uptime", false, "re-enable uptime exclusion (set the field to true)")
	updateCmd.Flags().BoolVar(&updateNoExcludeUptime, "no-exclude-from-uptime", false, "disable uptime exclusion (set the field to false)")
	updateCmd.Flags().BoolVar(&updatePauseChecks, "pause-checks", false, "enable check pausing (set the field to true)")
	updateCmd.Flags().BoolVar(&updateNoPauseChecks, "no-pause-checks", false, "disable check pausing (set the field to false)")
}

func applyUpdateFlags(cmd *cobra.Command, patch *spork.MaintenanceWindow) {
	if updateName != "" {
		patch.Name = updateName
	}
	if updateDescription != "" {
		patch.Description = updateDescription
	}
	if updateStart != "" {
		patch.StartAt = updateStart
	}
	if updateEnd != "" {
		patch.EndAt = updateEnd
	}
	if updateTimezone != "" {
		patch.Timezone = updateTimezone
	}
	if updateRecurrence != "" {
		patch.RecurrenceType = updateRecurrence
	}
	if len(updateRecurrenceDays) > 0 {
		patch.RecurrenceDays = updateRecurrenceDays
	}
	if updateRecurrenceUntil != "" {
		patch.RecurrenceUntil = updateRecurrenceUntil
	}
	// Pointer-bool flags. --no-X wins over explicit --X; without either,
	// leave nil so the server keeps its current value.
	switch {
	case updateNoSuppressAlerts:
		v := false
		patch.SuppressAlerts = &v
	case cmd.Flags().Changed("suppress-alerts"):
		v := updateSuppressAlerts
		patch.SuppressAlerts = &v
	}
	switch {
	case updateNoExcludeUptime:
		v := false
		patch.ExcludeFromUptime = &v
	case cmd.Flags().Changed("exclude-from-uptime"):
		v := updateExcludeUptime
		patch.ExcludeFromUptime = &v
	}
	switch {
	case updateNoPauseChecks:
		v := false
		patch.PauseChecks = &v
	case cmd.Flags().Changed("pause-checks"):
		v := updatePauseChecks
		patch.PauseChecks = &v
	}
}
