package apikey

import (
	"github.com/spf13/cobra"
)

// Cmd is the `spork api-key` parent command.
var Cmd = &cobra.Command{
	Use:     "api-key",
	Aliases: []string{"ak"},
	Short:   "Manage API keys",
	Long:    "Create, list, and delete API keys for programmatic access.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deleteCmd)
}
