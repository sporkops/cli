package members

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"golang.org/x/term"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <member-id>",
	Short: "Remove a member from the organization",
	Long:  "Remove a member from your organization. This revokes their access immediately.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		memberID := args[0]

		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}
		isJSON := cmd.Root().Flag("json").Changed

		force, _ := cmd.Flags().GetBool("force")
		yes, _ := cmd.Flags().GetBool("yes")

		if !force && !yes {
			if !term.IsTerminal(int(os.Stdout.Fd())) || isJSON {
				return fmt.Errorf("refusing to remove member without --yes in non-interactive mode")
			}
			fmt.Printf("Remove member %s? This will revoke their access. [y/N] ", memberID)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.RemoveMember(context.Background(), memberID); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error removing member: %s\n", err)
			return err
		}

		fmt.Printf("Member %s removed.\n", memberID)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	removeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
