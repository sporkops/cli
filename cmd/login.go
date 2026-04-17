package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/auth"
	"github.com/spf13/cobra"
)

// loginAPIKeyFromStdin is set true when the user invokes `spork login --api-key`
// (no value) or `spork login --api-key -`. The key is read from stdin rather
// than from argv so it cannot leak into the process list, shell history, or
// CI logs.
var loginAPIKeyFromStdin bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Spork",
	Long: "Authenticate with Spork.\n\n" +
		"By default, prints a URL and code to authenticate at sporkops.com/cli-auth.\n\n" +
		"For headless environments, either export SPORK_API_KEY or run\n" +
		"'spork login --api-key' (the key is read from stdin so it cannot leak\n" +
		"into process listings or CI logs).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginAPIKeyFromStdin {
			apiKey, err := readAPIKey(cmd.InOrStdin(), cmd.OutOrStderr())
			if err != nil {
				return err
			}
			if err := auth.SaveToken(apiKey); err != nil {
				return fmt.Errorf("saving credentials: %w", err)
			}
			fmt.Println("✓ API key saved")
			return nil
		}

		// Interactive device-code flow
		apiKey, err := auth.Login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %s\n\n", err)
			fmt.Fprintln(os.Stderr, "  Try signing up at: https://sporkops.com/signup?ref=cli")
			fmt.Fprintln(os.Stderr, "  Docs: https://sporkops.com/docs")
			return err
		}

		if err := auth.SaveToken(apiKey); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}
		fmt.Println("✓ Logged in successfully")
		return nil
	},
}

// readAPIKey pulls an API key from stdin. When stdin is a TTY it prompts
// the user; when it is a pipe it reads the first non-empty line. The full
// contents are never echoed.
func readAPIKey(in io.Reader, prompt io.Writer) (string, error) {
	if f, ok := in.(*os.File); ok {
		fi, err := f.Stat()
		if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
			fmt.Fprint(prompt, "Enter your API key: ")
		}
	}
	scanner := bufio.NewScanner(in)
	// Allow long keys (default bufio max is 64 KiB which is plenty).
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key == "" {
			continue
		}
		return key, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading API key: %w", err)
	}
	return "", fmt.Errorf("no API key provided on stdin")
}

func init() {
	// NoOptDefVal lets users pass `--api-key` without a value. We deliberately
	// do not accept `--api-key=sk_...` because values on argv are visible in
	// /proc/<pid>/cmdline, shell history, and CI logs. If a value is supplied
	// anyway we treat it as a request to read from stdin (ignored) and warn.
	loginCmd.Flags().BoolVar(&loginAPIKeyFromStdin, "api-key", false,
		"read API key from stdin for headless / non-interactive login (value on command line is NOT accepted)")
}
