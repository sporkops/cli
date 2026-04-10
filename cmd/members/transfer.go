package members

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	spork "github.com/sporkops/spork-go"
	"golang.org/x/term"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:   "transfer <member-id>",
	Short: "Transfer organization ownership",
	Long:  "Transfer ownership of the organization to another member. You will become a regular member.",
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
				return fmt.Errorf("refusing to transfer ownership without --yes in non-interactive mode")
			}
			fmt.Printf("Transfer organization ownership to member %s? You will become a regular member. [y/N] ", memberID)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		result, err := client.TransferOwnership(context.Background(), &spork.TransferOwnershipInput{
			MemberID: memberID,
		})
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error transferring ownership: %s\n", err)
			return err
		}

		if isJSON {
			return output.PrintJSON(result)
		}

		fmt.Println("Ownership transferred.")
		fmt.Printf("  New Owner:       %s (%s)\n", result.NewOwner.Email, result.NewOwner.ID)
		fmt.Printf("  Previous Owner:  %s (%s)\n", result.PreviousOwner.Email, result.PreviousOwner.ID)
		return nil
	},
}

func init() {
	transferCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	transferCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
