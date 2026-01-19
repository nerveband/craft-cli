package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
)

// outputDocuments prints documents in the specified format
func outputDocuments(docs []models.Document, format string) error {
	switch format {
	case "json":
		return outputJSON(docs)
	case "table":
		return outputTable(docs)
	case "markdown":
		return outputMarkdown(docs)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputDocument prints a single document in the specified format
func outputDocument(doc *models.Document, format string) error {
	switch format {
	case "json":
		return outputJSON(doc)
	case "table":
		return outputDocumentTable(doc)
	case "markdown":
		return outputDocumentMarkdown(doc)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputJSON prints data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputTable prints documents as a table
func outputTable(docs []models.Document) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tUPDATED")
	fmt.Fprintln(w, "---\t-----\t-------")

	for _, doc := range docs {
		title := doc.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", doc.ID, title, doc.UpdatedAt.Format("2006-01-02 15:04"))
	}

	return w.Flush()
}

// outputDocumentTable prints a single document as a table
func outputDocumentTable(doc *models.Document) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FIELD\tVALUE")
	fmt.Fprintln(w, "-----\t-----")
	fmt.Fprintf(w, "ID\t%s\n", doc.ID)
	fmt.Fprintf(w, "Title\t%s\n", doc.Title)
	fmt.Fprintf(w, "Space ID\t%s\n", doc.SpaceID)
	
	if doc.ParentID != "" {
		fmt.Fprintf(w, "Parent ID\t%s\n", doc.ParentID)
	}
	
	fmt.Fprintf(w, "Created\t%s\n", doc.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Updated\t%s\n", doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Has Children\t%v\n", doc.HasChildren)
	
	if doc.Content != "" {
		content := doc.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		fmt.Fprintf(w, "Content\t%s\n", content)
	}
	
	if doc.Markdown != "" {
		markdown := strings.ReplaceAll(doc.Markdown, "\n", " ")
		if len(markdown) > 100 {
			markdown = markdown[:97] + "..."
		}
		fmt.Fprintf(w, "Markdown\t%s\n", markdown)
	}

	return w.Flush()
}

// outputMarkdown prints documents as markdown
func outputMarkdown(docs []models.Document) error {
	fmt.Println("# Documents")
	for _, doc := range docs {
		fmt.Printf("## %s\n", doc.Title)
		fmt.Printf("- **ID**: %s\n", doc.ID)
		fmt.Printf("- **Updated**: %s\n", doc.UpdatedAt.Format("2006-01-02 15:04"))
		if doc.Content != "" {
			fmt.Printf("- **Content**: %s\n", doc.Content)
		}
		fmt.Println()
	}
	return nil
}

// outputDocumentMarkdown prints a single document as markdown
func outputDocumentMarkdown(doc *models.Document) error {
	fmt.Printf("# %s\n", doc.Title)
	fmt.Printf("- **ID**: %s\n", doc.ID)
	fmt.Printf("- **Space ID**: %s\n", doc.SpaceID)
	
	if doc.ParentID != "" {
		fmt.Printf("- **Parent ID**: %s\n", doc.ParentID)
	}
	
	fmt.Printf("- **Created**: %s\n", doc.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("- **Updated**: %s\n", doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("- **Has Children**: %v\n", doc.HasChildren)
	
	if doc.Markdown != "" {
		fmt.Println("\n## Content")
		fmt.Println(doc.Markdown)
	} else if doc.Content != "" {
		fmt.Println("\n## Content")
		fmt.Println(doc.Content)
	}
	
	return nil
}
