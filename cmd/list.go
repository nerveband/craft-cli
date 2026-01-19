package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents",
	Long:  "Retrieve and display all documents from Craft",
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
	rootCmd.AddCommand(listCmd)
}
