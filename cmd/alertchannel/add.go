package alertchannel

import (
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	addName       string
	addType       string
	addConfigArgs []string
)

var validTypes = []string{"email", "webhook", "slack", "discord", "teams", "pagerduty", "telegram", "googlechat"}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new alert channel",
	Long: `Create a new alert channel for downtime notifications.

Examples:
  spork alert-channel add --type email --name "Ops Team" --config to=ops@example.com
  spork alert-channel add --type slack --name "Eng Slack" --config url=https://hooks.slack.com/...
  spork alert-channel add --type webhook --name "Custom Hook" --config url=https://example.com/webhook
  spork alert-channel add --type pagerduty --name "PD Oncall" --config integration_key=abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		// Validate type
		validType := false
		for _, t := range validTypes {
			if addType == t {
				validType = true
				break
			}
		}
		if !validType {
			fmt.Fprintf(os.Stderr, "Error: invalid type %q. Must be one of: %s\n", addType, strings.Join(validTypes, ", "))
			return fmt.Errorf("invalid type: %s", addType)
		}

		// Parse --config key=value pairs
		config := make(map[string]string)
		for _, kv := range addConfigArgs {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Error: invalid config format %q, expected key=value\n", kv)
				return fmt.Errorf("invalid config: %s", kv)
			}
			config[parts[0]] = parts[1]
		}

		ch := &api.AlertChannel{
			Name:   addName,
			Type:   addType,
			Config: config,
		}

		result, err := client.CreateAlertChannel(ch)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating alert channel: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
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
