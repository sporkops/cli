package apikey

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/sporkops/cli/internal/output"
)

var (
	createName    string
	createExpires int
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	Long:  "Create a new API key for programmatic access.\n\nExample:\n  spork api-key create --name \"CI deploy\"\n  spork api-key create --name \"CI deploy\" --expires 90",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		key, err := client.CreateAPIKey(createName, createExpires)
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating API key: %s\n", err)
			return err
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(key)
		}

		expiresStr := "never"
		if key.ExpiresAt != nil {
			expiresStr = key.ExpiresAt.Format("2006-01-02")
		}

		fmt.Printf("%-10s%s\n", "Name:", key.Name)
		fmt.Printf("%-10s%s\n", "Key:", key.Key)
		fmt.Printf("%-10s%s\n", "Prefix:", key.Prefix)
		fmt.Printf("%-10s%s\n", "Expires:", expiresStr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "⚠  Save your key now — it won't be shown again.")
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "name for the API key (required)")
	_ = createCmd.MarkFlagRequired("name")
	createCmd.Flags().IntVar(&createExpires, "expires", 0, "expiry in days (1-365, 0 = never expires)")
}
