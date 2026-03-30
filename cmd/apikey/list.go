package apikey

import (
	"context"
	"fmt"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		keys, err := client.ListAPIKeys(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Printf("Error listing API keys: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(keys)
		}

		if len(keys) == 0 {
			fmt.Println("No API keys yet. Create one:")
			fmt.Println("  spork api-key create --name <name>")
			return nil
		}

		headers := []string{"ID", "NAME", "PREFIX", "EXPIRES", "CREATED"}
		rows := make([][]string, len(keys))
		for i, k := range keys {
			expiresStr := "never"
			if k.ExpiresAt != nil {
				expiresStr = k.ExpiresAt.Format("2006-01-02")
			}
			rows[i] = []string{
				k.ID,
				k.Name,
				k.Prefix,
				expiresStr,
				k.CreatedAt.Format("2006-01-02"),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
