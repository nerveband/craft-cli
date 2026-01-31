# Styling and Markdown

This CLI can **read and write styled Craft content**. You can use markdown shortcuts or Craft block JSON styling fields.

## Markdown Style Shortcuts

Full block commands:
- `# ` title
- `## ` subtitle
- `### ` heading
- `#### ` strong block
- `x ` or `[] ` uncompleted todo
- `[x] ` completed todo
- `- ` or `* ` bullet list
- `1. ` `2. ` `3. ` numbered list
- `+ ` toggle list
- `> ` or `| ` block quote
- ``` ``` or `''' ` code block
- `=-` (no spaces) strong line
- `---` (no spaces) regular line
- `.--` (no spaces) dotted line
- `..-` (no spaces) three-dots line
- `===` (no spaces) page break

Inline commands:
- `*italic*` or `_italic_`
- `**bold**` or `__bold__`
- `***bold+italic***` or `___bold+italic___`
- `~~strikethrough~~`
- `::highlight::` or `==highlight==`
- `[link name](url)`
- `` `inline code` ``
- `$equation$`
- `@` or `[[` page/block links

## Styling via Block JSON

Craft supports styling beyond markdown. When you fetch blocks via MCP/CLI structured output, you’ll see fields like:

- `textStyle`: `h1`, `h2`, `h3`, `h4`, `page`, `card`, `caption`
- `listStyle`: `bullet`, `numbered`, `task`, `toggle`
- `decorations`: `callout`, `quote`
- `color`: hex color like `#FFB100`
- `font`: `system`, `serif`, `mono`, `rounded`
- `textAlignment`: `left`, `center`, `right`
- `cardLayout`: `small`, `regular`, `large`
- `indentationLevel`: `0-5`

Example block (JSON):

```json
{
  "id": "block-id",
  "type": "text",
  "textStyle": "h2",
  "markdown": "Styled Header",
  "color": "#FFB100",
  "font": "serif",
  "textAlignment": "center",
  "decorations": ["callout"],
  "listStyle": "bullet",
  "indentationLevel": 1
}
```

## CLI Examples

Create a styled doc using markdown shortcuts:

```bash
craft create --title "Style Demo" --markdown "# Title\n## Subtitle\n- Bullet\n[x] Done" 
```

Inspect full block styling (MCP-style block tree):

```bash
craft get <doc-id> --format structured
```

Get MCP-style XML markdown:

```bash
craft get <doc-id> --format craft
```
