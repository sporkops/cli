package alertchannel

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <id>",
	Short: "Send a test notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id := args[0]

		if err := client.TestAlertChannel(id); err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error sending test notification: %s\n", err)
			return err
		}

		fmt.Printf("✓ Test notification sent to %s\n", id)
		return nil
	},
}
