package cmd

import (
	"fmt"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	updateTitle    string
	updateFile     string
	updateMarkdown string
)

var updateCmd = &cobra.Command{
	Use:   "update <document-id>",
	Short: "Update a document",
	Long: `Update an existing document in Craft.

Content can be provided via:
  --file <path>     Read content from a file (use - for stdin)
  --markdown <text> Provide content as argument
  <stdin>           Pipe content directly

Examples:
  craft update abc123 --title "New Title"
  craft update abc123 --file content.md
  echo "# Updated" | craft update abc123
  cat doc.md | craft update abc123 --title "Updated Doc"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		req := &models.UpdateDocumentRequest{
			Title: updateTitle,
		}

		// Read content from various sources
		content, err := readContent(updateFile, updateMarkdown)
		if err != nil {
			return err
		}
		req.Markdown = content

		if req.Title == "" && req.Markdown == "" {
			return fmt.Errorf("at least one of --title, --file, or --markdown is required")
		}

		// Dry run mode
		if isDryRun() {
			fmt.Printf("Would update document %s:\n", docID)
			if req.Title != "" {
				fmt.Printf("  New title: %s\n", req.Title)
			}
			if req.Markdown != "" {
				preview := req.Markdown
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				fmt.Printf("  New content: %s\n", preview)
			}
			return nil
		}

		doc, err := client.UpdateDocument(docID, req)
		if err != nil {
			return err
		}

		return outputCreated(doc, getOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New document title")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Read content from file (use - for stdin)")
	updateCmd.Flags().StringVar(&updateMarkdown, "markdown", "", "Markdown content")
}
