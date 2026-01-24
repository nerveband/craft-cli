package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

		fmt.Println("Craft CLI Information")
		fmt.Println("=====================")
		if cfg.ActiveProfile != "" {
			fmt.Printf("Active Profile: %s\n", cfg.ActiveProfile)
			if profile, ok := cfg.Profiles[cfg.ActiveProfile]; ok {
				fmt.Printf("API URL: %s\n", profile.URL)
			}
		}
		fmt.Printf("Default Format: %s\n", cfg.DefaultFormat)
		fmt.Println()

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
		return outputDocuments(result.Items, format)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(docsCmd)
}
