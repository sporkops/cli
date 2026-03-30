package apikey

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an API key",
	Long:  "Delete an API key by ID.\n\nExample:\n  spork api-key delete abc123\n  spork api-key delete abc123 --force",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id := args[0]

		if !forceDelete {
			fmt.Printf("Delete API key %q? [y/N] ", id)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteAPIKey(id); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error deleting API key: %s\n", err)
			return err
		}

		fmt.Printf("✓ API key deleted: %s\n", id)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation prompt")
}
