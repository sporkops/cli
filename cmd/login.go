package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/auth"
)

var loginAPIKey string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Spork",
	Long: "Authenticate with Spork.\n\n" +
		"By default, prints a URL and code to authenticate at sporkops.com/cli-auth.\n\n" +
		"For headless environments, use --api-key to provide a key directly.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var apiKey string

		if cmd.Flags().Changed("api-key") {
			// --api-key mode: use provided value or prompt from stdin
			if loginAPIKey != "" {
				apiKey = strings.TrimSpace(loginAPIKey)
			} else {
				fmt.Print("Enter your API key: ")
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					apiKey = strings.TrimSpace(scanner.Text())
				}
				if apiKey == "" {
					return fmt.Errorf("no API key provided")
				}
			}

			if err := auth.SaveToken(apiKey); err != nil {
				return fmt.Errorf("saving credentials: %w", err)
			}
			fmt.Println("✓ API key saved")
			return nil
		}

		// Interactive device-code flow
		var err error
		apiKey, err = auth.Login()
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

func init() {
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key for headless / non-interactive login")
}
