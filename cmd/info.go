package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/spf13/cobra"
)

var testPermissions bool

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show API information and scope",
	Long:  "Display information about the configured API and available documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		// Get current config
		cfg, err := cfgManager.Load()
		if err != nil {
			return err
		}

		// Get config file path
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, config.ConfigDirName, config.ConfigFileName)

		fmt.Println("Craft CLI Information")
		fmt.Println("=====================")
		if cfg.ActiveProfile != "" {
			fmt.Printf("Active Profile:   %s\n", cfg.ActiveProfile)
			if profile, ok := cfg.Profiles[cfg.ActiveProfile]; ok {
				fmt.Printf("API URL:          %s\n", profile.URL)
				if profile.APIKey != "" {
					fmt.Printf("Authentication:   API Key (configured)\n")
				} else {
					fmt.Printf("Authentication:   Public Link\n")
				}
			}
		}
		fmt.Printf("Config File:      %s\n", configPath)
		fmt.Printf("Default Format:   %s\n", cfg.DefaultFormat)
		fmt.Println()

		// Test permissions if requested
		if testPermissions {
			fmt.Println("Testing Permissions...")
			fmt.Println()

			// Test read permission
			canRead := true
			_, err := client.GetDocuments()
			if err != nil {
				canRead = false
			}

			if canRead {
				fmt.Println("✓ Read:   Allowed")
			} else {
				fmt.Println("✗ Read:   Denied")
			}

			// Test write permission (dry-run create)
			// Note: We can't truly test write without making changes
			fmt.Println("  Write:  Use 'craft create --dry-run' to test")

			// Test delete permission
			// Note: Can't safely test without actually deleting
			fmt.Println("  Delete: Use 'craft delete --dry-run' to test")

			fmt.Println()
			fmt.Println("Note: Write and delete permissions are difficult to test")
			fmt.Println("without making actual changes. Try the operations with")
			fmt.Println("--dry-run flag to see if you get PERMISSION_DENIED errors.")
			fmt.Println()
		}

		// Try to fetch documents to show scope
		result, err := client.GetDocuments()
		if err != nil {
			fmt.Printf("Error fetching documents: %v\n", err)
			return nil
		}

		fmt.Printf("Total Documents: %d\n", len(result.Items))
		fmt.Println()

		if len(result.Items) > 0 {
			fmt.Println("Recent Documents:")
			limit := 5
			if len(result.Items) < limit {
				limit = len(result.Items)
			}
			for i := 0; i < limit; i++ {
				doc := result.Items[i]
				fmt.Printf("  - %s (ID: %s)\n", doc.Title, doc.ID)
			}
		}

		return nil
	},
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Show available documents",
	Long:  "List all available documents in the configured space",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetDocuments()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == FormatJSON {
			return outputDocumentsPayload(result, format)
		}
		return outputDocuments(result.Items, format)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(docsCmd)

	infoCmd.Flags().BoolVar(&testPermissions, "test-permissions", false, "Test read/write/delete permissions")
}
