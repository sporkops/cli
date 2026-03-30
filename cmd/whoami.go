package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/auth"
	"github.com/sporkops/cli/internal/output"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current account info",
	Long:  "Display your account details: email, plan, monitor limit, check interval, and member since date.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := auth.LoadToken()
		if err != nil {
			return fmt.Errorf("loading credentials: %w", err)
		}
		if token == "" {
			fmt.Fprintln(os.Stderr, "Not logged in. Run: spork login")
			return fmt.Errorf("not logged in")
		}

		client := api.NewClient(token)
		account, err := client.GetAccount()
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 401 {
				fmt.Fprintln(os.Stderr, "Session expired. Run: spork login")
			}
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(account)
		}

		fmt.Printf("%-17s%s\n", "Email:", account.Email)
		fmt.Printf("%-17s%s\n", "Plan:", account.Plan)
		fmt.Printf("%-17s%d\n", "Monitor Limit:", account.MonitorLimit)
		fmt.Printf("%-17s%ds\n", "Check Interval:", account.CheckIntervalS)
		fmt.Printf("%-17s%s\n", "Member Since:", account.CreatedAt.Format("2006-01-02"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
