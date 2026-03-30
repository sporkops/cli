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
var accentColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var validThemes = map[string]bool{"light": true, "dark": true, "blue": true, "midnight": true}
var validFontFamilies = map[string]bool{"system": true, "sans-serif": true, "serif": true, "monospace": true}
var validHeaderStyles = map[string]bool{"default": true, "banner": true, "minimal": true}

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
Component group format: name=<name>[,description=<text>][,order=<n>]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		// Validate slug format
		if !slugRegex.MatchString(createSlug) {
			return fmt.Errorf("invalid --slug %q: must be 2-63 lowercase alphanumeric characters or hyphens, cannot start/end with hyphen", createSlug)
		}

		if !validThemes[createTheme] {
			return fmt.Errorf("invalid --theme %q: must be light, dark, blue, or midnight", createTheme)
		}
		if createAccentColor != "" && !accentColorRegex.MatchString(createAccentColor) {
			return fmt.Errorf("invalid --accent-color %q: must be a hex color like #ff0000", createAccentColor)
		}
		if !validFontFamilies[createFontFamily] {
			return fmt.Errorf("invalid --font-family %q: must be system, sans-serif, serif, or monospace", createFontFamily)
		}
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
	createCmd.Flags().StringVar(&createTheme, "theme", "light", "color theme: light, dark, blue, midnight")
	createCmd.Flags().StringVar(&createAccentColor, "accent-color", "", "accent color as hex (e.g. #0066ff)")
	createCmd.Flags().StringVar(&createLogoURL, "logo-url", "", "logo URL (must be https)")
	createCmd.Flags().BoolVar(&createPublic, "public", true, "whether the status page is publicly accessible")
	createCmd.Flags().StringVar(&createPassword, "password", "", "password for private status pages (used when --public=false)")
	createCmd.Flags().StringVar(&createDomain, "domain", "", "custom domain (requires CNAME to status.sporkops.com)")
	createCmd.Flags().StringArrayVar(&createComponents, "component", nil, "component as monitor_id=<id>,name=<name>[,description=<text>][,group_id=<id>][,order=<n>] (repeatable)")
	createCmd.Flags().StringArrayVar(&createComponentGroups, "component-group", nil, "component group as name=<name>[,description=<text>][,order=<n>] (repeatable)")
	createCmd.Flags().StringVar(&createFontFamily, "font-family", "system", "Font family for the status page (system, sans-serif, serif, monospace)")
	createCmd.Flags().StringVar(&createHeaderStyle, "header-style", "default", "Header style for the status page (default, banner, minimal)")
	createCmd.Flags().BoolVar(&createEmailSubscribers, "email-subscribers", false, "enable email subscriber notifications")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("slug")
}

var componentValidKeys = map[string]bool{"monitor_id": true, "name": true, "description": true, "group_id": true, "order": true}

// parseComponents parses --component flags into StatusComponent structs.
// Format: monitor_id=<id>,name=<display_name>[,description=<text>][,group_id=<id>][,order=<n>]
func parseComponents(args []string) ([]api.StatusComponent, error) {
	if len(args) == 0 {
		return nil, nil
	}

	components := make([]api.StatusComponent, 0, len(args))
	for i, arg := range args {
		fields, err := parseKeyValuePairs(arg, componentValidKeys)
		if err != nil {
			return nil, fmt.Errorf("component %d: %w", i+1, err)
		}

		monitorID, ok := fields["monitor_id"]
		if !ok || monitorID == "" {
			return nil, fmt.Errorf("component %d: monitor_id is required", i+1)
		}
		name, ok := fields["name"]
		if !ok || name == "" {
			return nil, fmt.Errorf("component %d: name is required", i+1)
		}
		if len(name) > 100 {
			return nil, fmt.Errorf("component %d: name too long (max 100 characters)", i+1)
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

// parseKeyValuePairs parses a string like "name=Foo,description=Hello, World,order=1"
// into a map, handling commas inside values by scanning for known key boundaries.
func parseKeyValuePairs(arg string, validKeys map[string]bool) (map[string]string, error) {
	fields := make(map[string]string)

	// Find positions of all "key=" occurrences at comma boundaries or start of string
	type keyPos struct {
		key   string
		start int // index of the key in the arg
		valAt int // index where the value begins (after "key=")
	}
	var positions []keyPos
	for key := range validKeys {
		prefix := key + "="
		idx := 0
		for idx <= len(arg)-len(prefix) {
			pos := strings.Index(arg[idx:], prefix)
			if pos == -1 {
				break
			}
			absPos := idx + pos
			// Valid if at start of string or preceded by comma
			if absPos == 0 || arg[absPos-1] == ',' {
				positions = append(positions, keyPos{key: key, start: absPos, valAt: absPos + len(prefix)})
			}
			idx = absPos + 1
		}
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("no valid key=value pairs found in %q", arg)
	}

	// Sort by position
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[j].start < positions[i].start {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	// Extract values: each value runs from valAt to the start of the next key (minus the comma)
	for idx, kp := range positions {
		var val string
		if idx+1 < len(positions) {
			// Value ends at the comma before the next key
			val = arg[kp.valAt : positions[idx+1].start-1]
		} else {
			val = arg[kp.valAt:]
		}
		if _, exists := fields[kp.key]; exists {
			return nil, fmt.Errorf("duplicate key %q", kp.key)
		}
		fields[kp.key] = val
	}

	// Check for unexpected text before the first key
	if positions[0].start > 0 {
		return nil, fmt.Errorf("unexpected text before first key in %q", arg)
	}

	return fields, nil
}

var componentGroupValidKeys = map[string]bool{"name": true, "description": true, "order": true}

// parseComponentGroups parses --component-group flags into ComponentGroup structs.
// Format: name=<name>[,description=<text>][,order=<n>]
func parseComponentGroups(args []string) ([]api.ComponentGroup, error) {
	if len(args) == 0 {
		return nil, nil
	}

	groups := make([]api.ComponentGroup, 0, len(args))
	for i, arg := range args {
		fields, err := parseKeyValuePairs(arg, componentGroupValidKeys)
		if err != nil {
			return nil, fmt.Errorf("component-group %d: %w", i+1, err)
		}

		name, ok := fields["name"]
		if !ok || name == "" {
			return nil, fmt.Errorf("component-group %d: name is required", i+1)
		}
		if len(name) > 100 {
			return nil, fmt.Errorf("component-group %d: name too long (max 100 characters)", i+1)
		}

		desc := fields["description"]
		if len(desc) > 500 {
			return nil, fmt.Errorf("component-group %d: description too long (max 500 characters)", i+1)
		}

		group := api.ComponentGroup{
			Name:        name,
			Description: desc,
			Order:       i,
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
