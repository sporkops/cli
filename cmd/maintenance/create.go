package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	createFile              string
	createName              string
	createDescription       string
	createMonitorIDs        []string
	createTags              []string
	createAllMonitors       bool
	createTimezone          string
	createStart             string
	createEnd               string
	createRecurrence        string
	createRecurrenceDays    []int
	createRecurrenceUntil   string
	createSuppressAlerts    bool
	createNoSuppressAlerts  bool
	createExcludeUptime     bool
	createNoExcludeUptime   bool
	createPauseChecks       bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new maintenance window",
	Long: `Create a new maintenance window.

The window takes effect between --start and --end (RFC3339). Supply --timezone
as an IANA name (e.g., "America/Los_Angeles") so recurrence is expanded with
the correct DST offset.

Target exactly one of:
  --all-monitors
  --monitors id1,id2,...
  --tags tagA,tagB,...

For a recurring window, pass --recurrence daily|weekly|monthly plus
--recurrence-days (Sun=0 for weekly, 1-31 for monthly). --recurrence-until
caps the series.

You may alternatively supply --file pointing to a JSON or YAML file with the
full body. Flags override file values.`,
	Example: `  spork maintenance create --name "Weekly DB" --all-monitors \
    --timezone America/Los_Angeles \
    --start 2026-05-05T09:00:00Z --end 2026-05-05T10:00:00Z \
    --recurrence weekly --recurrence-days 2

  spork maintenance create --file window.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		mw := &spork.MaintenanceWindow{}
		if createFile != "" {
			if err := loadWindowFile(createFile, mw); err != nil {
				return err
			}
		}
		applyCreateFlags(cmd, mw)

		if err := validateTargeting(mw); err != nil {
			return err
		}

		result, err := client.CreateMaintenanceWindow(context.Background(), mw)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating maintenance window: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}
		fmt.Printf("✓ Maintenance window created: %s (%s)\n", result.Name, result.ID)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createFile, "file", "", "path to a JSON or YAML file with the maintenance window body")
	createCmd.Flags().StringVar(&createName, "name", "", "display name")
	createCmd.Flags().StringVar(&createDescription, "description", "", "optional description")
	createCmd.Flags().StringSliceVar(&createMonitorIDs, "monitors", nil, "comma-separated monitor IDs to target")
	createCmd.Flags().StringSliceVar(&createTags, "tags", nil, "comma-separated monitor tags to target (OR semantics)")
	createCmd.Flags().BoolVar(&createAllMonitors, "all-monitors", false, "target every monitor in the organization")
	createCmd.Flags().StringVar(&createTimezone, "timezone", "", "IANA timezone name (required)")
	createCmd.Flags().StringVar(&createStart, "start", "", "RFC3339 start time in UTC (required)")
	createCmd.Flags().StringVar(&createEnd, "end", "", "RFC3339 end time in UTC (required)")
	createCmd.Flags().StringVar(&createRecurrence, "recurrence", "", "recurrence type: daily, weekly, or monthly")
	createCmd.Flags().IntSliceVar(&createRecurrenceDays, "recurrence-days", nil, "weekly: 0-6 (Sun=0); monthly: 1-31")
	createCmd.Flags().StringVar(&createRecurrenceUntil, "recurrence-until", "", "RFC3339 UTC cap on the series")
	createCmd.Flags().BoolVar(&createSuppressAlerts, "suppress-alerts", true, "suppress alert deliveries during the window")
	createCmd.Flags().BoolVar(&createNoSuppressAlerts, "no-suppress-alerts", false, "deliver alerts even during the window")
	createCmd.Flags().BoolVar(&createExcludeUptime, "exclude-from-uptime", true, "drop in-window checks from uptime-percentage")
	createCmd.Flags().BoolVar(&createNoExcludeUptime, "no-exclude-from-uptime", false, "include in-window checks in uptime-percentage")
	createCmd.Flags().BoolVar(&createPauseChecks, "pause-checks", false, "skip dispatch entirely during the window (default keeps checks running)")
}

func applyCreateFlags(cmd *cobra.Command, mw *spork.MaintenanceWindow) {
	if createName != "" {
		mw.Name = createName
	}
	if createDescription != "" {
		mw.Description = createDescription
	}
	if len(createMonitorIDs) > 0 {
		mw.MonitorIDs = createMonitorIDs
	}
	if len(createTags) > 0 {
		mw.TagSelectors = createTags
	}
	if cmd.Flags().Changed("all-monitors") {
		v := createAllMonitors
		mw.AllMonitors = &v
	}
	if createTimezone != "" {
		mw.Timezone = createTimezone
	}
	if createStart != "" {
		mw.StartAt = createStart
	}
	if createEnd != "" {
		mw.EndAt = createEnd
	}
	if createRecurrence != "" {
		mw.RecurrenceType = createRecurrence
	}
	if len(createRecurrenceDays) > 0 {
		mw.RecurrenceDays = createRecurrenceDays
	}
	if createRecurrenceUntil != "" {
		mw.RecurrenceUntil = createRecurrenceUntil
	}
	// Pointer-bool flags: honor --no-X overrides, then explicit --X,
	// then leave unset so the server default applies.
	if createNoSuppressAlerts {
		v := false
		mw.SuppressAlerts = &v
	} else if cmd.Flags().Changed("suppress-alerts") {
		v := createSuppressAlerts
		mw.SuppressAlerts = &v
	}
	if createNoExcludeUptime {
		v := false
		mw.ExcludeFromUptime = &v
	} else if cmd.Flags().Changed("exclude-from-uptime") {
		v := createExcludeUptime
		mw.ExcludeFromUptime = &v
	}
	if cmd.Flags().Changed("pause-checks") {
		v := createPauseChecks
		mw.PauseChecks = &v
	}
}

func validateTargeting(mw *spork.MaintenanceWindow) error {
	targets := 0
	if len(mw.MonitorIDs) > 0 {
		targets++
	}
	if len(mw.TagSelectors) > 0 {
		targets++
	}
	if mw.AllMonitors != nil && *mw.AllMonitors {
		targets++
	}
	if targets != 1 {
		return fmt.Errorf("exactly one of --monitors, --tags, or --all-monitors must be set")
	}
	return nil
}

func loadWindowFile(path string, mw *spork.MaintenanceWindow) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
		if err := yaml.Unmarshal(data, mw); err != nil {
			return fmt.Errorf("parsing YAML %s: %w", path, err)
		}
		return nil
	}
	if err := json.Unmarshal(data, mw); err != nil {
		return fmt.Errorf("parsing JSON %s: %w", path, err)
	}
	return nil
}
