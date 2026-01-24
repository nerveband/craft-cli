package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <document-id>",
	Short: "Move a document to trash",
	Long: `Soft-delete a document by moving it to Craft trash.

This uses DELETE /documents (you can restore via documents/move).

Use --dry-run to preview what would be deleted without making changes.

Examples:
  craft delete abc123
  craft delete abc123 --dry-run    # Preview without deleting
  craft delete abc123 -q           # Silent delete`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]

		// Dry run mode - try to get doc info first
		if isDryRun() {
			client, err := getAPIClient()
			if err != nil {
				return err
			}

			doc, err := client.GetDocument(docID)
			if err != nil {
				return fmt.Errorf("document not found: %s", docID)
			}

			fmt.Printf("Would move document to trash:\n")
			fmt.Printf("  ID: %s\n", doc.ID)
			fmt.Printf("  Title: %s\n", doc.Title)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteDocument(docID); err != nil {
			return err
		}

		outputDeleted(docID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
