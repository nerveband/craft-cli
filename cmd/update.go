package cmd

import (
	"fmt"
	"strings"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	updateTitle      string
	updateFile       string
	updateMarkdown   string
	updateStdin      bool
	updateMode       string
	updateSection    string
	updateChunkBytes int
)

var updateCmd = &cobra.Command{
	Use:   "update <document-id>",
	Short: "Update a document",
	Long: `Update an existing document in Craft.

By default, content updates append new blocks at the end.
Use --mode replace to clear existing content blocks and replace with new content.
Use --section to replace a specific section by heading (requires --mode replace).

Content can be provided via:
  --file <path>     Read content from a file (use - for stdin)
  --markdown <text> Provide content as argument
  <stdin>           Pipe content directly

Examples:
  craft update abc123 --title "New Title"
  craft update abc123 --file content.md
  craft update abc123 --mode replace --file content.md
  craft update abc123 --mode replace --section "Overview" --file overview.md
  echo "# Updated" | craft update abc123
  cat doc.md | craft update abc123 --title "Updated Doc"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]

		mode := updateMode
		if mode == "" {
			mode = "append"
		}
		switch mode {
		case "append", "replace":
		default:
			return fmt.Errorf("invalid --mode %q (expected append or replace)", mode)
		}

		if updateStdin {
			if updateFile != "" {
				return fmt.Errorf("--stdin cannot be used with --file")
			}
			updateFile = "-"
		}

		// Read content from various sources
		content, err := readContent(updateFile, updateMarkdown)
		if err != nil {
			return err
		}

		if updateSection != "" && mode != "replace" {
			return fmt.Errorf("--section requires --mode replace")
		}

		if updateTitle == "" && strings.TrimSpace(content) == "" && updateSection == "" {
			return fmt.Errorf("at least one of --title, --file, --markdown, or --section is required")
		}

		// Dry run mode
		if isDryRun() {
			fmt.Printf("Would update document %s:\n", docID)
			if updateTitle != "" {
				fmt.Printf("  New title: %s\n", updateTitle)
			}
			if strings.TrimSpace(content) != "" || updateSection != "" {
				fmt.Printf("  Mode: %s\n", mode)
				if updateSection != "" {
					fmt.Printf("  Section: %s\n", updateSection)
				}
				chunkBytes := updateChunkBytes
				if chunkBytes <= 0 {
					chunkBytes = 30000
				}
				planned := content
				if updateSection != "" {
					existing, err := client.GetDocumentContentMarkdown(docID)
					if err != nil {
						return err
					}
					updated, err := replaceSectionByHeading(existing, updateSection, content)
					if err != nil {
						return err
					}
					planned = updated
				}
				if strings.TrimSpace(planned) != "" {
					chunks := api.SplitMarkdownIntoChunks(planned, chunkBytes)
					fmt.Printf("  Chunk bytes: %d\n", chunkBytes)
					fmt.Printf("  Chunks: %d\n", len(chunks))
				}

				preview := content
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				if strings.TrimSpace(preview) != "" {
					fmt.Printf("  Content preview: %s\n", preview)
				}
			}
			return nil
		}

		chunkBytes := updateChunkBytes
		if chunkBytes <= 0 {
			chunkBytes = 30000
		}

		// Title update (root page block)
		if updateTitle != "" {
			if err := client.UpdateBlockMarkdown(docID, updateTitle); err != nil {
				return err
			}
		}

		finalContent := content
		if updateSection != "" {
			existing, err := client.GetDocumentContentMarkdown(docID)
			if err != nil {
				return err
			}
			updated, err := replaceSectionByHeading(existing, updateSection, content)
			if err != nil {
				return err
			}
			finalContent = updated
		}

		if strings.TrimSpace(finalContent) != "" {
			switch mode {
			case "append":
				_, err := client.AppendMarkdown(docID, finalContent, chunkBytes)
				if err != nil {
					return err
				}
			case "replace":
				if err := client.ReplaceDocumentContent(docID, finalContent, chunkBytes); err != nil {
					return err
				}
			}
		}

		return outputCreated(&models.Document{ID: docID, Title: updateTitle}, getOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New document title")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Read content from file (use - for stdin)")
	updateCmd.Flags().StringVar(&updateMarkdown, "markdown", "", "Markdown content")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin")
	updateCmd.Flags().StringVar(&updateMode, "mode", "append", "Update mode (append, replace)")
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Replace a section by heading (requires --mode replace)")
	updateCmd.Flags().IntVar(&updateChunkBytes, "chunk-bytes", 30000, "Max bytes per insert chunk (helps avoid API payload limits)")
}
