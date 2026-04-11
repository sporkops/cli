package members

import "github.com/spf13/cobra"

// Cmd is the parent command for member management.
var Cmd = &cobra.Command{
	Use:   "members",
	Short: "Manage organization members",
	Long:  "Invite, list, remove, and transfer ownership of organization members.",
}

func init() {
	Cmd.AddCommand(listCmd, inviteCmd, removeCmd, transferCmd)
}
