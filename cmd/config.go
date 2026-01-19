package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Manage Craft CLI configuration settings",
}

var setAPICmd = &cobra.Command{
	Use:   "set-api [url]",
	Short: "Set the Craft API URL",
	Long:  "Store the Craft API URL in the configuration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		if err := cfgManager.SetAPIURL(url); err != nil {
			return fmt.Errorf("failed to set API URL: %w", err)
		}
		fmt.Printf("API URL set to: %s\n", url)
		return nil
	},
}

var getAPICmd = &cobra.Command{
	Use:   "get-api",
	Short: "Show the configured API URL",
	Long:  "Display the currently configured Craft API URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, err := cfgManager.GetAPIURL()
		if err != nil {
			return fmt.Errorf("failed to get API URL: %w", err)
		}
		if url == "" {
			fmt.Println("No API URL configured")
		} else {
			fmt.Println(url)
		}
		return nil
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all configuration",
	Long:  "Remove the configuration file and reset all settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgManager.Reset(); err != nil {
			return fmt.Errorf("failed to reset config: %w", err)
		}
		fmt.Println("Configuration reset successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setAPICmd)
	configCmd.AddCommand(getAPICmd)
	configCmd.AddCommand(resetCmd)
}
