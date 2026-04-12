package incident

import (
	"github.com/spf13/cobra"
)

// Cmd is the `spork incident` parent command.
var Cmd = &cobra.Command{
	Use:     "incident",
	Aliases: []string{"inc"},
	Short:   "Manage status page incidents",
	Long:    "Create, list, and manage incidents on your status pages.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(updateAddCmd)
	Cmd.AddCommand(updatesCmd)
	Cmd.AddCommand(recentCmd)
}

var validStatuses = map[string]bool{
	"investigating": true,
	"identified":    true,
	"monitoring":    true,
	"resolved":      true,
}

var validTypes = map[string]bool{
	"incident":    true,
	"maintenance": true,
}

var validImpacts = map[string]bool{
	"none":     true,
	"minor":    true,
	"major":    true,
	"critical": true,
}
