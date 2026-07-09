package cmd

import (
	"fmt"
	"os"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/config"
	craftmcp "github.com/ashrafali/craft-cli/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	profilesAPIKey            string
	profilesAPIKeyEnv         string
	profilesRESTAccessMode    string
	profilesRESTPermission    string
	profilesRESTDocumentScope string
	profilesMCPAccessMode     string
	profilesMCPPermission     string
	profilesMCPDocumentScope  string
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage REST and MCP profiles",
	Long: `Manage named Craft REST and MCP profiles.

Existing 'craft config' commands remain supported for REST API profiles.

Examples:
  craft profiles add-rest work --api-url https://connect.craft.do/links/.../api/v1
  craft profiles add-mcp mcp --mcp-url https://mcp.craft.do/links/.../mcp
  craft profiles list
  craft profiles use work
  craft profiles test mcp
  craft profiles capabilities work
  craft profiles remove old --dry-run`,
}

var profilesAddRESTCmd = &cobra.Command{
	Use:   "add-rest <name>",
	Short: "Add or update a REST API profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("api-url")
		if url == "" {
			return fmt.Errorf("--api-url is required")
		}
		apiKey := profilesAPIKey
		if profilesAPIKeyEnv != "" {
			apiKey = os.Getenv(profilesAPIKeyEnv)
			if apiKey == "" {
				return fmt.Errorf("--api-key-env %s is not set", profilesAPIKeyEnv)
			}
		}
		if err := cfgManager.AddRESTProfile(args[0], url, apiKey, profilesRESTAccessMode, profilesRESTPermission, profilesRESTDocumentScope); err != nil {
			return err
		}
		if !isQuiet() {
			fmt.Printf("REST profile '%s' saved\n", args[0])
		}
		return nil
	},
}

var profilesAddMCPCmd = &cobra.Command{
	Use:   "add-mcp <name>",
	Short: "Add or update a Craft MCP profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("mcp-url")
		if url == "" {
			return fmt.Errorf("--mcp-url is required")
		}
		if err := cfgManager.AddMCPProfile(args[0], url, profilesMCPAccessMode, profilesMCPPermission, profilesMCPDocumentScope); err != nil {
			return err
		}
		if !isQuiet() {
			fmt.Printf("MCP profile '%s' saved\n", args[0])
		}
		return nil
	},
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List REST and MCP profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := cfgManager.ListProfiles()
		if err != nil {
			return err
		}
		if isJSONFormat(getOutputFormat()) {
			return outputJSON(profiles)
		}
		for _, p := range profiles {
			marker := "  "
			if p.Active {
				marker = "* "
			}
			target := p.URL
			if p.Type == "mcp" {
				target = p.MCPURL
			}
			keyIndicator := ""
			if p.HasAPIKey {
				keyIndicator = " [key]"
			}
			fmt.Printf("%s%-12s %-4s %s%s\n", marker, p.Name, p.Type, target, keyIndicator)
		}
		return nil
	},
}

var profilesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := cfgManager.GetProfile(args[0])
		if err != nil {
			return err
		}
		if profile.APIKey != "" {
			profile.APIKey = "***redacted***"
		}
		return outputJSON(profile)
	},
}

var profilesUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgManager.UseProfile(args[0]); err != nil {
			return err
		}
		if !isQuiet() {
			fmt.Printf("Switched to profile '%s'\n", args[0])
		}
		return nil
	},
}

var profilesTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test a profile connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := cfgManager.GetProfile(args[0])
		if err != nil {
			return err
		}

		result := map[string]interface{}{
			"name":           args[0],
			"type":           profile.TypeOrDefault(),
			"access_mode":    profile.AccessMode,
			"permission":     profile.Permission,
			"document_scope": profile.DocumentScope,
		}

		switch profile.TypeOrDefault() {
		case "rest":
			var client *api.Client
			if profile.APIKey != "" {
				client = api.NewClientWithKey(profile.URL, profile.APIKey)
			} else {
				client = api.NewClient(profile.URL)
			}
			_, err = client.GetConnection()
		case "mcp":
			_, err = craftmcp.NewClient(profile.MCPURL).Initialize()
		default:
			err = fmt.Errorf("unknown profile type: %s", profile.TypeOrDefault())
		}
		result["ok"] = err == nil
		if err != nil {
			result["error"] = err.Error()
		}
		return outputJSON(result)
	},
}

var profilesCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <name>",
	Short: "Show profile capability metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := cfgManager.GetProfile(args[0])
		if err != nil {
			return err
		}
		result := map[string]interface{}{
			"name":           args[0],
			"type":           profile.TypeOrDefault(),
			"access_mode":    profile.AccessMode,
			"permission":     profile.Permission,
			"document_scope": profile.DocumentScope,
			"backends":       []string{profile.TypeOrDefault()},
			"capabilities":   inferProfileCapabilities(profile),
		}
		return outputJSON(result)
	},
}

var profilesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			return dryRunOutput("remove profile", map[string]interface{}{
				"name": args[0], "destructive": true,
			})
		}
		if err := cfgManager.RemoveProfile(args[0]); err != nil {
			return err
		}
		if !isQuiet() {
			fmt.Printf("Removed profile '%s'\n", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.AddCommand(profilesAddRESTCmd)
	profilesCmd.AddCommand(profilesAddMCPCmd)
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesUseCmd)
	profilesCmd.AddCommand(profilesTestCmd)
	profilesCmd.AddCommand(profilesCapabilitiesCmd)
	profilesCmd.AddCommand(profilesRemoveCmd)

	profilesAddRESTCmd.Flags().String("api-url", "", "Craft REST API URL")
	profilesAddRESTCmd.Flags().StringVar(&profilesAPIKey, "api-key", "", "API key for authentication")
	profilesAddRESTCmd.Flags().StringVar(&profilesAPIKeyEnv, "api-key-env", "", "Environment variable containing API key")
	profilesAddRESTCmd.Flags().StringVar(&profilesRESTAccessMode, "access", "", "Access mode: public or api-key")
	profilesAddRESTCmd.Flags().StringVar(&profilesRESTPermission, "permission", "", "Permission: read-only, write-only, read-write")
	profilesAddRESTCmd.Flags().StringVar(&profilesRESTDocumentScope, "scope", "", "Scope: all-documents, selected-documents, daily-notes")

	profilesAddMCPCmd.Flags().String("mcp-url", "", "Craft MCP URL")
	profilesAddMCPCmd.Flags().StringVar(&profilesMCPAccessMode, "access", "public", "Access mode: public or api-key")
	profilesAddMCPCmd.Flags().StringVar(&profilesMCPPermission, "permission", "", "Permission: read-only, write-only, read-write")
	profilesAddMCPCmd.Flags().StringVar(&profilesMCPDocumentScope, "scope", "connection-defined", "Scope: connection-defined, all-documents, selected-documents, daily-notes")
}

func inferProfileCapabilities(profile config.Profile) []string {
	capabilities := []string{}
	switch profile.TypeOrDefault() {
	case "rest":
		capabilities = append(capabilities, "rest")
	case "mcp":
		capabilities = append(capabilities, "mcp", "mcp.tools")
	}
	switch profile.Permission {
	case "read-only":
		capabilities = append(capabilities, "read")
	case "write-only":
		capabilities = append(capabilities, "write")
	case "read-write", "":
		capabilities = append(capabilities, "read", "write")
	}
	return uniqueStrings(capabilities)
}
