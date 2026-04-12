package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/spork-go"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var exportFormat string
var exportResourceName string

var exportCmd = &cobra.Command{
	Use:   "export <id|url>",
	Short: "Export a monitor as Terraform HCL, JSON, or YAML",
	Long: `Export a monitor's configuration in a format suitable for version control.

The default output (HCL) is a ready-to-use Terraform resource block that
'terraform apply' will produce an identical monitor for — use this when
you prototype a monitor in the UI or CLI and want to codify it in an
infrastructure repo.

JSON and YAML output round-trip through the same struct the API uses, so
they're appropriate for other tooling (jq pipelines, config-management
systems) that speaks those formats.`,
	Example: `  spork monitor export mon_abc123
  spork monitor export mon_abc123 --format yaml > monitor.yaml
  spork monitor export mon_abc123 --resource-name prod_api > prod_api.tf`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}
		id, _, err := resolveMonitorID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}
		m, err := client.GetMonitor(context.Background(), id)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching monitor: %s\n", err)
			return err
		}

		switch strings.ToLower(exportFormat) {
		case "", "hcl":
			return writeMonitorHCL(os.Stdout, m, exportResourceName)
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(m)
		case "yaml":
			// Round-trip through JSON so we honor the SDK's json tags
			// (same approach as output.PrintYAML).
			b, err := json.Marshal(m)
			if err != nil {
				return err
			}
			var intermediate any
			if err := json.Unmarshal(b, &intermediate); err != nil {
				return err
			}
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			if err := enc.Encode(intermediate); err != nil {
				return err
			}
			return enc.Close()
		default:
			return fmt.Errorf("invalid --format %q: must be hcl, json, or yaml", exportFormat)
		}
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "hcl", "output format: hcl, json, or yaml")
	exportCmd.Flags().StringVar(&exportResourceName, "resource-name", "", "Terraform resource name (default: derived from monitor name)")
	Cmd.AddCommand(exportCmd)
}

// writeMonitorHCL emits the monitor as a sporkops_monitor resource block.
// The output is deliberately formatted to match what terraform-registry docs
// show: one attribute per line, two-space indent, no trailing blank lines.
func writeMonitorHCL(w io.Writer, m *spork.Monitor, resourceName string) error {
	// Always run the caller-supplied name through hclIdentifier — a
	// user passing --resource-name "my monitor" would otherwise
	// produce invalid HCL (`resource "sporkops_monitor" "my monitor"`
	// doesn't parse). Sanitising here means the exporter is
	// copy-paste-safe regardless of how the flag was spelled.
	resourceName = hclIdentifier(resourceName)
	if resourceName == "" {
		resourceName = hclIdentifier(m.Name)
		if resourceName == "" {
			resourceName = hclIdentifier(m.ID)
		}
	}

	// Build attributes in a stable order. Terraform itself accepts any order
	// in HCL, but emitting the same order every time means an "export,
	// commit, re-export" loop produces empty diffs.
	var b strings.Builder
	fmt.Fprintf(&b, "# Exported by `spork monitor export %s` on %s\n", m.ID, "sporkops.com")
	fmt.Fprintf(&b, "# Review before apply: the Terraform provider is the source of truth once this\n")
	fmt.Fprintf(&b, "# resource is in state — the UI will show it as managed-in-code.\n")
	fmt.Fprintf(&b, "resource \"sporkops_monitor\" %q {\n", resourceName)

	writeStr(&b, "name", m.Name)
	writeStr(&b, "target", m.Target)
	if m.Type != "" && m.Type != "http" {
		writeStr(&b, "type", m.Type)
	}
	if m.Method != "" && m.Method != "GET" {
		writeStr(&b, "method", m.Method)
	}
	if m.ExpectedStatus > 0 && m.ExpectedStatus != 200 {
		writeInt(&b, "expected_status", int64(m.ExpectedStatus))
	}
	if m.Interval > 0 && m.Interval != 60 {
		writeInt(&b, "interval", int64(m.Interval))
	}
	if m.Timeout > 0 && m.Timeout != 30 {
		writeInt(&b, "timeout", int64(m.Timeout))
	}
	if len(m.Regions) > 0 && !(len(m.Regions) == 1 && m.Regions[0] == "us-central1") {
		writeStringList(&b, "regions", m.Regions)
	}
	if len(m.AlertChannelIDs) > 0 {
		writeStringList(&b, "alert_channel_ids", m.AlertChannelIDs)
	}
	if len(m.Tags) > 0 {
		writeStringList(&b, "tags", m.Tags)
	}
	if m.Paused != nil && *m.Paused {
		fmt.Fprintln(&b, "  paused            = true")
	}
	if m.Keyword != "" {
		writeStr(&b, "keyword", m.Keyword)
	}
	if m.KeywordType != "" {
		writeStr(&b, "keyword_type", m.KeywordType)
	}
	if m.SSLWarnDays > 0 {
		writeInt(&b, "ssl_warn_days", int64(m.SSLWarnDays))
	}
	if m.Body != "" {
		writeStr(&b, "body", m.Body)
	}
	if len(m.Headers) > 0 {
		writeStringMap(&b, "headers", m.Headers)
	}

	fmt.Fprintln(&b, "}")
	_, err := io.WriteString(w, b.String())
	return err
}

var hclIdentifierRe = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

// hclIdentifier turns an arbitrary string into a valid HCL resource label:
// lowercase, underscores-only, does not start with a digit. Empty strings
// round-trip to "" so callers can detect the fallback path.
func hclIdentifier(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = hclIdentifierRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return ""
	}
	if s[0] >= '0' && s[0] <= '9' {
		s = "m_" + s
	}
	return s
}

const hclIndent = "  "

func writeStr(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "%s%-17s = %q\n", hclIndent, key, value)
}

func writeInt(b *strings.Builder, key string, value int64) {
	fmt.Fprintf(b, "%s%-17s = %d\n", hclIndent, key, value)
}

func writeStringList(b *strings.Builder, key string, values []string) {
	// Emit inline for 1-2 elements, multiline otherwise — matches the
	// convention terraform fmt uses.
	if len(values) <= 2 {
		parts := make([]string, len(values))
		for i, v := range values {
			parts[i] = fmt.Sprintf("%q", v)
		}
		fmt.Fprintf(b, "%s%-17s = [%s]\n", hclIndent, key, strings.Join(parts, ", "))
		return
	}
	fmt.Fprintf(b, "%s%-17s = [\n", hclIndent, key)
	for _, v := range values {
		fmt.Fprintf(b, "%s%s%q,\n", hclIndent, hclIndent, v)
	}
	fmt.Fprintf(b, "%s]\n", hclIndent)
}

func writeStringMap(b *strings.Builder, key string, values map[string]string) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(b, "%s%-17s = {\n", hclIndent, key)
	for _, k := range keys {
		fmt.Fprintf(b, "%s%s%q = %q\n", hclIndent, hclIndent, k, values[k])
	}
	fmt.Fprintf(b, "%s}\n", hclIndent)
}
