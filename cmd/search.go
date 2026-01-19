package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for documents",
	Long:  "Search for documents matching the specified query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		query := args[0]
		result, err := client.SearchDocuments(query)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputDocuments(result.Items, format)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
