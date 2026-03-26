package ping

import (
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/api"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause <id|url>",
	Short: "Pause a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id, _, err := resolveMonitorID(client, args[0])
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		paused := true
		result, err := client.UpdateMonitor(id, &api.Monitor{Paused: &paused})
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error pausing monitor: %s\n", err)
			return err
		}

		fmt.Printf("✓ Monitor paused: %s (%s)\n", result.Name, result.Target)
		return nil
	},
}

var unpauseCmd = &cobra.Command{
	Use:   "unpause <id|url>",
	Short: "Resume a paused monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id, _, err := resolveMonitorID(client, args[0])
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		paused := false
		result, err := client.UpdateMonitor(id, &api.Monitor{Paused: &paused})
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error unpausing monitor: %s\n", err)
			return err
		}

		fmt.Printf("✓ Monitor unpaused: %s (%s)\n", result.Name, result.Target)
		return nil
	},
}
