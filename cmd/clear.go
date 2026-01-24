package cmd

import (
	"fmt"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear <document-id>",
	Short: "Delete all content blocks in a document",
	Long: `Clear a document by deleting all of its content blocks.

This does NOT delete the document itself.
Use craft delete to move the document to trash.

WARNING: This operation is destructive. Use --dry-run to preview first.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if isDryRun() {
			blocks, err := client.GetDocumentBlocks(docID)
			if err != nil {
				return err
			}
			var countBlocks func(bs []models.Block) int
			countBlocks = func(bs []models.Block) int {
				n := 0
				for _, b := range bs {
					n++
					n += countBlocks(b.Content)
				}
				return n
			}
			count := countBlocks(blocks.Content)
			fmt.Printf("Would clear document %s (%d blocks)\n", docID, count)
			return nil
		}

		deleted, err := client.ClearDocumentContent(docID)
		if err != nil {
			return err
		}
		outputCleared(docID, deleted)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
