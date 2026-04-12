package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// colorOverride, when non-nil, forces colors on (true) or off (false) and
// bypasses the NO_COLOR + TTY auto-detection. Set by SetColor(); the root
// command wires it from --no-color / --color=auto|always|never.
var colorOverride *bool

// SetColor overrides automatic color detection. Pass nil to restore the
// default (honor NO_COLOR + TTY).
func SetColor(enabled *bool) {
	colorOverride = enabled
}

// colorEnabled returns true if color output should be used.
func colorEnabled() bool {
	if colorOverride != nil {
		return *colorOverride
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// PrintTable prints a formatted table to stdout.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// ColorStatus returns a color-coded status string for terminal output.
// Falls back to plain text when stdout is not a terminal or NO_COLOR is set.
func ColorStatus(status string) string {
	if !colorEnabled() {
		return status
	}
	switch status {
	case "up", "success":
		return colorGreen + status + colorReset
	case "down", "fail":
		return colorRed + status + colorReset
	case "pending":
		return colorYellow + "pending" + colorReset
	case "degraded":
		return colorYellow + "degraded" + colorReset
	case "paused":
		return colorCyan + "paused" + colorReset
	default:
		return status
	}
}

// PrintJSON prints the value as formatted JSON to stdout.
func PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// PrintYAML prints the value as YAML to stdout. Uses 2-space indent to match
// Kubernetes/Helm conventions; works for any struct the SDK marshals as JSON
// (we round-trip through JSON so field names and omitempty behaviour match
// exactly what the API speaks — no second set of tags to keep in sync).
func PrintYAML(v any) error {
	// Round-trip through JSON so we pick up json tags (including
	// omitempty) without requiring every SDK struct to also carry yaml
	// tags. The intermediate map is then YAML-encoded.
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling for YAML: %w", err)
	}
	var intermediate any
	if err := json.Unmarshal(b, &intermediate); err != nil {
		return fmt.Errorf("unmarshaling for YAML: %w", err)
	}
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	if err := enc.Encode(intermediate); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return enc.Close()
}
