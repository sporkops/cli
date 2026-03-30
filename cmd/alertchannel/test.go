package alertchannel

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <id>",
	Short: "Send a test notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id := args[0]

		if err := client.TestAlertChannel(context.Background(), id); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error sending test notification: %s\n", err)
			return err
		}

		fmt.Printf("✓ Test notification sent to %s\n", id)
		return nil
	},
}
