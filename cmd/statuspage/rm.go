package statuspage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var forceRemove bool

var rmCmd = &cobra.Command{
	Use:   "rm <id|name|slug>",
	Short: "Remove a status page",
	Long: `Remove a status page by ID, name, or slug.

Example:
  spork status-page rm sp_abc123
  spork status-page rm acme-status --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id, name, err := resolveStatusPageID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		label := args[0]
		if name != "" {
			label = name
		}

		if !forceRemove {
			fmt.Printf("Remove status page %q? [y/N] ", label)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteStatusPage(context.Background(), id); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error removing status page: %s\n", err)
			return err
		}

		fmt.Printf("✓ Status page removed: %s\n", label)
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "skip confirmation prompt")
}
