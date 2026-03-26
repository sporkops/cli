package alertchannel

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var forceRemove bool

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove an alert channel",
	Long:  "Remove an alert channel by ID.\n\nExample:\n  spork alert-channel rm abc123\n  spork alert-channel rm abc123 --force",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id := args[0]

		if !forceRemove {
			fmt.Printf("Remove alert channel %q? [y/N] ", id)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteAlertChannel(id); err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error removing alert channel: %s\n", err)
			return err
		}

		fmt.Printf("✓ Alert channel removed: %s\n", id)
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "skip confirmation prompt")
}
