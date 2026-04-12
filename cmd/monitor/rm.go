package monitor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sporkops/cli/internal/cmdutil"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	forceRemove bool
	yesRemove   bool
)

var rmCmd = &cobra.Command{
	Use:   "rm <id|url>",
	Short: "Remove a monitor",
	Long:  "Remove an uptime monitor by ID or URL.",
	Example: `  spork monitor rm https://example.com
  spork monitor rm abc123 --yes
  spork monitor rm abc123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.RequireAuth()
		if err != nil {
			return err
		}

		id, name, err := resolveMonitorID(client, args[0])
		if err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		skipPrompt := forceRemove || yesRemove
		if !skipPrompt {
			// Skip prompt in non-interactive mode
			isJSON := cmd.Root().Flag("json").Changed
			if !term.IsTerminal(int(os.Stdout.Fd())) || isJSON {
				return fmt.Errorf("refusing to delete without --yes in non-interactive mode")
			}
			label := id
			if name != "" {
				label = name
			}
			fmt.Printf("Delete monitor %q? [y/N] ", label)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteMonitor(context.Background(), id); err != nil {
			if cmdutil.HandleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error removing monitor: %s\n", err)
			return err
		}

		label := id
		if name != "" {
			label = name
		}
		fmt.Printf("✓ Monitor removed: %s\n", label)
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "skip confirmation prompt")
	rmCmd.Flags().BoolVarP(&yesRemove, "yes", "y", false, "skip confirmation prompt")
}
