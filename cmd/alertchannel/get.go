package alertchannel

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show details of an alert channel",
	Long: `Show full details of a single alert channel by ID.

Example:
  spork alert-channel get ac_abc123
  spork alert-channel get ac_abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		ch, err := client.GetAlertChannel(context.Background(), args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching alert channel: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(ch)
		}

		fmt.Printf("%-20s %s\n", "ID:", ch.ID)
		fmt.Printf("%-20s %s\n", "Name:", ch.Name)
		fmt.Printf("%-20s %s\n", "Type:", ch.Type)
		fmt.Printf("%-20s %s\n", "Verified:", strconv.FormatBool(ch.Verified))

		if len(ch.Config) > 0 {
			fmt.Printf("%-20s\n", "Config:")
			for k, v := range ch.Config {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		if ch.LastDeliveryStatus != "" {
			fmt.Printf("%-20s %s\n", "Last Delivery:", ch.LastDeliveryStatus)
		}
		if ch.LastDeliveryAt != "" {
			fmt.Printf("%-20s %s\n", "Last Delivery At:", ch.LastDeliveryAt)
		}
		fmt.Printf("%-20s %s\n", "Created:", ch.CreatedAt)
		fmt.Printf("%-20s %s\n", "Updated:", ch.UpdatedAt)

		return nil
	},
}
