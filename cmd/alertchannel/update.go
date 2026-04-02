package alertchannel

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	updateName       string
	updateConfigArgs []string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing alert channel",
	Long: `Update an existing alert channel's name or configuration.

Examples:
  spork alert-channel update ch_abc123 --name "New Name"
  spork alert-channel update ch_abc123 --config to=newemail@example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id := args[0]

		// First fetch the existing channel so we can send a full PUT.
		existing, err := client.GetAlertChannel(context.Background(), id)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching alert channel: %s\n", err)
			return err
		}

		hasChanges := false

		if cmd.Flags().Changed("name") {
			existing.Name = updateName
			hasChanges = true
		}
		if cmd.Flags().Changed("config") {
			for _, kv := range updateConfigArgs {
				k, v, err := cmdutil.ParseKeyValue(kv)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid config: %s\n", err)
					return fmt.Errorf("invalid config: %s", kv)
				}
				existing.Config[k] = v
			}
			hasChanges = true
		}

		if !hasChanges {
			fmt.Fprintln(os.Stderr, "Nothing to update. Specify at least one flag:")
			fmt.Fprintln(os.Stderr, "  --name, --config")
			return fmt.Errorf("no changes specified")
		}

		result, err := client.UpdateAlertChannel(context.Background(), id, existing)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error updating alert channel: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
		}

		fmt.Printf("✓ Alert channel updated: %s (%s)\n", result.Name, result.Type)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "new channel name")
	updateCmd.Flags().StringArrayVar(&updateConfigArgs, "config", nil, "channel config as key=value (repeatable)")
}
