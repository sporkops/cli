package alertchannel

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	addName       string
	addType       string
	addConfigArgs []string
)

var validTypes = map[string]bool{
	"email": true, "webhook": true, "slack": true, "discord": true,
	"teams": true, "pagerduty": true, "telegram": true, "googlechat": true,
}

var addCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "Create a new alert channel",
	Long: `Create a new alert channel for downtime notifications.

Examples:
  spork alert-channel create --type email --name "Ops Team" --config to=ops@example.com
  spork alert-channel create --type slack --name "Eng Slack" --config url=https://hooks.slack.com/...
  spork alert-channel create --type webhook --name "Custom Hook" --config url=https://example.com/webhook
  spork alert-channel add --type pagerduty --name "PD Oncall" --config integration_key=abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		// Validate type
		if !validTypes[addType] {
			fmt.Fprintf(os.Stderr, "Error: invalid type %q. Must be one of: email, webhook, slack, discord, teams, pagerduty, telegram, googlechat\n", addType)
			return fmt.Errorf("invalid type: %s", addType)
		}

		// Parse --config key=value pairs
		config := make(map[string]string)
		for _, kv := range addConfigArgs {
			k, v, err := cmdutil.ParseKeyValue(kv)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid config: %s\n", err)
				return fmt.Errorf("invalid config: %s", kv)
			}
			config[k] = v
		}

		ch := &spork.AlertChannel{
			Name:   addName,
			Type:   addType,
			Config: config,
		}

		result, err := client.CreateAlertChannel(context.Background(), ch)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating alert channel: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}

		fmt.Printf("✓ Alert channel added: %s (%s)\n", result.Name, result.Type)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "channel name (required)")
	addCmd.Flags().StringVar(&addType, "type", "", "channel type: email, webhook, slack, discord, teams, pagerduty, telegram, googlechat (required)")
	addCmd.Flags().StringArrayVar(&addConfigArgs, "config", nil, "channel config as key=value (repeatable)")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("type")
	addCmd.MarkFlagRequired("config")
}
