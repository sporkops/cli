package members

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	spork "github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite <email>",
	Short: "Invite a member to the organization",
	Long:  "Send an invitation email to add a new member to your organization.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := strings.TrimSpace(args[0])
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email address: %s", email)
		}

		role, _ := cmd.Flags().GetString("role")
		if role != "member" {
			return fmt.Errorf("invalid role: %s (only 'member' is allowed; use 'transfer' to change ownership)", role)
		}

		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}
		isJSON := cmd.Root().Flag("json").Changed

		member, err := client.InviteMember(context.Background(), &spork.InviteMemberInput{
			Email: email,
			Role:  role,
		})
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error inviting member: %s\n", err)
			return err
		}

		if isJSON {
			return output.PrintJSON(member)
		}

		fmt.Printf("Invited %s as %s (ID: %s, Status: %s)\n", member.Email, member.Role, member.ID, member.Status)
		return nil
	},
}

func init() {
	inviteCmd.Flags().String("role", "member", "Role for the new member")
}
