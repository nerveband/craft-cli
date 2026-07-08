package cmd

import (
	"fmt"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	clearJSON  string
	clearStdin bool
)

var clearCmd = &cobra.Command{
	Use:   "clear <document-id>",
	Short: "Delete all content blocks in a document",
	Long: `Clear a document by deleting all of its content blocks.

This does NOT delete the document itself.
Use craft delete to move the document to trash.

WARNING: This operation is destructive. Use --dry-run to preview first.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if clearJSON != "" || clearStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if clearJSON != "" || clearStdin {
			payload, err := readRawPayload(clearJSON, clearStdin)
			if err != nil {
				return err
			}
			docID := payloadString(payload, "id", "documentId")
			if docID == "" {
				return fmt.Errorf("raw clear payload must include \"id\" or \"documentId\"")
			}
			args = []string{docID}
		}
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
			return dryRunOutput("clear", map[string]interface{}{
				"id": docID, "block_count": count, "destructive": true,
			})
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
	clearCmd.Flags().StringVar(&clearJSON, "json", "", "Raw clear payload JSON")
	clearCmd.Flags().BoolVar(&clearStdin, "stdin", false, "Read raw clear payload from stdin")
}
