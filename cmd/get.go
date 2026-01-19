package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFile string
)

var getCmd = &cobra.Command{
	Use:   "get [document-id]",
	Short: "Get a document by ID",
	Long:  "Retrieve and display a specific document from Craft by its ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		doc, err := client.GetDocument(docID)
		if err != nil {
			return err
		}

		// If output file is specified, write markdown to file
		if outputFile != "" {
			content := doc.Markdown
			if content == "" {
				content = doc.Content
			}
			
			if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
			
			fmt.Printf("Document saved to: %s\n", outputFile)
			return nil
		}

		format := getOutputFormat()
		return outputDocument(doc, format)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write document content to file")
}
