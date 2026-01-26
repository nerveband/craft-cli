package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	outputFile string
)

var getCmd = &cobra.Command{
	Use:   "get [document-id]",
	Short: "Get a document by ID",
	Long: `Retrieve and display a specific document from Craft by its ID.

Output Formats:
  json        Full document metadata as JSON (default)
  table       Tabular view of document fields
  markdown    Plain markdown content
  structured  Full block tree with all metadata (for LLMs)
  craft       MCP-style markdown with XML tags (matches Craft MCP server)
  rich        Terminal output with ANSI colors and Unicode

Examples:
  craft get abc123                        # Default JSON output
  craft get abc123 --format structured    # Full block tree for AI processing
  craft get abc123 --format craft         # MCP-compatible format
  craft get abc123 --format rich          # Pretty terminal output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		format := getOutputFormat()

		// For structured/craft/rich formats, get full block response
		if format == FormatStructured || format == FormatCraft || format == FormatRich {
			blocksResp, err := client.GetDocumentBlocks(docID)
			if err != nil {
				return err
			}

			// If output file is specified, write to file
			if outputFile != "" {
				return writeBlocksToFile(&blocksResp, format)
			}

			switch format {
			case FormatStructured:
				return outputBlocksStructured(&blocksResp)
			case FormatCraft:
				return outputBlocksCraft(&blocksResp)
			case FormatRich:
				return outputBlocksRich(&blocksResp)
			}
		}

		// Legacy formats use Document model
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

			if !isQuiet() {
				fmt.Fprintf(os.Stderr, "Document saved to: %s\n", outputFile)
			}
			return nil
		}

		return outputDocument(doc, format)
	},
}

// writeBlocksToFile writes block output to a file
func writeBlocksToFile(blocksResp *models.BlocksResponse, format string) error {
	var content string

	switch format {
	case FormatStructured:
		data, err := json.MarshalIndent(blocksResp, "", "  ")
		if err != nil {
			return err
		}
		content = string(data)
	case FormatCraft:
		var sb strings.Builder
		renderBlockCraft(&sb, blockFromResponse(blocksResp), 0)
		content = sb.String()
	case FormatRich:
		// Strip ANSI codes for file output
		var sb strings.Builder
		renderBlockCraft(&sb, blockFromResponse(blocksResp), 0) // Use craft format for files
		content = sb.String()
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	if !isQuiet() {
		fmt.Fprintf(os.Stderr, "Document saved to: %s\n", outputFile)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write document content to file")
}
