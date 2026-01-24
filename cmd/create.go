package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	createTitle    string
	createFile     string
	createMarkdown string
	createParentID string
	batchCreate    bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new document",
	Long: `Create a new document in Craft.

Content can be provided via:
  --file <path>     Read content from a file (use - for stdin)
  --markdown <text> Provide content as argument
  <stdin>           Pipe content directly
  --batch           Read JSON array of documents from stdin

Examples:
  craft create --title "Note" --file content.md
  craft create --title "Note" --file -              # Read from stdin
  echo "# Hello" | craft create --title "Note"      # Pipe content
  cat doc.md | craft create --title "Imported"

  # Batch create
  echo '[{"title":"Doc1"},{"title":"Doc2"}]' | craft create --batch

  # Chain-friendly (returns just the ID)
  ID=$(craft create -q --title "Note")
  craft update $ID --file content.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle batch mode
		if batchCreate {
			return runBatchCreate()
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		req := &models.CreateDocumentRequest{
			Title:    createTitle,
			ParentID: createParentID,
		}

		// Read content from various sources
		content, err := readContent(createFile, createMarkdown)
		if err != nil {
			return err
		}
		req.Markdown = content

		if req.Title == "" {
			return fmt.Errorf("title is required (use --title)")
		}

		// Dry run mode
		if isDryRun() {
			fmt.Println("Would create document:")
			fmt.Printf("  Title: %s\n", req.Title)
			if req.ParentID != "" {
				fmt.Printf("  Parent: %s\n", req.ParentID)
			}
			if req.Markdown != "" {
				preview := req.Markdown
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				fmt.Printf("  Content: %s\n", preview)
			}
			return nil
		}

		doc, err := client.CreateDocument(req)
		if err != nil {
			return err
		}

		return outputCreated(doc, getOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&createTitle, "title", "", "Document title (required)")
	createCmd.Flags().StringVar(&createFile, "file", "", "Read content from file (use - for stdin)")
	createCmd.Flags().StringVar(&createMarkdown, "markdown", "", "Markdown content")
	createCmd.Flags().StringVar(&createParentID, "parent", "", "Parent document ID")
	createCmd.Flags().BoolVar(&batchCreate, "batch", false, "Batch create from JSON array on stdin")
}

// readContent reads content from file, argument, or stdin
func readContent(filePath, markdown string) (string, error) {
	// Explicit file path provided
	if filePath != "" {
		if filePath == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read stdin: %w", err)
			}
			return string(data), nil
		}
		// Read from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	}

	// Markdown argument provided
	if markdown != "" {
		return markdown, nil
	}

	// Check if stdin has data (piped input)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}

	// No content provided
	return "", nil
}

// runBatchCreate creates multiple documents from JSON stdin
func runBatchCreate() error {
	client, err := getAPIClient()
	if err != nil {
		return err
	}

	// Read JSON from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var requests []models.CreateDocumentRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if isDryRun() {
		fmt.Printf("Would create %d documents:\n", len(requests))
		for i, req := range requests {
			fmt.Printf("  %d. %s\n", i+1, req.Title)
		}
		return nil
	}

	var results []models.Document
	for _, req := range requests {
		doc, err := client.CreateDocument(&req)
		if err != nil {
			printStatus("Error creating '%s': %v\n", req.Title, err)
			continue
		}
		results = append(results, *doc)
		printStatus("Created: %s (%s)\n", doc.Title, doc.ID)
	}

	// Output results
	if isQuiet() {
		for _, doc := range results {
			fmt.Println(doc.ID)
		}
		return nil
	}

	return outputDocuments(results, getOutputFormat())
}
