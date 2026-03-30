package incident

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
	Use:   "rm <id>",
	Short: "Delete an incident",
	Long:  "Delete an incident by ID.",
	Example: `  spork incident rm inc_abc123
  spork incident rm inc_abc123 --yes
  spork incident rm inc_abc123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		skipPrompt := rmForce || rmYes
		if !skipPrompt {
			isJSON := cmd.Root().Flag("json").Changed
			if !term.IsTerminal(int(os.Stdout.Fd())) || isJSON {
				return fmt.Errorf("refusing to delete without --yes in non-interactive mode")
			}
			fmt.Printf("Delete incident %q? [y/N] ", args[0])
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteIncident(context.Background(), args[0]); err != nil {
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
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "skip confirmation prompt")
}
