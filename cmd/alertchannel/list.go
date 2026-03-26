package alertchannel

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all alert channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		channels, err := client.ListAlertChannels()
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing alert channels: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(channels)
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
