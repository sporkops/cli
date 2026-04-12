package members

import "github.com/spf13/cobra"

// Cmd is the parent command for member management. Named in the
// singular to match every other top-level command group (`monitor`,
// `incident`, `alert-channel`, `api-key`, `status-page`, `webhook`);
// `members` is retained as an alias so existing scripts keep working.
var Cmd = &cobra.Command{
	Use:     "member",
	Aliases: []string{"members", "mem"},
	Short:   "Manage organization members",
	Long:    "Invite, list, remove, and transfer ownership of organization members.",
}

func init() {
	Cmd.AddCommand(listCmd, inviteCmd, removeCmd, transferCmd)
}
