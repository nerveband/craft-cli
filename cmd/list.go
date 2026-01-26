package cmd

import (
	"github.com/spf13/cobra"
)

var (
	listFolderID   string
	listLocation   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents",
	Long: `Retrieve and display documents from Craft.

Filters:
  --folder ID     List documents in a specific folder
  --location LOC  List documents in a special location:
                  unsorted, trash, templates, daily_notes

Examples:
  craft list                          # List all documents
  craft list --format table           # List as table
  craft list --folder abc123          # List documents in folder
  craft list --location unsorted      # List unsorted documents`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetDocumentsFiltered(listFolderID, listLocation)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputDocuments(result.Items, format)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listFolderID, "folder", "", "Filter by folder ID")
	listCmd.Flags().StringVar(&listLocation, "location", "", "Filter by location: unsorted, trash, templates, daily_notes")
}
