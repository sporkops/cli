package alertchannel

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all alert channels",
	Example: `  spork alert-channel list
  spork alert-channel list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		channels, err := client.ListAlertChannels(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing alert channels: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, channels)
		}

		if len(channels) == 0 {
			fmt.Println("No alert channels yet. Add one:")
			fmt.Println("  spork alert-channel add --type email --name \"Alerts\" --config to=you@example.com")
			return nil
		}

		headers := []string{"ID", "NAME", "TYPE", "VERIFIED"}
		rows := make([][]string, len(channels))
		for i, ch := range channels {
			rows[i] = []string{
				ch.ID,
				ch.Name,
				ch.Type,
				strconv.FormatBool(ch.Verified),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
