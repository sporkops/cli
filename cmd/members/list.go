package members

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List organization members",
	Long:  "List all members of your organization, including pending invitations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}
		isJSON := cmdutil.Structured(cmd)

		spinner := output.NewSpinner("Loading members...")
		if spinner != nil {
			spinner.Start()
		}
		members, err := client.ListMembers(context.Background())
		if spinner != nil {
			spinner.Stop()
		}
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing members: %s\n", err)
			return err
		}

		if isJSON {
			return cmdutil.PrintStructured(cmd, members)
		}

		if len(members) == 0 {
			fmt.Println("No members found.")
			return nil
		}

		headers := []string{"ID", "EMAIL", "ROLE", "STATUS", "JOINED"}
		var rows [][]string
		for _, m := range members {
			rows = append(rows, []string{
				m.ID,
				m.Email,
				m.Role,
				output.ColorStatus(m.Status),
				m.CreatedAt.Format("2006-01-02"),
			})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
