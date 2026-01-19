package cmd

import (
	"fmt"
	"os"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	createTitle    string
	createFile     string
	createMarkdown string
	createParentID string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new document",
	Long:  "Create a new document in Craft",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		req := &models.CreateDocumentRequest{
			Title:    createTitle,
			ParentID: createParentID,
		}

		// If file is specified, read content from file
		if createFile != "" {
			content, err := os.ReadFile(createFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			req.Markdown = string(content)
		} else if createMarkdown != "" {
			req.Markdown = createMarkdown
		}

		if req.Title == "" {
			return fmt.Errorf("title is required (use --title)")
		}

		doc, err := client.CreateDocument(req)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputDocument(doc, format)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&createTitle, "title", "", "Document title (required)")
	createCmd.Flags().StringVar(&createFile, "file", "", "Read content from file")
	createCmd.Flags().StringVar(&createMarkdown, "markdown", "", "Markdown content")
	createCmd.Flags().StringVar(&createParentID, "parent", "", "Parent document ID")
}
