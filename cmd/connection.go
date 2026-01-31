package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Show space connection info",
	Long: `Display space metadata including timezone, current time, and deep link URL templates.

Examples:
  craft connection
  craft connection --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		info, err := client.GetConnection()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(info)
		}

		fmt.Println("Space Connection")
		fmt.Println("================")
		fmt.Printf("Space ID:      %s\n", info.Space.ID)
		fmt.Printf("Timezone:      %s\n", info.Space.Timezone)
		fmt.Printf("Local Time:    %s\n", info.Space.Time)
		fmt.Printf("Date:          %s\n", info.Space.FriendlyDate)
		fmt.Printf("UTC Time:      %s\n", info.UTC.Time)
		if info.URLTemplates.App != "" {
			fmt.Printf("Deep Link:     %s\n", info.URLTemplates.App)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectionCmd)
}
