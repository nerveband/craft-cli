package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [document-id]",
	Short: "Delete a document",
	Long:  "Delete a document from Craft by its ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		if err := client.DeleteDocument(docID); err != nil {
			return err
		}

		fmt.Printf("Document %s deleted successfully\n", docID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
