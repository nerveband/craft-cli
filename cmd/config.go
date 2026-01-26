package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage Craft CLI configuration settings and API profiles.

Configuration is stored in: ~/.craft-cli/config.json

You can edit this file directly or use these commands to manage it.

Examples:
  # Add a public link (permissions set in Craft)
  craft config add work https://connect.craft.do/links/LINK/api/v1

  # Add a profile with API key authentication
  craft config add secure https://connect.craft.do/.../api/v1 --key pdk_xxx

  # List all profiles (* = active, [key] = has API key)
  craft config list

  # Switch active profile
  craft config use personal

  # Remove a profile
  craft config remove old-profile

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

var profileAPIKey string

var addProfileCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add or update a profile",
	Long: `Add a new API profile or update an existing one.

Optionally include an API key for authentication:
  craft config add myspace https://connect.craft.do/links/abc123/api/v1 --key pdk_xxxx`,
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
			fmt.Println("No profiles configured. Run 'craft config add <name> <url>' to add one.")
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
			fmt.Printf("%s%-12s %s%s\n", marker, p.Name, p.URL, keyIndicator)
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
	configCmd.AddCommand(removeProfileCmd)
	configCmd.AddCommand(useProfileCmd)
	configCmd.AddCommand(listProfilesCmd)
	configCmd.AddCommand(resetCmd)

	addProfileCmd.Flags().StringVarP(&profileAPIKey, "key", "k", "", "API key for authentication")
	resetCmd.Flags().BoolVarP(&forceReset, "force", "f", false, "Skip confirmation prompt")
}
