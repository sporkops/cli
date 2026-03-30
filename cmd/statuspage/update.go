package statuspage

import (
	"fmt"
	"os"

	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	updateName             string
	updateSlug             string
	updateTheme            string
	updateAccentColor      string
	updateLogoURL          string
	updatePublic           bool
	updatePassword         string
	updateDomain           string
	updateRemoveDomain     bool
	updateComponents       []string
	updateComponentGroups  []string
	updateEmailSubscribers bool
	updateFontFamily       string
	updateHeaderStyle      string
	updateWebhookURL       string
)

var updateCmd = &cobra.Command{
	Use:   "update <id|name|slug>",
	Short: "Update a status page",
	Long: `Update an existing status page's settings.

Example:
  spork status-page update acme-status --name "Acme Status v2"
  spork status-page update acme-status --theme dark --accent-color "#0066ff"
  spork status-page update acme-status --domain status.acme.com
  spork status-page update acme-status --remove-domain
  spork status-page update acme-status \
    --component monitor_id=mon_abc,name=API,order=1 \
    --component monitor_id=mon_def,name=Website,order=2

Component format: monitor_id=<id>,name=<display_name>[,description=<text>][,order=<n>]
Note: --component replaces all existing components.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		id, name, err := resolveStatusPageID(client, args[0])
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return err
		}

		// Fetch current state to merge with updates (PUT requires full object)
		current, err := client.GetStatusPage(id)
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error fetching status page: %s\n", err)
			return err
		}

		hasChanges := false
		hasDomainChange := false

		if cmd.Flags().Changed("name") {
			current.Name = updateName
			hasChanges = true
		}
		if cmd.Flags().Changed("slug") {
			if !slugRegex.MatchString(updateSlug) {
				return fmt.Errorf("invalid --slug %q: must be 2-63 lowercase alphanumeric characters or hyphens", updateSlug)
			}
			current.Slug = updateSlug
			hasChanges = true
		}
		if cmd.Flags().Changed("theme") {
			if !validThemes[updateTheme] {
				return fmt.Errorf("invalid --theme %q: must be light, dark, blue, or midnight", updateTheme)
			}
			current.Theme = updateTheme
			hasChanges = true
		}
		if cmd.Flags().Changed("font-family") {
			if !validFontFamilies[updateFontFamily] {
				return fmt.Errorf("invalid --font-family %q: must be system, sans-serif, serif, or monospace", updateFontFamily)
			}
			current.FontFamily = updateFontFamily
			hasChanges = true
		}
		if cmd.Flags().Changed("header-style") {
			if !validHeaderStyles[updateHeaderStyle] {
				return fmt.Errorf("invalid --header-style %q: must be default, banner, or minimal", updateHeaderStyle)
			}
			current.HeaderStyle = updateHeaderStyle
			hasChanges = true
		}
		if cmd.Flags().Changed("accent-color") {
			if updateAccentColor != "" && !accentColorRegex.MatchString(updateAccentColor) {
				return fmt.Errorf("invalid --accent-color %q: must be a hex color like #ff0000", updateAccentColor)
			}
			current.AccentColor = updateAccentColor
			hasChanges = true
		}
		if cmd.Flags().Changed("logo-url") {
			if updateLogoURL != "" && !strings.HasPrefix(updateLogoURL, "https://") {
				return fmt.Errorf("invalid --logo-url %q: must start with https://", updateLogoURL)
			}
			current.LogoURL = updateLogoURL
			hasChanges = true
		}
		if cmd.Flags().Changed("public") {
			current.IsPublic = updatePublic
			hasChanges = true
		}
		if cmd.Flags().Changed("password") {
			current.Password = updatePassword
			hasChanges = true
		}
		if cmd.Flags().Changed("component") {
			components, err := parseComponents(updateComponents)
			if err != nil {
				return err
			}
			current.Components = components
			hasChanges = true
		}
		if cmd.Flags().Changed("component-group") {
			groups, err := parseComponentGroups(updateComponentGroups)
			if err != nil {
				return err
			}
			current.ComponentGroups = groups
			hasChanges = true
		}
		if cmd.Flags().Changed("webhook-url") {
			if updateWebhookURL != "" && !strings.HasPrefix(updateWebhookURL, "https://") {
				return fmt.Errorf("invalid --webhook-url %q: must start with https://", updateWebhookURL)
			}
			current.WebhookURL = updateWebhookURL
			hasChanges = true
		}
		if cmd.Flags().Changed("email-subscribers") {
			current.EmailSubscribersEnabled = updateEmailSubscribers
			hasChanges = true
		}
		if cmd.Flags().Changed("domain") {
			hasDomainChange = true
		}
		if cmd.Flags().Changed("remove-domain") && updateRemoveDomain {
			hasDomainChange = true
		}

		if !hasChanges && !hasDomainChange {
			fmt.Fprintln(os.Stderr, "Nothing to update. Specify at least one flag:")
			fmt.Fprintln(os.Stderr, "  --name, --slug, --theme, --accent-color, --font-family, --header-style, --logo-url, --webhook-url, --public, --password, --domain, --component, --component-group, --email-subscribers")
			return fmt.Errorf("no changes specified")
		}

		// Update the status page resource
		if hasChanges {
			// Clear server-managed fields before sending
			sp := &api.StatusPage{
				Name:                    current.Name,
				Slug:                    current.Slug,
				Theme:                   current.Theme,
				AccentColor:             current.AccentColor,
				FontFamily:              current.FontFamily,
				HeaderStyle:             current.HeaderStyle,
				LogoURL:                 current.LogoURL,
				WebhookURL:              current.WebhookURL,
				IsPublic:                current.IsPublic,
				Password:                current.Password,
				Components:              current.Components,
				ComponentGroups:         current.ComponentGroups,
				EmailSubscribersEnabled: current.EmailSubscribersEnabled,
			}
			result, err := client.UpdateStatusPage(id, sp)
			if err != nil {
				if handleAPIError(err) {
					return err
				}
				fmt.Fprintf(os.Stderr, "Error updating status page: %s\n", err)
				return err
			}
			current = result
		}

		// Handle custom domain changes
		if hasDomainChange {
			if updateRemoveDomain {
				if err := client.RemoveCustomDomain(id); err != nil {
					if handleAPIError(err) {
						return err
					}
					fmt.Fprintf(os.Stderr, "Error removing custom domain: %s\n", err)
					return err
				}
				current.CustomDomain = ""
				current.DomainStatus = ""
			} else if updateDomain != "" {
				if err := client.SetCustomDomain(id, updateDomain); err != nil {
					if handleAPIError(err) {
						return err
					}
					fmt.Fprintf(os.Stderr, "Error setting custom domain: %s\n", err)
					return err
				}
				current.CustomDomain = updateDomain
				current.DomainStatus = "pending"
			}
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(current)
		}

		label := current.Name
		if label == "" {
			label = name
		}
		fmt.Printf("✓ Status page updated: %s\n", label)
		if current.CustomDomain != "" {
			fmt.Printf("  Custom domain: %s (%s)\n", current.CustomDomain, current.DomainStatus)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "new status page name")
	updateCmd.Flags().StringVar(&updateSlug, "slug", "", "new URL slug")
	updateCmd.Flags().StringVar(&updateTheme, "theme", "", "color theme: light, dark, blue, midnight")
	updateCmd.Flags().StringVar(&updateAccentColor, "accent-color", "", "accent color as hex (e.g. #0066ff)")
	updateCmd.Flags().StringVar(&updateLogoURL, "logo-url", "", "logo URL (must be https)")
	updateCmd.Flags().BoolVar(&updatePublic, "public", true, "whether the status page is publicly accessible")
	updateCmd.Flags().StringVar(&updatePassword, "password", "", "password for private status pages")
	updateCmd.Flags().StringVar(&updateDomain, "domain", "", "set custom domain (requires CNAME to status.sporkops.com)")
	updateCmd.Flags().BoolVar(&updateRemoveDomain, "remove-domain", false, "remove the custom domain")
	updateCmd.Flags().StringArrayVar(&updateComponents, "component", nil, "component as monitor_id=<id>,name=<name>[,description=<text>][,group_id=<id>][,order=<n>] (replaces all)")
	updateCmd.Flags().StringArrayVar(&updateComponentGroups, "component-group", nil, "component group as name=<name>[,order=<n>] (replaces all)")
	updateCmd.Flags().StringVar(&updateFontFamily, "font-family", "", "Font family for the status page (system, sans-serif, serif, monospace)")
	updateCmd.Flags().StringVar(&updateHeaderStyle, "header-style", "", "Header style for the status page (default, banner, minimal)")
	updateCmd.Flags().BoolVar(&updateEmailSubscribers, "email-subscribers", false, "enable/disable email subscriber notifications")
	updateCmd.Flags().StringVar(&updateWebhookURL, "webhook-url", "", "webhook URL for incident notifications (must be https)")
}
