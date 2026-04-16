package cmd

import (
	"context"
	"fmt"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "List available monitoring regions",
	Long:  "Display all monitoring regions available for uptime checks.",
	Example: `  spork regions
  spork regions --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		regions, err := client.ListRegions(context.Background())
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			return fmt.Errorf("listing regions: %w", err)
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, regions)
		}

		headers := []string{"ID", "NAME"}
		rows := make([][]string, len(regions))
		for i, r := range regions {
			rows[i] = []string{r.ID, r.Name}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(regionsCmd)
}
