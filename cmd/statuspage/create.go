package statuspage

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sporkops/cli/internal/api"
	"github.com/sporkops/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	createName             string
	createSlug             string
	createTheme            string
	createAccentColor      string
	createLogoURL          string
	createPublic           bool
	createPassword         string
	createDomain           string
	createComponents       []string
	createComponentGroups  []string
	createEmailSubscribers bool
	createFontFamily       string
	createHeaderStyle      string
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new status page",
	Long: `Create a new public status page for your monitors.

Examples:
  spork status-page create --name "Acme Status" --slug acme-status
  spork status-page create --name "Acme Status" --slug acme-status --theme dark
  spork status-page create --name "Acme Status" --slug acme-status \
    --component monitor_id=mon_abc,name=API,order=1 \
    --component monitor_id=mon_def,name=Website,order=2
  spork status-page create --name "Acme Status" --slug acme-status \
    --component-group name=Core,order=1 \
    --component monitor_id=mon_abc,name=API,group_id=grp_1,order=1

Component format: monitor_id=<id>,name=<display_name>[,description=<text>][,group_id=<id>][,order=<n>]
Component group format: name=<name>[,order=<n>]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		// Validate slug format
		if !slugRegex.MatchString(createSlug) {
			return fmt.Errorf("invalid --slug %q: must be 2-63 lowercase alphanumeric characters or hyphens, cannot start/end with hyphen", createSlug)
		}

		// Validate theme
		if createTheme != "light" && createTheme != "dark" && createTheme != "blue" && createTheme != "midnight" {
			return fmt.Errorf("invalid --theme %q: must be light, dark, blue, or midnight", createTheme)
		}

		// Validate accent color if provided
		if createAccentColor != "" {
			matched, _ := regexp.MatchString(`^#[0-9a-fA-F]{6}$`, createAccentColor)
			if !matched {
				return fmt.Errorf("invalid --accent-color %q: must be a hex color like #ff0000", createAccentColor)
			}
		}

		validFontFamilies := map[string]bool{"system": true, "sans-serif": true, "serif": true, "monospace": true}
		if !validFontFamilies[createFontFamily] {
			return fmt.Errorf("invalid --font-family %q: must be system, sans-serif, serif, or monospace", createFontFamily)
		}
		validHeaderStyles := map[string]bool{"default": true, "banner": true, "minimal": true}
		if !validHeaderStyles[createHeaderStyle] {
			return fmt.Errorf("invalid --header-style %q: must be default, banner, or minimal", createHeaderStyle)
		}

		// Validate logo URL if provided
		if createLogoURL != "" && !strings.HasPrefix(createLogoURL, "https://") {
			return fmt.Errorf("invalid --logo-url %q: must start with https://", createLogoURL)
		}

		// Parse components
		components, err := parseComponents(createComponents)
		if err != nil {
			return err
		}

		// Parse component groups
		groups, err := parseComponentGroups(createComponentGroups)
		if err != nil {
			return err
		}

		sp := &api.StatusPage{
			Name:                    createName,
			Slug:                    createSlug,
			Theme:                   createTheme,
			AccentColor:             createAccentColor,
			FontFamily:              createFontFamily,
			HeaderStyle:             createHeaderStyle,
			LogoURL:                 createLogoURL,
			IsPublic:                createPublic,
			Password:                createPassword,
			Components:              components,
			ComponentGroups:         groups,
			EmailSubscribersEnabled: createEmailSubscribers,
		}

		result, err := client.CreateStatusPage(sp)
		if err != nil {
			if handleAPIError(err) {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error creating status page: %s\n", err)
			return err
		}

		// Set custom domain if specified
		if createDomain != "" {
			if err := client.SetCustomDomain(result.ID, createDomain); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: status page created but custom domain failed: %s\n", err)
			} else {
				result.CustomDomain = createDomain
				result.DomainStatus = "pending"
			}
		}

		if cmd.Root().Flag("json").Changed {
			return output.PrintJSON(result)
		}

		fmt.Printf("✓ Status page created: %s\n", result.Name)
		fmt.Printf("  URL: https://%s.status.sporkops.com\n", result.Slug)
		if result.CustomDomain != "" {
			fmt.Printf("  Custom domain: %s (status: %s)\n", result.CustomDomain, result.DomainStatus)
			fmt.Println("  CNAME your domain to status.sporkops.com")
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "status page name (required)")
	createCmd.Flags().StringVar(&createSlug, "slug", "", "URL slug (required, 2-63 lowercase alphanumeric/hyphens)")
	createCmd.Flags().StringVar(&createTheme, "theme", "light", "color theme: light, dark")
	createCmd.Flags().StringVar(&createAccentColor, "accent-color", "", "accent color as hex (e.g. #0066ff)")
	createCmd.Flags().StringVar(&createLogoURL, "logo-url", "", "logo URL (must be https)")
	createCmd.Flags().BoolVar(&createPublic, "public", true, "whether the status page is publicly accessible")
	createCmd.Flags().StringVar(&createPassword, "password", "", "password for private status pages (used when --public=false)")
	createCmd.Flags().StringVar(&createDomain, "domain", "", "custom domain (requires CNAME to status.sporkops.com)")
	createCmd.Flags().StringArrayVar(&createComponents, "component", nil, "component as monitor_id=<id>,name=<name>[,description=<text>][,group_id=<id>][,order=<n>] (repeatable)")
	createCmd.Flags().StringArrayVar(&createComponentGroups, "component-group", nil, "component group as name=<name>[,order=<n>] (repeatable)")
	createCmd.Flags().StringVar(&createFontFamily, "font-family", "system", "Font family for the status page (system, sans-serif, serif, monospace)")
	createCmd.Flags().StringVar(&createHeaderStyle, "header-style", "default", "Header style for the status page (default, banner, minimal)")
	createCmd.Flags().BoolVar(&createEmailSubscribers, "email-subscribers", false, "enable email subscriber notifications")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("slug")
}

// parseComponents parses --component flags into StatusComponent structs.
// Format: monitor_id=<id>,name=<display_name>[,description=<text>][,order=<n>]
func parseComponents(args []string) ([]api.StatusComponent, error) {
	if len(args) == 0 {
		return nil, nil
	}

	components := make([]api.StatusComponent, 0, len(args))
	for i, arg := range args {
		fields := make(map[string]string)
		for _, pair := range strings.Split(arg, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid component format %q: expected key=value pairs separated by commas", arg)
			}
			fields[parts[0]] = parts[1]
		}

		monitorID, ok := fields["monitor_id"]
		if !ok || monitorID == "" {
			return nil, fmt.Errorf("component %d: monitor_id is required", i+1)
		}
		name, ok := fields["name"]
		if !ok || name == "" {
			return nil, fmt.Errorf("component %d: name is required", i+1)
		}

		comp := api.StatusComponent{
			MonitorID:   monitorID,
			DisplayName: name,
			Description: fields["description"],
			GroupID:     fields["group_id"],
			Order:       i,
		}

		if orderStr, ok := fields["order"]; ok {
			order, err := strconv.Atoi(orderStr)
			if err != nil {
				return nil, fmt.Errorf("component %d: invalid order %q", i+1, orderStr)
			}
			comp.Order = order
		}

		components = append(components, comp)
	}

	return components, nil
}

// parseComponentGroups parses --component-group flags into ComponentGroup structs.
// Format: name=<name>[,order=<n>]
func parseComponentGroups(args []string) ([]api.ComponentGroup, error) {
	if len(args) == 0 {
		return nil, nil
	}

	groups := make([]api.ComponentGroup, 0, len(args))
	for i, arg := range args {
		fields := make(map[string]string)
		for _, pair := range strings.Split(arg, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid component-group format %q: expected key=value pairs separated by commas", arg)
			}
			fields[parts[0]] = parts[1]
		}

		name, ok := fields["name"]
		if !ok || name == "" {
			return nil, fmt.Errorf("component-group %d: name is required", i+1)
		}

		group := api.ComponentGroup{
			Name:  name,
			Order: i,
		}

		if orderStr, ok := fields["order"]; ok {
			order, err := strconv.Atoi(orderStr)
			if err != nil {
				return nil, fmt.Errorf("component-group %d: invalid order %q", i+1, orderStr)
			}
			group.Order = order
		}

		groups = append(groups, group)
	}

	return groups, nil
}
