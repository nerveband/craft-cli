package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ashrafali/craft-cli/internal/models"
)

// ANSI color codes for rich output
const (
	colorReset     = "\033[0m"
	colorBold      = "\033[1m"
	colorDim       = "\033[2m"
	colorItalic    = "\033[3m"
	colorUnderline = "\033[4m"

	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorMagenta   = "\033[35m"
	colorCyan      = "\033[36m"
	colorWhite     = "\033[37m"

	colorBgBlue    = "\033[44m"
	colorBgYellow  = "\033[43m"
	colorBgCyan    = "\033[46m"
)

// Unicode box drawing characters
const (
	boxTopLeft     = "╔"
	boxTopRight    = "╗"
	boxBottomLeft  = "╚"
	boxBottomRight = "╝"
	boxHorizontal  = "═"
	boxVertical    = "║"

	cardTopLeft     = "┌"
	cardTopRight    = "┐"
	cardBottomLeft  = "└"
	cardBottomRight = "┘"
	cardHorizontal  = "─"
	cardVertical    = "│"

	taskTodo      = "☐"
	taskDone      = "✅"
	taskCanceled  = "⊘"
	toggleClosed  = "▶"
	toggleOpen    = "▼"
	bullet        = "•"
)

// outputBlocksStructured outputs the full block tree as JSON (for LLMs)
func outputBlocksStructured(resp *models.BlocksResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resp)
}

// outputBlocksCraft outputs MCP-style markdown with XML tags
func outputBlocksCraft(resp *models.BlocksResponse) error {
	var sb strings.Builder
	renderBlockCraft(&sb, blockFromResponse(resp), 0)
	fmt.Print(sb.String())
	return nil
}

// outputBlocksRich outputs with ANSI colors and Unicode for terminal
func outputBlocksRich(resp *models.BlocksResponse) error {
	var sb strings.Builder
	renderBlockRich(&sb, blockFromResponse(resp), 0)
	fmt.Print(sb.String())
	return nil
}

// blockFromResponse converts BlocksResponse to Block for unified rendering
func blockFromResponse(resp *models.BlocksResponse) *models.Block {
	return &models.Block{
		ID:         resp.ID,
		Type:       resp.Type,
		TextStyle:  resp.TextStyle,
		Markdown:   resp.Markdown,
		Content:    resp.Content,
		CardLayout: resp.CardLayout,
		Metadata:   resp.Metadata,
	}
}

// renderBlockCraft renders a block in MCP-style XML markdown format
func renderBlockCraft(sb *strings.Builder, block *models.Block, depth int) {
	indent := strings.Repeat("  ", depth)

	switch block.Type {
	case "page":
		// Build attributes string
		attrs := buildCraftAttrs(block)
		sb.WriteString(fmt.Sprintf("%s<page id=\"%s\"%s>\n", indent, block.ID, attrs))
		sb.WriteString(fmt.Sprintf("%s  <pageTitle>%s</pageTitle>\n", indent, block.Markdown))
		if len(block.Content) > 0 {
			sb.WriteString(fmt.Sprintf("%s  <content>\n", indent))
			for _, child := range block.Content {
				renderBlockCraft(sb, &child, depth+2)
			}
			sb.WriteString(fmt.Sprintf("%s  </content>\n", indent))
		}
		sb.WriteString(fmt.Sprintf("%s</page>\n", indent))

	case "text":
		md := block.Markdown
		// Wrap in decorations if present
		if sliceContains(block.Decorations, "callout") && sliceContains(block.Decorations, "quote") {
			md = fmt.Sprintf("<callout>%s</callout>", md)
		} else if sliceContains(block.Decorations, "callout") {
			md = fmt.Sprintf("<callout>%s</callout>", md)
		} else if sliceContains(block.Decorations, "quote") {
			md = fmt.Sprintf("> %s", md)
		}

		// Add styling attributes as comments if significant
		if block.Font != "" || block.Color != "" || block.TextAlignment != "" {
			attrs := buildCraftAttrs(block)
			sb.WriteString(fmt.Sprintf("%s<text%s>%s</text>\n", indent, attrs, md))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, md))
		}

		// Render children
		for _, child := range block.Content {
			renderBlockCraft(sb, &child, depth)
		}

	case "code":
		sb.WriteString(fmt.Sprintf("%s%s\n", indent, block.Markdown))

	case "table":
		sb.WriteString(fmt.Sprintf("%s%s\n", indent, block.Markdown))

	case "line":
		style := block.LineStyle
		if style == "" {
			style = "regular"
		}
		sb.WriteString(fmt.Sprintf("%s<line style=\"%s\" />\n", indent, style))

	case "image":
		if block.AltText != "" {
			sb.WriteString(fmt.Sprintf("%s![%s](%s)\n", indent, block.AltText, block.URL))
		} else {
			sb.WriteString(fmt.Sprintf("%s![](%s)\n", indent, block.URL))
		}

	case "file":
		sb.WriteString(fmt.Sprintf("%s<file name=\"%s\" url=\"%s\" />\n", indent, block.FileName, block.URL))

	case "richUrl":
		attrs := ""
		if block.Layout != "" {
			attrs = fmt.Sprintf(" layout=\"%s\"", block.Layout)
		}
		sb.WriteString(fmt.Sprintf("%s<richUrl title=\"%s\" url=\"%s\"%s>%s</richUrl>\n",
			indent, block.Title, block.URL, attrs, block.Description))

	default:
		// Default: just output markdown
		if block.Markdown != "" {
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, block.Markdown))
		}
		for _, child := range block.Content {
			renderBlockCraft(sb, &child, depth)
		}
	}
}

// buildCraftAttrs builds XML attribute string for craft format
func buildCraftAttrs(block *models.Block) string {
	var attrs []string
	if block.TextStyle != "" && block.TextStyle != "page" {
		attrs = append(attrs, fmt.Sprintf("textStyle=\"%s\"", block.TextStyle))
	}
	if block.CardLayout != "" {
		attrs = append(attrs, fmt.Sprintf("cardLayout=\"%s\"", block.CardLayout))
	}
	if block.Font != "" {
		attrs = append(attrs, fmt.Sprintf("font=\"%s\"", block.Font))
	}
	if block.Color != "" {
		attrs = append(attrs, fmt.Sprintf("color=\"%s\"", block.Color))
	}
	if block.ListStyle != "" {
		attrs = append(attrs, fmt.Sprintf("listStyle=\"%s\"", block.ListStyle))
	}
	if block.IndentationLevel > 0 {
		attrs = append(attrs, fmt.Sprintf("indent=\"%d\"", block.IndentationLevel))
	}
	if len(attrs) > 0 {
		return " " + strings.Join(attrs, " ")
	}
	return ""
}

// renderBlockRich renders a block with ANSI colors and Unicode
func renderBlockRich(sb *strings.Builder, block *models.Block, depth int) {
	indent := strings.Repeat("  ", depth)

	switch block.Type {
	case "page":
		if depth == 0 {
			// Root page - fancy header
			title := block.Markdown
			width := len(title) + 4
			if width < 60 {
				width = 60
			}

			sb.WriteString(colorBold + colorCyan)
			sb.WriteString(boxTopLeft + strings.Repeat(boxHorizontal, width) + boxTopRight + "\n")
			sb.WriteString(boxVertical + "  " + title + strings.Repeat(" ", width-len(title)-2) + boxVertical + "\n")
			sb.WriteString(boxBottomLeft + strings.Repeat(boxHorizontal, width) + boxBottomRight + "\n")
			sb.WriteString(colorReset + "\n")
		} else {
			// Sub-page/card
			renderCard(sb, block, indent)
		}

		// Render content
		for _, child := range block.Content {
			renderBlockRich(sb, &child, depth)
		}

	case "text":
		renderTextRich(sb, block, indent)

	case "code":
		sb.WriteString(colorDim)
		sb.WriteString(block.Markdown)
		sb.WriteString(colorReset + "\n")

	case "table":
		renderTableRich(sb, block, indent)

	case "line":
		style := block.LineStyle
		char := "─"
		if style == "strong" {
			char = "━"
		} else if style == "extraLight" {
			char = "┄"
		}
		sb.WriteString(colorDim + indent + strings.Repeat(char, 50) + colorReset + "\n")

	case "image":
		sb.WriteString(fmt.Sprintf("%s🖼  %s%s%s\n", indent, colorUnderline, block.URL, colorReset))
		if block.AltText != "" {
			sb.WriteString(fmt.Sprintf("%s   %s(%s)%s\n", indent, colorDim, block.AltText, colorReset))
		}

	case "file":
		sb.WriteString(fmt.Sprintf("%s📎 %s%s%s\n", indent, colorCyan, block.FileName, colorReset))

	case "richUrl":
		sb.WriteString(fmt.Sprintf("%s🔗 %s%s%s\n", indent, colorBold, block.Title, colorReset))
		if block.Description != "" {
			sb.WriteString(fmt.Sprintf("%s   %s%s%s\n", indent, colorDim, block.Description, colorReset))
		}

	default:
		if block.Markdown != "" {
			sb.WriteString(indent + block.Markdown + "\n")
		}
		for _, child := range block.Content {
			renderBlockRich(sb, &child, depth)
		}
	}
}

// renderCard renders a sub-page/card with box drawing
func renderCard(sb *strings.Builder, block *models.Block, indent string) {
	title := block.Markdown
	icon := "📋"
	if block.TextStyle == "card" {
		icon = "📑"
	}

	// Determine width based on card layout
	width := 60
	switch block.CardLayout {
	case "small":
		width = 40
	case "large":
		width = 80
	}

	titleLine := fmt.Sprintf("%s %s ", icon, title)
	padding := width - len(titleLine) - 2
	if padding < 0 {
		padding = 0
	}

	sb.WriteString("\n" + colorYellow)
	sb.WriteString(indent + cardTopLeft + cardHorizontal + titleLine + strings.Repeat(cardHorizontal, padding) + cardTopRight + "\n")
	sb.WriteString(colorReset)

	// Render card content
	for _, child := range block.Content {
		sb.WriteString(colorYellow + indent + cardVertical + colorReset)
		var childSb strings.Builder
		renderBlockRich(&childSb, &child, 0)
		lines := strings.Split(strings.TrimRight(childSb.String(), "\n"), "\n")
		for i, line := range lines {
			if i > 0 {
				sb.WriteString(colorYellow + indent + cardVertical + colorReset)
			}
			sb.WriteString(" " + line)
			sb.WriteString(colorYellow + colorReset + "\n")
		}
	}

	sb.WriteString(colorYellow + indent + cardBottomLeft + strings.Repeat(cardHorizontal, width) + cardBottomRight + colorReset + "\n\n")
}

// renderTextRich renders a text block with styling
func renderTextRich(sb *strings.Builder, block *models.Block, indent string) {
	md := block.Markdown
	if md == "" {
		sb.WriteString("\n")
		return
	}

	// Apply indentation
	extraIndent := strings.Repeat("  ", block.IndentationLevel)
	fullIndent := indent + extraIndent

	// Handle list styles
	prefix := ""
	switch block.ListStyle {
	case "bullet":
		prefix = bullet + " "
	case "numbered":
		prefix = "1. "
	case "task":
		if block.TaskInfo != nil {
			switch block.TaskInfo.State {
			case "done":
				prefix = taskDone + " "
				md = colorGreen + md + colorReset
			case "canceled":
				prefix = taskCanceled + " "
				md = colorDim + md + colorReset
			default:
				prefix = taskTodo + " "
			}
		} else {
			prefix = taskTodo + " "
		}
	case "toggle":
		prefix = toggleClosed + " "
	}

	// Handle decorations
	if sliceContains(block.Decorations, "callout") {
		color := colorBgYellow
		if block.Color != "" {
			color = getColorFromHex(block.Color)
		}
		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", fullIndent, color, md, colorReset))
	} else if sliceContains(block.Decorations, "quote") {
		sb.WriteString(fmt.Sprintf("%s%s│%s %s\n", fullIndent, colorDim, colorReset+colorItalic, md+colorReset))
	} else {
		// Apply text style
		styledMd := md
		switch block.TextStyle {
		case "h1":
			styledMd = colorBold + colorCyan + "# " + md + colorReset
		case "h2":
			styledMd = colorBold + "## " + md + colorReset
		case "h3":
			styledMd = colorBold + "### " + md + colorReset
		case "h4":
			styledMd = colorBold + "#### " + md + colorReset
		case "caption":
			styledMd = colorDim + colorItalic + md + colorReset
		}

		// Apply font style
		if block.Font == "mono" {
			styledMd = colorDim + "`" + md + "`" + colorReset
		}

		sb.WriteString(fullIndent + prefix + styledMd + "\n")
	}

	// Render children
	for _, child := range block.Content {
		renderBlockRich(sb, &child, 0)
	}
}

// renderTableRich renders a table with box drawing
func renderTableRich(sb *strings.Builder, block *models.Block, indent string) {
	// For now, just output the markdown table - could enhance with box drawing later
	sb.WriteString(indent + block.Markdown + "\n")
}

// getColorFromHex converts hex color to ANSI escape code
func getColorFromHex(hex string) string {
	// Map common Craft colors to ANSI
	switch strings.ToLower(hex) {
	case "#0064ff", "gradient-blue":
		return colorBgBlue
	case "#ff9500", "gradient-orange":
		return "\033[48;5;208m"
	case "#ff3b30", "gradient-red":
		return "\033[41m"
	case "#34c759", "gradient-green":
		return "\033[42m"
	case "#af52de", "gradient-purple":
		return "\033[45m"
	case "#ffcc00", "gradient-yellow":
		return colorBgYellow
	default:
		return colorBgCyan
	}
}

// sliceContains checks if a slice contains a string
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// OutputFormat constants
const (
	FormatJSON       = "json"
	FormatCompact    = "compact"
	FormatTable      = "table"
	FormatMarkdown   = "markdown"
	FormatStructured = "structured"
	FormatCraft      = "craft"
	FormatRich       = "rich"
)

// ValidOutputFormats lists all valid output formats
var ValidOutputFormats = []string{
	FormatJSON,
	FormatCompact,
	FormatTable,
	FormatMarkdown,
	FormatStructured,
	FormatCraft,
	FormatRich,
}

// IsValidFormat checks if a format string is valid
func IsValidFormat(format string) bool {
	for _, f := range ValidOutputFormats {
		if f == format {
			return true
		}
	}
	return false
}

func isJSONFormat(format string) bool {
	return format == FormatJSON || format == FormatCompact
}
