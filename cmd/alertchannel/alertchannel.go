package alertchannel

import (
	"github.com/spf13/cobra"
)

// Cmd is the `spork alert-channel` parent command.
var Cmd = &cobra.Command{
	Use:     "alert-channel",
	Aliases: []string{"ac"},
	Short:   "Manage alert channels",
	Long:    "Create, list, and manage alert channels for downtime notifications.",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(testCmd)
}
