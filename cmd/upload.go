package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	uploadPageID    string
	uploadDate      string
	uploadSiblingID string
	uploadPosition  string
	uploadStdin     bool
	uploadJSON      string
)

var uploadCmd = &cobra.Command{
	Use:   "upload [file-path]",
	Short: "Upload a file to a Craft document",
	Long: `Upload a file (image, PDF, etc.) to a Craft document.

The file is uploaded as a binary asset and inserted as a block. You must
specify where to place the block using --page, --date, or --sibling.

Content can be provided via:
  <file-path>   Path to the file to upload
  --file        Use the positional file path; @file:// and @data:// encodings are planned for structured payloads
  --json        Structured payload with filePath or dataBase64 and placement metadata
  --stdin       Read file data from stdin

Examples:
  craft upload photo.png --page PAGE_ID
  craft upload report.pdf --page PAGE_ID --position start
  craft upload diagram.svg --date 2025-01-15
  craft upload image.jpg --sibling BLOCK_ID --position before
  craft upload --json '{"filePath":"photo.png","pageId":"PAGE_ID","position":"end"}' --dry-run
  cat photo.png | craft upload --stdin --page PAGE_ID

  # Chain-friendly (returns just the block ID)
  BLOCK_ID=$(craft upload -q photo.png --page PAGE_ID)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if uploadJSON != "" {
			if uploadStdin || len(args) > 0 {
				return fmt.Errorf("--json cannot be combined with file path or --stdin")
			}
			payload, err := parseJSONObject(uploadJSON)
			if err != nil {
				return err
			}
			return runUploadRaw(payload)
		}

		// Read file data
		var fileData []byte
		var fileName string

		if uploadStdin {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify both a file path and --stdin")
			}
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			fileData = data
			fileName = "(stdin)"
		} else {
			if len(args) == 0 {
				return fmt.Errorf("file path is required (or use --stdin)")
			}
			filePath := args[0]
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			fileData = data
			fileName = filepath.Base(filePath)
		}

		// Validate placement: at least one of --page, --date, --sibling must be provided
		if uploadPageID == "" && uploadDate == "" && uploadSiblingID == "" {
			return fmt.Errorf("at least one of --page, --date, or --sibling is required")
		}

		// Dry run mode
		if isDryRun() {
			return dryRunOutput("upload file", map[string]interface{}{
				"file":     fileName,
				"size":     len(fileData),
				"page":     uploadPageID,
				"date":     uploadDate,
				"sibling":  uploadSiblingID,
				"position": uploadPosition,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.UploadFile(fileData, uploadPageID, uploadDate, uploadSiblingID, uploadPosition)
		if err != nil {
			return err
		}

		// Output based on mode
		if isQuiet() {
			fmt.Println(result.BlockID)
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(result)
		}

		// Default text output
		fmt.Printf("Block ID:  %s\n", result.BlockID)
		fmt.Printf("Asset URL: %s\n", result.AssetURL)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVar(&uploadPageID, "page", "", "Target page ID for placement")
	uploadCmd.Flags().StringVar(&uploadDate, "date", "", "Target daily note date (YYYY-MM-DD)")
	uploadCmd.Flags().StringVar(&uploadSiblingID, "sibling", "", "Sibling block ID for relative placement")
	uploadCmd.Flags().StringVarP(&uploadPosition, "position", "p", "end", "Position: start, end, before, after")
	uploadCmd.Flags().BoolVar(&uploadStdin, "stdin", false, "Read file data from stdin")
	uploadCmd.Flags().StringVar(&uploadJSON, "json", "", "Structured upload payload JSON")
}

func runUploadRaw(payload map[string]interface{}) error {
	pageID := payloadString(payload, "pageId", "pageID", "page")
	date := payloadString(payload, "date")
	siblingID := payloadString(payload, "siblingId", "siblingID", "sibling")
	position := payloadString(payload, "position")
	if position == "" {
		position = "end"
	}
	if pageID == "" && date == "" && siblingID == "" {
		return fmt.Errorf("raw upload payload must include pageId, date, or siblingId")
	}

	var fileData []byte
	fileName := payloadString(payload, "fileName", "filename", "name")
	filePath := payloadString(payload, "filePath", "file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		fileData = data
		if fileName == "" {
			fileName = filepath.Base(filePath)
		}
	} else if encoded := payloadString(payload, "dataBase64", "base64"); encoded != "" {
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("invalid base64 upload data: %w", err)
		}
		fileData = data
		if fileName == "" {
			fileName = "(base64)"
		}
	} else {
		return fmt.Errorf("raw upload payload must include filePath or dataBase64")
	}

	if isDryRun() {
		return dryRunOutput("upload file", map[string]interface{}{
			"file":     fileName,
			"size":     len(fileData),
			"page":     pageID,
			"date":     date,
			"sibling":  siblingID,
			"position": position,
		})
	}

	client, err := getAPIClient()
	if err != nil {
		return err
	}
	result, err := client.UploadFile(fileData, pageID, date, siblingID, position)
	if err != nil {
		return err
	}
	if isQuiet() {
		fmt.Println(result.BlockID)
		return nil
	}
	return outputJSON(result)
}
