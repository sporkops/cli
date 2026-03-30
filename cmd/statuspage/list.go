package statuspage

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all status pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		pages, err := client.ListStatusPages(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error listing status pages: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(pages)
		}

		if len(pages) == 0 {
			fmt.Println("No status pages yet. Create one:")
			fmt.Println("  spork status-page create --name \"My Status\" --slug my-status")
			return nil
		}

		headers := []string{"ID", "NAME", "SLUG", "COMPONENTS", "PUBLIC", "DOMAIN"}
		rows := make([][]string, len(pages))
		for i, p := range pages {
			domain := "-"
			if p.CustomDomain != "" {
				domain = p.CustomDomain
			}
			rows[i] = []string{
				p.ID,
				p.Name,
				p.Slug,
				strconv.Itoa(len(p.Components)),
				strconv.FormatBool(p.IsPublic),
				domain,
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}
