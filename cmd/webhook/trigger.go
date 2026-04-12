package webhook

import (
	"context"
	"fmt"
	"os"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
)

var (
	triggerChannel string
)

var triggerCmd = &cobra.Command{
	Use:   "trigger <event>",
	Short: "Fire a synthetic, signed event at a webhook alert channel",
	Long: `Fire a synthetic, signed event at a webhook alert channel.

The server builds a realistic WebhookPayload marked as a test delivery,
signs it with the channel's configured secret using the same
X-Sporkops-Signature scheme production deliveries use, and POSTs it to
the channel's URL. Unlike production deliveries the trigger does not
retry — you see the receiver's first response directly, which is what
you want for debugging.

Events: monitor.down, monitor.up`,
	Example: `  spork webhook trigger monitor.down --channel ach_abc
  spork webhook trigger monitor.up --channel ach_abc --json`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"monitor.down", "monitor.up"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if triggerChannel == "" {
			return fmt.Errorf("--channel is required")
		}

		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		event := args[0]
		result, err := client.TriggerWebhook(context.Background(), &spork.TriggerWebhookInput{
			AlertChannelID: triggerChannel,
			Event:          event,
		})
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error triggering webhook: %s\n", err)
			return err
		}

		if cmdutil.Structured(cmd) {
			return cmdutil.PrintStructured(cmd, result)
		}

		renderTriggerResult(os.Stdout, event, triggerChannel, result)
		if !result.Delivered {
			// Non-zero exit for scripts that chain on delivery success.
			return fmt.Errorf("delivery failed")
		}
		return nil
	},
}

func init() {
	triggerCmd.Flags().StringVarP(&triggerChannel, "channel", "c", "", "alert channel ID to fire the event at (required)")
	_ = triggerCmd.MarkFlagRequired("channel")
}

// renderTriggerResult writes a human-readable summary of the delivery to
// w. Called when output is the default table mode; structured output
// (--json, --output yaml) uses the typed TriggerWebhookResult directly.
func renderTriggerResult(w *os.File, event, channelID string, r *spork.TriggerWebhookResult) {
	status := "FAILED"
	if r.Delivered {
		status = "DELIVERED"
	}
	status = output.ColorStatus(statusColorKey(r.Delivered)) + " " + status
	fmt.Fprintf(w, "\n  Event:       %s\n", event)
	fmt.Fprintf(w, "  Channel:     %s\n", channelID)
	fmt.Fprintf(w, "  Outcome:     %s\n", status)
	if r.StatusCode != 0 {
		fmt.Fprintf(w, "  HTTP status: %d\n", r.StatusCode)
	}
	fmt.Fprintf(w, "  Duration:    %dms\n", r.DurationMs)
	if r.Signature != "" {
		fmt.Fprintf(w, "  Signature:   %s\n", r.Signature)
	}
	if r.Error != "" {
		fmt.Fprintf(w, "  Error:       %s\n", r.Error)
	}
	if r.ResponsePreview != "" {
		fmt.Fprintf(w, "\n  Response body:\n    %s\n", indent(r.ResponsePreview, "    "))
	}
	fmt.Fprintln(w)
}

// statusColorKey maps the delivered flag to the semantic colour key the
// output package understands. "up" for delivered, "down" for failed — the
// semantics are "healthy integration" vs "broken integration."
func statusColorKey(delivered bool) string {
	if delivered {
		return "up"
	}
	return "down"
}

// indent prepends prefix to every line of s except the first (which is
// already indented by the caller). Keeps multi-line response previews
// aligned under "Response body:".
func indent(s, prefix string) string {
	out := ""
	first := true
	for _, line := range splitLines(s) {
		if first {
			out += line
			first = false
			continue
		}
		out += "\n" + prefix + line
	}
	return out
}

func splitLines(s string) []string {
	var lines []string
	var cur []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, string(cur))
			cur = cur[:0]
			continue
		}
		cur = append(cur, s[i])
	}
	lines = append(lines, string(cur))
	return lines
}
