package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"golang.org/x/term"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// colorEnabled returns true if color output should be used.
func colorEnabled() bool {
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
