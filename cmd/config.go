package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage Craft CLI configuration settings and profiles.

Configuration is stored in: ~/.craft-cli/config.json

You can edit this file directly or use these commands to manage it.

Examples:
  # Add typed REST and MCP profiles
  craft config add-rest work --api-url https://connect.craft.do/links/LINK/api/v1
  craft config add-mcp work-mcp --mcp-url https://mcp.craft.do/links/LINK/mcp

  # Add a legacy REST profile (permissions set in Craft)
  craft config add work https://connect.craft.do/links/LINK/api/v1

  # Add a legacy REST profile with API key authentication
  craft config add secure https://connect.craft.do/.../api/v1 --key pdk_xxx

  # List all profiles (* = active, [key] = has API key)
  craft config list

  # Switch active profile
  craft config use personal

  # Remove a profile
  craft config remove old-profile

REST vs MCP:
  'craft config add' is kept for legacy REST profiles. For Craft MCP,
  use 'craft config add-mcp' or 'craft profiles add-mcp'. Do not edit
  mcp_url by hand unless the profile also has "type": "mcp".

Permissions:
  Both public links and API keys can have different permission levels
  configured in Craft (not in this CLI):
    - Read-only:  Can list, get, and search documents
    - Write-only: Can create, update, and delete documents
    - Read-write: Full access to all operations

  If you get PERMISSION_DENIED errors, check the link/key permissions
  in your Craft workspace settings. Use 'craft info --test-permissions'
  to see what your current profile can do.`,
}

var (
	profileAPIKey           string
	configProfilesAPIKey    string
	configProfilesAPIKeyEnv string
	configRESTAccessMode    string
	configRESTPermission    string
	configRESTDocumentScope string
	configMCPAccessMode     string
	configMCPPermission     string
	configMCPDocumentScope  string
)

var addProfileCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add or update a legacy REST profile",
	Long: `Add a new legacy REST API profile or update an existing one.

Optionally include an API key for authentication:
  craft config add myspace https://connect.craft.do/links/abc123/api/v1 --key pdk_xxxx

For MCP connections, use:
  craft config add-mcp myspace-mcp --mcp-url https://mcp.craft.do/links/abc123/mcp`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url := args[1]
		if err := cfgManager.AddProfileWithKey(name, url, profileAPIKey); err != nil {
			return fmt.Errorf("failed to add profile: %w", err)
		}
		if profileAPIKey != "" {
			fmt.Printf("Profile '%s' added (with API key)\n", name)
		} else {
			fmt.Printf("Profile '%s' added\n", name)
		}
		return nil
	},
}

var configAddRESTCmd = &cobra.Command{
	Use:   "add-rest <name>",
	Short: "Add or update a REST API profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("api-url")
		if url == "" {
			return fmt.Errorf("--api-url is required")
		}
		apiKey := configProfilesAPIKey
		if configProfilesAPIKeyEnv != "" {
			apiKey = os.Getenv(configProfilesAPIKeyEnv)
			if apiKey == "" {
				return fmt.Errorf("--api-key-env %s is not set", configProfilesAPIKeyEnv)
			}
		}
		if err := cfgManager.AddRESTProfile(args[0], url, apiKey, configRESTAccessMode, configRESTPermission, configRESTDocumentScope); err != nil {
			return fmt.Errorf("failed to add REST profile: %w", err)
		}
		if !isQuiet() {
			fmt.Printf("REST profile '%s' saved\n", args[0])
		}
		return nil
	},
}

var configAddMCPCmd = &cobra.Command{
	Use:   "add-mcp <name>",
	Short: "Add or update a Craft MCP profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("mcp-url")
		if url == "" {
			return fmt.Errorf("--mcp-url is required")
		}
		if err := cfgManager.AddMCPProfile(args[0], url, configMCPAccessMode, configMCPPermission, configMCPDocumentScope); err != nil {
			return fmt.Errorf("failed to add MCP profile: %w", err)
		}
		if !isQuiet() {
			fmt.Printf("MCP profile '%s' saved\n", args[0])
		}
		return nil
	},
}

var removeProfileCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a profile",
	Long:  "Delete a saved API profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := cfgManager.RemoveProfile(name); err != nil {
			return fmt.Errorf("failed to remove profile: %w", err)
		}
		fmt.Printf("Profile '%s' removed\n", name)
		return nil
	},
}

var useProfileCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch active profile",
	Long:  "Set which profile to use for API requests",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := cfgManager.UseProfile(name); err != nil {
			return fmt.Errorf("failed to switch profile: %w", err)
		}
		fmt.Printf("Switched to profile '%s'\n", name)
		return nil
	},
}

var listProfilesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  "Show all saved API profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := cfgManager.ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles configured. Run 'craft config add-rest <name> --api-url URL' or 'craft config add-mcp <name> --mcp-url URL' to add one.")
			return nil
		}

		for _, p := range profiles {
			marker := "  "
			if p.Active {
				marker = "* "
			}
			keyIndicator := ""
			if p.HasAPIKey {
				keyIndicator = " [key]"
			}
			target := p.URL
			if p.Type == "mcp" {
				target = p.MCPURL
			}
			fmt.Printf("%s%-12s %-4s %s%s\n", marker, p.Name, p.Type, target, keyIndicator)
		}
		return nil
	},
}

var forceReset bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all configuration",
	Long:  "Remove the configuration file and reset all settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Skip confirmation if --force flag is used or in quiet mode
		if !forceReset && !isQuiet() {
			profiles, _ := cfgManager.ListProfiles()
			if len(profiles) > 0 {
				fmt.Println("This will delete all profiles:")
				for _, p := range profiles {
					fmt.Printf("  - %s\n", p.Name)
				}
				fmt.Println()
				fmt.Print("Are you sure? (y/N): ")

				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" && response != "yes" {
					fmt.Println("Reset cancelled")
					return nil
				}
			}
		}

		if err := cfgManager.Reset(); err != nil {
			return fmt.Errorf("failed to reset config: %w", err)
		}
		fmt.Println("Configuration reset successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(addProfileCmd)
	configCmd.AddCommand(configAddRESTCmd)
	configCmd.AddCommand(configAddMCPCmd)
	configCmd.AddCommand(removeProfileCmd)
	configCmd.AddCommand(useProfileCmd)
	configCmd.AddCommand(listProfilesCmd)
	configCmd.AddCommand(resetCmd)

	addProfileCmd.Flags().StringVarP(&profileAPIKey, "key", "k", "", "API key for authentication")
	configAddRESTCmd.Flags().String("api-url", "", "Craft REST API URL")
	configAddRESTCmd.Flags().StringVar(&configProfilesAPIKey, "api-key", "", "API key for authentication")
	configAddRESTCmd.Flags().StringVar(&configProfilesAPIKeyEnv, "api-key-env", "", "Environment variable containing API key")
	configAddRESTCmd.Flags().StringVar(&configRESTAccessMode, "access", "", "Access mode: public or api-key")
	configAddRESTCmd.Flags().StringVar(&configRESTPermission, "permission", "", "Permission: read-only, write-only, read-write")
	configAddRESTCmd.Flags().StringVar(&configRESTDocumentScope, "scope", "", "Scope: all-documents, selected-documents, daily-notes")
	configAddMCPCmd.Flags().String("mcp-url", "", "Craft MCP URL")
	configAddMCPCmd.Flags().StringVar(&configMCPAccessMode, "access", "public", "Access mode: public or api-key")
	configAddMCPCmd.Flags().StringVar(&configMCPPermission, "permission", "", "Permission: read-only, write-only, read-write")
	configAddMCPCmd.Flags().StringVar(&configMCPDocumentScope, "scope", "connection-defined", "Scope: connection-defined, all-documents, selected-documents, daily-notes")
	resetCmd.Flags().BoolVarP(&forceReset, "force", "f", false, "Skip confirmation prompt")
}
