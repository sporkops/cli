package incident

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
)

var rmForce bool

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Delete an incident",
	Long: `Delete an incident by ID.

Example:
  spork incident rm inc_abc123
  spork incident rm inc_abc123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		if !rmForce {
			fmt.Printf("Delete incident %s? [y/N] ", args[0])
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteIncident(args[0]); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error deleting incident: %s\n", err)
			return err
		}

		fmt.Printf("✓ Incident %s deleted\n", args[0])
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "skip confirmation prompt")
}
