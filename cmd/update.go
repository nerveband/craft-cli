package cmd

import (
	"fmt"
	"io"
	"os"
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
	updateJSON       string
)

var updateCmd = &cobra.Command{
	Use:   "update <document-id>",
	Short: "Update a document",
	Long: `Update an existing document in Craft.

By default, content updates append new blocks at the end.
Use --mode replace to clear existing content blocks and replace with new content.
Use --section to replace a specific section by heading (requires --mode replace).

Important:
  Replace mode is markdown-based and recreates content blocks. It changes block IDs
  and can wipe block-level styling/metadata such as colors, fonts, alignment,
  decorations, list/task state, media/embed fields, comments, and revert anchors.
  Page-level MCP styling on the root page may survive because the root page block
  is not deleted. For edits to already-styled documents, prefer block-level
  updates such as 'craft blocks update BLOCK_ID --markdown ...', which preserve
  omitted styling fields on that block.
  Section replacement uses a narrower block-boundary delta update when possible:
  only the target section's top-level blocks are deleted/reinserted, preserving
  block IDs and styling outside that section.

Content can be provided via:
  --file <path>     Read content from a file (use - for stdin)
  --markdown <text> Provide content as argument
  --json <payload>  Structured update payload
  <stdin>           Pipe content directly

Examples:
  craft update abc123 --title "New Title"
  craft update abc123 --file content.md
  craft update abc123 --mode replace --file content.md
  craft update abc123 --mode replace --section "Overview" --file overview.md
  craft update abc123 --json '{"title":"New Title","markdown":"# Body","mode":"replace"}'
  echo "# Updated" | craft update abc123
  cat doc.md | craft update abc123 --title "Updated Doc"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}
		if updateJSON != "" {
			payload, err := parseJSONObject(updateJSON)
			if err != nil {
				return err
			}
			applyUpdatePayload(payload)
		}

		if updateStdin {
			if updateFile != "" {
				return fmt.Errorf("--stdin cannot be used with --file")
			}
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			if payload, err := parseJSONObject(string(data)); err == nil && updateJSON == "" && updateTitle == "" && updateMarkdown == "" {
				applyUpdatePayload(payload)
			} else {
				updateMarkdown = string(data)
			}
		}

		mode := updateMode
		if mode == "" {
			mode = "append"
		}
		switch mode {
		case "append", "replace":
		default:
			return fmt.Errorf("invalid --mode %q (expected append or replace)", mode)
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
					delta, err := planSectionDelta(client, docID, updateSection, content)
					if err != nil {
						return err
					}
					planned = delta.Replacement
					fmt.Printf("  Strategy: section delta\n")
					fmt.Printf("  Deletes: %d section blocks\n", len(delta.DeleteIDs))
					if delta.InsertBeforeID != "" {
						fmt.Printf("  Insert before: %s\n", delta.InsertBeforeID)
					} else {
						fmt.Printf("  Insert position: end\n")
					}
				}
				if strings.TrimSpace(planned) != "" {
					chunks := api.SplitMarkdownIntoChunks(planned, chunkBytes)
					fmt.Printf("  Chunk bytes: %d\n", chunkBytes)
					fmt.Printf("  Chunks: %d\n", len(chunks))
				}
				if mode == "replace" && updateSection == "" {
					if risk, err := inspectReplaceStyleRisk(client, docID); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not inspect existing block styling before replace: %v\n", err)
					} else {
						fmt.Printf("  Replace risk: deletes %d content blocks; %d have block-only styling/state that markdown cannot preserve\n", risk.TotalBlocks, risk.StyledBlocks)
						if risk.StyledBlocks > 0 {
							fmt.Printf("  Safer edit: update affected blocks with 'craft blocks update BLOCK_ID --markdown ...' instead of replacing the whole document\n")
						}
					}
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
			if mode == "replace" {
				if err := replaceSectionDelta(client, docID, updateSection, content, chunkBytes); err != nil {
					return err
				}
				return outputCreated(&models.Document{ID: docID, Title: updateTitle}, getOutputFormat())
			}
		}

		if strings.TrimSpace(finalContent) != "" {
			switch mode {
			case "append":
				_, err := client.AppendMarkdown(docID, finalContent, chunkBytes)
				if err != nil {
					return err
				}
			case "replace":
				if !isQuiet() {
					if risk, err := inspectReplaceStyleRisk(client, docID); err == nil && risk.StyledBlocks > 0 {
						fmt.Fprintf(os.Stderr, "Warning: replace will recreate %d content blocks and %d have block-only styling/state that markdown cannot preserve. Prefer block-level updates for styled documents.\n", risk.TotalBlocks, risk.StyledBlocks)
					}
				}
				if err := client.ReplaceDocumentContent(docID, finalContent, chunkBytes); err != nil {
					return err
				}
			}
		}

		return outputCreated(&models.Document{ID: docID, Title: updateTitle}, getOutputFormat())
	},
}

type replaceStyleRisk struct {
	TotalBlocks  int
	StyledBlocks int
}

type sectionDeltaPlan struct {
	Replacement    string
	DeleteIDs      []string
	InsertBeforeID string
}

func replaceSectionDelta(client *api.Client, docID, heading, replacement string, chunkBytes int) error {
	plan, err := planSectionDelta(client, docID, heading, replacement)
	if err != nil {
		return err
	}
	if len(plan.DeleteIDs) == 0 {
		return fmt.Errorf("section heading not found: %s", heading)
	}
	position := map[string]interface{}{"pageId": docID, "position": "end"}
	if plan.InsertBeforeID != "" {
		position = map[string]interface{}{"siblingId": plan.InsertBeforeID, "position": "before"}
	}
	if err := client.DeleteBlocks(plan.DeleteIDs); err != nil {
		return err
	}
	_, err = client.AddMarkdownAtPosition(plan.Replacement, position, chunkBytes)
	return err
}

func planSectionDelta(client *api.Client, docID, heading, replacement string) (sectionDeltaPlan, error) {
	blocks, err := client.GetDocumentBlocks(docID)
	if err != nil {
		return sectionDeltaPlan{}, err
	}
	return buildSectionDeltaPlan(blocks.Content, heading, replacement)
}

func buildSectionDeltaPlan(blocks []models.Block, heading, replacement string) (sectionDeltaPlan, error) {
	target := normalizeHeadingText(heading)
	if target == "" {
		return sectionDeltaPlan{}, fmt.Errorf("section heading is required")
	}
	start := -1
	level := 0
	originalHeading := ""
	for i := range blocks {
		blockLevel, text, ok := blockHeading(blocks[i].Markdown)
		if !ok {
			continue
		}
		if normalizeHeadingText(text) == target {
			start = i
			level = blockLevel
			originalHeading = strings.TrimSpace(blocks[i].Markdown)
			break
		}
	}
	if start == -1 {
		return sectionDeltaPlan{}, fmt.Errorf("section heading not found: %s", heading)
	}
	end := len(blocks)
	for i := start + 1; i < len(blocks); i++ {
		blockLevel, _, ok := blockHeading(blocks[i].Markdown)
		if ok && blockLevel <= level {
			end = i
			break
		}
	}

	repl := strings.TrimSpace(strings.ReplaceAll(replacement, "\r\n", "\n"))
	repl = strings.ReplaceAll(repl, "\r", "\n")
	if repl == "" {
		return sectionDeltaPlan{}, fmt.Errorf("replacement content is required")
	}
	firstLine := strings.SplitN(repl, "\n", 2)[0]
	if _, _, ok := blockHeading(strings.TrimSpace(firstLine)); !ok {
		repl = originalHeading + "\n\n" + repl
	}

	var ids []string
	for i := start; i < end; i++ {
		collectBlockIDsFromModel(&blocks[i], &ids)
	}
	plan := sectionDeltaPlan{Replacement: repl, DeleteIDs: ids}
	if end < len(blocks) {
		plan.InsertBeforeID = blocks[end].ID
	}
	return plan, nil
}

func blockHeading(markdown string) (int, string, bool) {
	line := strings.TrimSpace(strings.SplitN(markdown, "\n", 2)[0])
	m := headingRe.FindStringSubmatch(line)
	if m == nil {
		return 0, "", false
	}
	return len(m[1]), m[2], true
}

func collectBlockIDsFromModel(block *models.Block, ids *[]string) {
	if block.ID != "" {
		*ids = append(*ids, block.ID)
	}
	for _, child := range block.Content {
		collectBlockIDsFromModel(&child, ids)
	}
}

func inspectReplaceStyleRisk(client *api.Client, docID string) (replaceStyleRisk, error) {
	blocks, err := client.GetDocumentBlocks(docID)
	if err != nil {
		return replaceStyleRisk{}, err
	}
	var risk replaceStyleRisk
	for _, block := range blocks.Content {
		countReplaceStyleRisk(&block, &risk)
	}
	return risk, nil
}

func countReplaceStyleRisk(block *models.Block, risk *replaceStyleRisk) {
	risk.TotalBlocks++
	if blockHasMarkdownUnsafeState(block) {
		risk.StyledBlocks++
	}
	for _, child := range block.Content {
		countReplaceStyleRisk(&child, risk)
	}
}

func blockHasMarkdownUnsafeState(block *models.Block) bool {
	if block.Color != "" ||
		block.CardLayout != "" ||
		block.IndentationLevel != 0 ||
		block.Font != "" ||
		block.TextAlignment != "" ||
		block.TaskInfo != nil ||
		block.URL != "" ||
		block.AltText != "" ||
		block.FileName != "" ||
		block.Title != "" ||
		block.Description != "" ||
		block.Layout != "" ||
		len(block.Rows) > 0 ||
		block.Metadata != nil {
		return true
	}
	return false
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New document title")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Read content from file (use - for stdin)")
	updateCmd.Flags().StringVar(&updateMarkdown, "markdown", "", "Markdown content")
	updateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read content from stdin")
	updateCmd.Flags().StringVar(&updateJSON, "json", "", "Structured update payload JSON")
	updateCmd.Flags().StringVar(&updateMode, "mode", "append", "Update mode (append, replace)")
	updateCmd.Flags().StringVar(&updateSection, "section", "", "Replace a section by heading (requires --mode replace)")
	updateCmd.Flags().IntVar(&updateChunkBytes, "chunk-bytes", 30000, "Max bytes per insert chunk (helps avoid API payload limits)")
}

func applyUpdatePayload(payload map[string]interface{}) {
	if title, ok := payload["title"].(string); ok {
		updateTitle = title
	}
	if markdown, ok := payload["markdown"].(string); ok {
		updateMarkdown = markdown
	}
	if content, ok := payload["content"].(string); ok && updateMarkdown == "" {
		updateMarkdown = content
	}
	if mode, ok := payload["mode"].(string); ok {
		updateMode = mode
	}
	if section, ok := payload["section"].(string); ok {
		updateSection = section
	}
	if chunkBytes, ok := payload["chunkBytes"].(float64); ok {
		updateChunkBytes = int(chunkBytes)
	}
}
