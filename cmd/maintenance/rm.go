package maintenance

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	rmForce bool
	rmYes   bool
)

var rmCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a maintenance window",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		skipPrompt := rmForce || rmYes
		if !skipPrompt {
			isJSON := cmdutil.Structured(cmd)
			if !term.IsTerminal(int(os.Stdout.Fd())) || isJSON {
				return fmt.Errorf("refusing to delete without --yes in non-interactive mode")
			}
			fmt.Printf("Delete maintenance window %q? [y/N] ", args[0])
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteMaintenanceWindow(context.Background(), args[0]); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error deleting maintenance window: %s\n", err)
			return err
		}

		fmt.Printf("✓ Maintenance window deleted: %s\n", args[0])
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "skip confirmation prompt")
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "skip confirmation prompt")
}
