package cmd

import (
	"fmt"
	"os"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	updateTitle    string
	updateFile     string
	updateMarkdown string
)

var updateCmd = &cobra.Command{
	Use:   "update [document-id]",
	Short: "Update a document",
	Long:  "Update an existing document in Craft",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		req := &models.UpdateDocumentRequest{
			Title: updateTitle,
		}

		// If file is specified, read content from file
		if updateFile != "" {
			content, err := os.ReadFile(updateFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			req.Markdown = string(content)
		} else if updateMarkdown != "" {
			req.Markdown = updateMarkdown
		}

		if req.Title == "" && req.Markdown == "" {
			return fmt.Errorf("at least one of --title, --file, or --markdown is required")
		}

		doc, err := client.UpdateDocument(docID, req)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputDocument(doc, format)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New document title")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Read content from file")
	updateCmd.Flags().StringVar(&updateMarkdown, "markdown", "", "Markdown content")
}
