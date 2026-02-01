# Craft Styling and Formatting Reference

This is the definitive reference for all Craft document styling. It covers every block type, every formatting option, and every visual feature available in Craft documents.

Craft documents are **block-based**. Every piece of content — a paragraph, heading, divider, image, card — is a block. Blocks can contain child blocks (pages contain content, cards contain content). Understanding this tree structure is essential.

---

## How to Create Content

There are two ways to add styled content to Craft documents:

### Method 1: `blocks_add` (JSON blocks with full styling control)

Use `blocks_add` when you need precise control over block type, styling fields, and structure. Each block is a JSON object with `type` and styling properties.

```json
blocks_add({
  "pageId": "TARGET_PAGE_ID",
  "position": "end",
  "blocks": [
    {
      "type": "text",
      "textStyle": "h1",
      "markdown": "# My Heading"
    },
    {
      "type": "line",
      "lineStyle": "regular"
    },
    {
      "type": "text",
      "decorations": ["callout"],
      "color": "#00ca85",
      "markdown": "<callout>An important note</callout>"
    }
  ]
})
```

### Method 2: `markdown_add` (markdown string, auto-converted to blocks)

Use `markdown_add` for quick content insertion using markdown syntax. Craft-specific HTML tags (`<highlight>`, `<callout>`, `<caption>`, `<page>`) work inside the markdown string.

```json
markdown_add({
  "pageId": "TARGET_PAGE_ID",
  "position": "end",
  "markdown": "# My Heading\n\n---\n\n<callout>An important note</callout>"
})
```

### Method 3: CLI (`craft blocks add`)

```bash
craft blocks add PAGE_ID --markdown "# My Heading"
craft blocks add PAGE_ID --markdown "<callout>An important note</callout>"
```

### Positioning

Both methods support positioning:

| Parameter | Values | Description |
|-----------|--------|-------------|
| `pageId` + `position` | `start`, `end` | Add at beginning or end of a page |
| `siblingId` + `position` | `before`, `after` | Add relative to a specific block |
| `date` + `position` | `start`, `end` | Add to a daily note by date |

### Updating existing blocks

Use `blocks_update` to change an existing block's content or styling:

```json
blocks_update({
  "blocks": [
    {
      "id": "EXISTING_BLOCK_ID",
      "markdown": "Updated text with <highlight color=\"yellow\">highlighting</highlight>"
    }
  ]
})
```

---

## Block Types Overview

| Type | Description | Key Fields |
|------|-------------|------------|
| `text` | Paragraphs, headings, lists, tasks, toggles | `textStyle`, `listStyle`, `decorations`, `color`, `font`, `textAlignment`, `indentationLevel` |
| `page` | Sub-page or card (contains child blocks) | `textStyle` (`page` or `card`), `cardLayout`, `content[]` |
| `code` | Code block or math formula | `language`, `rawCode` |
| `line` | Divider / separator / page break | `lineStyle` |
| `richUrl` | Smart link embed (YouTube, Figma, websites) | `url`, `title`, `description`, `layout` |
| `image` | Image block | `url`, `altText`, `size`, `width` |
| `file` | File attachment | `url`, `fileName`, `blockLayout` |
| `table` | Table block | `rows` |
| `whiteboard` | Embedded whiteboard | `url` |

---

## 1. Text Styles (Headings and Body)

Text blocks use `textStyle` to control their visual weight. When omitted, the default is body text.

| textStyle | Markdown Shortcut | Appearance |
|-----------|------------------|------------|
| `h1` | `# ` | Title — largest heading |
| `h2` | `## ` | Subtitle |
| `h3` | `### ` | Heading |
| `h4` | `#### ` | Strong text |
| _(omitted)_ | _(plain text)_ | Body — normal paragraph |
| `caption` | `<caption>text</caption>` | Small, subdued caption text |

### JSON Examples

**Title (h1):**
```json
{ "type": "text", "textStyle": "h1", "markdown": "# My Title" }
```

**Subtitle (h2):**
```json
{ "type": "text", "textStyle": "h2", "markdown": "## My Subtitle" }
```

**Heading (h3):**
```json
{ "type": "text", "textStyle": "h3", "markdown": "### Section Heading" }
```

**Strong (h4):**
```json
{ "type": "text", "textStyle": "h4", "markdown": "#### Strong Text" }
```

**Body (default):**
```json
{ "type": "text", "markdown": "This is a normal paragraph." }
```

**Caption:**
```json
{ "type": "text", "textStyle": "caption", "markdown": "<caption>Small caption text</caption>" }
```

### Markdown Shortcut (via `markdown_add` or `blocks_add --markdown`)

```markdown
# Title

## Subtitle

### Heading

#### Strong

Body text here

<caption>Caption text</caption>
```

---

## 2. Inline Formatting

Inline formatting is applied within the `markdown` field of any text block. These can be combined.

| Format | Syntax | Example |
|--------|--------|---------|
| Bold | `**text**` or `__text__` | `**bold**` |
| Italic | `*text*` or `_text_` | `*italic*` |
| Bold + Italic | `***text***` or `___text___` | `***both***` |
| Strikethrough | `~text~` | `~struck~` |
| Inline code | `` `text` `` | `` `code` `` |
| Link | `[label](url)` | `[My Site](https://example.com)` |
| Inline equation | `$equation$` | `$E=mc^2$` |
| Page/block link | `@` or `[[` | `@Page Name` or `[[Page Name]]` |

### JSON Example — Mixed inline formatting

```json
{
  "type": "text",
  "markdown": "This has **bold**, *italics*, ~strikethrough~, and `code inline`."
}
```

---

## 3. Highlighting

Highlights use the `<highlight>` HTML tag inside the markdown field. Color is specified via the `color` attribute.

### Solid Highlight Colors (9 colors)

| Color | Value | Syntax |
|-------|-------|--------|
| Yellow | `yellow` | `<highlight color="yellow">text</highlight>` |
| Green | `green` | `<highlight color="green">text</highlight>` |
| Mint | `mint` | `<highlight color="mint">text</highlight>` |
| Cyan | `cyan` | `<highlight color="cyan">text</highlight>` |
| Blue | `blue` | `<highlight color="blue">text</highlight>` |
| Purple | `purple` | `<highlight color="purple">text</highlight>` |
| Pink | `pink` | `<highlight color="pink">text</highlight>` |
| Red | `red` | `<highlight color="red">text</highlight>` |
| Gray | `gray` | `<highlight color="gray">text</highlight>` |

### Gradient Highlight Colors (5 colors)

| Color | Value | Syntax |
|-------|-------|--------|
| Gradient Blue | `gradient-blue` | `<highlight color="gradient-blue">text</highlight>` |
| Gradient Purple | `gradient-purple` | `<highlight color="gradient-purple">text</highlight>` |
| Gradient Red | `gradient-red` | `<highlight color="gradient-red">text</highlight>` |
| Gradient Yellow | `gradient-yellow` | `<highlight color="gradient-yellow">text</highlight>` |
| Gradient Brown | `gradient-brown` | `<highlight color="gradient-brown">text</highlight>` |

### JSON Example — Multiple highlight styles

```json
{
  "type": "text",
  "markdown": "A <highlight color=\"yellow\">yellow</highlight> word and a <highlight color=\"gradient-blue\">gradient blue</highlight> word."
}
```

### Highlights combined with other formatting

Highlights can be nested inside bold, italic, or other inline formatting:

```json
{
  "type": "text",
  "markdown": "**<highlight color=\"gradient-brown\">bold and highlighted</highlight> <highlight color=\"purple\">*italic and highlighted*</highlight>**"
}
```

### How to Create Highlighted Text

**Via `blocks_add`:**
```json
blocks_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "blocks": [
    {
      "type": "text",
      "markdown": "Check the <highlight color=\"red\">critical</highlight> items and the <highlight color=\"gradient-blue\">featured</highlight> section."
    }
  ]
})
```

**Via `markdown_add`:**
```json
markdown_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "markdown": "Check the <highlight color=\"red\">critical</highlight> items and the <highlight color=\"gradient-blue\">featured</highlight> section."
})
```

**Via CLI:**
```bash
craft blocks add PAGE_ID --markdown 'Check the <highlight color="red">critical</highlight> items.'
```

### Markdown Shortcut

Plain highlights (no color) can be created with `::text::` or `==text==`. For colored highlights, use the HTML tag syntax shown above.

---

## 4. Dividers and Lines

Dividers are `line` type blocks. The `lineStyle` field controls thickness and purpose. There are 5 distinct styles from thinnest to thickest, plus a page break.

| lineStyle | Visual | Markdown Shortcut |
|-----------|--------|------------------|
| `extraLight` | Thinnest, barely visible dotted line | `..−` (three dots + dash) |
| `light` | Thin line | `.--` |
| `regular` | Standard divider | `---` |
| `strong` | Thick, bold separator | `=-` |
| `pageBreak` | Forces a new page in print/export | `===` |

### JSON Examples

**Extra light divider:**
```json
{ "type": "line", "lineStyle": "extraLight" }
```

**Light divider:**
```json
{ "type": "line", "lineStyle": "light" }
```

**Regular divider:**
```json
{ "type": "line", "lineStyle": "regular" }
```

**Strong divider:**
```json
{ "type": "line", "lineStyle": "strong" }
```

**Page break:**
```json
{ "type": "line", "lineStyle": "pageBreak" }
```

### How to Create Dividers

**Via `blocks_add`:**
```json
blocks_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "blocks": [
    { "type": "line", "lineStyle": "strong" }
  ]
})
```

**Via `markdown_add`:**
```json
markdown_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "markdown": "---"
})
```
Note: `markdown_add` creates a `regular` divider with `---`. For other styles, use `blocks_add`.

**Via CLI:**
```bash
craft blocks add PAGE_ID --markdown "---"
```

### Usage Notes

- Use `extraLight` for subtle visual separation within a section.
- Use `regular` for standard section breaks.
- Use `strong` for major section dividers.
- Use `pageBreak` when the document will be printed or exported to PDF.

---

## 5. Lists

Lists use the `listStyle` field on text blocks. Nesting is controlled by `indentationLevel` (0-5).

### Bullet Lists

```json
{ "type": "text", "listStyle": "bullet", "markdown": "- First item" }
```

**Nested bullet (indent level 1):**
```json
{ "type": "text", "listStyle": "bullet", "indentationLevel": 1, "markdown": "  - Sub item" }
```

**Deeply nested (indent level 2):**
```json
{ "type": "text", "listStyle": "bullet", "indentationLevel": 2, "markdown": "    - Sub sub item" }
```

### Numbered Lists

```json
{ "type": "text", "listStyle": "numbered", "markdown": "1. First item" }
```

**Nested numbered:**
```json
{ "type": "text", "listStyle": "numbered", "indentationLevel": 1, "markdown": "  1. Sub item" }
```

### Task Lists (Checkboxes)

Tasks have a `taskInfo` object with a `state` field.

**Unchecked task:**
```json
{
  "type": "text",
  "listStyle": "task",
  "taskInfo": { "state": "todo" },
  "markdown": "- [ ] Buy groceries"
}
```

**Completed task:**
```json
{
  "type": "text",
  "listStyle": "task",
  "taskInfo": { "state": "done" },
  "markdown": "- [x] Buy groceries"
}
```

**Canceled task:**
```json
{
  "type": "text",
  "listStyle": "task",
  "taskInfo": { "state": "canceled" },
  "markdown": "- [-] No longer needed"
}
```

Tasks can also have `scheduleDate`, `deadlineDate`, and `repeat` configuration in `taskInfo`.

### Toggle Lists (Collapsible Sections)

Toggles create expandable/collapsible sections. Child items are nested with `indentationLevel`.

**Top-level toggle:**
```json
{ "type": "text", "listStyle": "toggle", "markdown": "+ Section title" }
```

**Toggle child items (nested under the toggle above):**
```json
{ "type": "text", "listStyle": "toggle", "indentationLevel": 1, "markdown": "  + Sub item 1" }
```
```json
{ "type": "text", "listStyle": "toggle", "indentationLevel": 2, "markdown": "    + Nested under sub item" }
```

### Markdown Shortcuts for Lists

```markdown
- Bullet item
* Also a bullet
1. Numbered item
- [ ] Todo task
- [x] Done task
+ Toggle section
```

---

## 6. Decorations (Callouts and Quotes)

Decorations wrap blocks in visual containers. The `decorations` array can contain `"callout"` and/or `"quote"`.

### Quote (Focus Block)

A quote creates a left-border styled block (like a blockquote). In Craft's UI this is called "Focus".

```json
{
  "type": "text",
  "decorations": ["quote"],
  "markdown": "> This is a focused/quoted block"
}
```

### Callout (Block)

A callout creates a highlighted background container.

```json
{
  "type": "text",
  "decorations": ["callout"],
  "markdown": "<callout>This is inside a callout block</callout>"
}
```

### Decorations with Color

Both quotes and callouts can have a `color` field (hex `#RRGGBB`) to tint the decoration:

**Purple quote with caption style:**
```json
{
  "type": "text",
  "textStyle": "caption",
  "decorations": ["quote"],
  "color": "#c400ff",
  "markdown": "> <caption>Purple focus with caption text</caption>"
}
```

**Green callout with h1 style:**
```json
{
  "type": "text",
  "textStyle": "h1",
  "decorations": ["callout"],
  "color": "#00ca85",
  "markdown": "<callout># Green callout with title styling</callout>"
}
```

### How to Create Decorations

**Via `blocks_add`:**
```json
blocks_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "blocks": [
    {
      "type": "text",
      "decorations": ["callout"],
      "color": "#00ca85",
      "textStyle": "h2",
      "markdown": "<callout>## Important Decision</callout>"
    },
    {
      "type": "text",
      "decorations": ["quote"],
      "color": "#c400ff",
      "markdown": "> A purple focus block"
    }
  ]
})
```

**Via `markdown_add`:**
```json
markdown_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "markdown": "<callout>Important note in a callout</callout>\n\n> A quoted/focus block"
})
```

### Key Point

Decorations, text styles, and colors can all be combined on a single block. A callout can contain an h1, be colored green, and use a serif font — all at once.

---

## 7. Text Alignment

The `textAlignment` field controls horizontal alignment. When omitted, text is left-aligned (default).

| textAlignment | Effect |
|--------------|--------|
| `left` | Left-aligned (default) |
| `center` | Centered |
| `right` | Right-aligned |
| `justify` | Justified (stretched to fill width) |

### JSON Examples

```json
{ "type": "text", "textAlignment": "center", "markdown": "Centered text" }
```

```json
{ "type": "text", "textAlignment": "right", "markdown": "Right-aligned text" }
```

```json
{ "type": "text", "textAlignment": "justify", "markdown": "Justified paragraph text" }
```

---

## 8. Block Colors

Any text block can have a `color` field with a `#RRGGBB` hex value. This changes the text color of the entire block.

### Common Craft Colors

| Color | Hex |
|-------|-----|
| Red | `#ef052a` |
| Orange | `#ff9200` |
| Green | `#00ca85` |
| Blue | `#0400ff` |
| Purple | `#c400ff` |
| Brown | `#864d00` |

### JSON Example

```json
{
  "type": "text",
  "color": "#ef052a",
  "markdown": "This text is red"
}
```

Colors can be combined with any other styling (alignment, font, decorations):

```json
{
  "type": "text",
  "textAlignment": "justify",
  "font": "serif",
  "color": "#0400ff",
  "markdown": "Blue serif justified text"
}
```

---

## 9. Fonts

The `font` field changes the typeface of a text block. When omitted, the system font is used.

| font | Description |
|------|-------------|
| `system` | Default system font (San Francisco on Apple) |
| `serif` | Traditional serif typeface |
| `mono` | Monospaced / code-like typeface |
| `rounded` | Rounded, friendly typeface |

### JSON Examples

```json
{ "type": "text", "font": "serif", "markdown": "Serif font text" }
```

```json
{ "type": "text", "font": "mono", "markdown": "Monospaced font text" }
```

```json
{ "type": "text", "font": "rounded", "markdown": "Rounded font text" }
```

### Combining font + color + alignment

```json
{
  "type": "text",
  "font": "rounded",
  "color": "#c400ff",
  "textAlignment": "justify",
  "markdown": "Purple rounded justified text"
}
```

---

## 10. Code Blocks

Code blocks use `type: "code"` with a `language` field and `rawCode` for the actual code content.

### Supported Languages

`ada`, `bash`, `cpp`, `cs`, `css`, `dart`, `dockerfile`, `matlab`, `go`, `groovy`, `haskell`, `html`, `java`, `javascript`, `json`, `julia`, `kotlin`, `lua`, `markdown`, `objectivec`, `perl`, `php`, `prolog`, `plaintext`, `python`, `r`, `ruby`, `rust`, `scala`, `shell`, `sql`, `swift`, `typescript`, `vbnet`, `xml`, `yaml`, `math_formula`, `other`

### JSON Example — JavaScript code block

```json
{
  "type": "code",
  "language": "javascript",
  "rawCode": "const greeting = \"Hello, world!\";\n\nfunction sayHello(name) {\n  console.log(`${greeting} My name is ${name}.`);\n}\n\nsayHello(\"Claude\");",
  "markdown": "```javascript\nconst greeting = \"Hello, world!\";\n\nfunction sayHello(name) {\n  console.log(`${greeting} My name is ${name}.`);\n}\n\nsayHello(\"Claude\");\n```"
}
```

### Math Formula

Craft treats math formulas as a special code block with `language: "math_formula"`. The `rawCode` contains LaTeX:

```json
{
  "type": "code",
  "language": "math_formula",
  "rawCode": "d_1 = \\frac{1}{\\sigma \\sqrt{t}} \\left[ \\ln\\left(\\frac{S_t}{K}\\right) + \\left(r + \\frac{\\sigma^2}{2}\\right)t \\right]"
}
```

### Markdown Shortcut

````markdown
```javascript
const x = 1;
```
````

Or for math: `$E=mc^2$` (inline) or a code block with `math_formula` language.

---

## 11. Rich URLs (Smart Links / Embeds)

Rich URLs create visual link previews with title, description, and optional thumbnail. They are used for embedding websites, YouTube videos, Figma files, and any URL.

### JSON Structure

```json
{
  "type": "richUrl",
  "url": "https://example.com",
  "title": "Page Title",
  "description": "A description of the linked page.",
  "layout": "regular"
}
```

### Layout Options

| layout | Description |
|--------|-------------|
| `small` | Compact inline preview |
| `regular` | Standard card-style preview |
| `card` | Full card with large thumbnail |

### JSON Example — YouTube embed

```json
{
  "type": "richUrl",
  "url": "https://www.youtube.com/watch?v=Bl8GK7tjqOU",
  "title": "Easy One-Pot Chicken Dinner Recipe",
  "description": "A delicious recipe walkthrough..."
}
```

### JSON Example — Figma embed

```json
{
  "type": "richUrl",
  "url": "https://www.figma.com/make/tZ3VCGR1NybMXpYPtUCFQT/Interactive-Components-Tutorial",
  "title": "Figma",
  "description": "Created with Figma"
}
```

### JSON Example — Website smart link

```json
{
  "type": "richUrl",
  "url": "https://ashrafali.net",
  "title": "Ashraf Ali",
  "description": "Personal website"
}
```

### How to Create Rich URLs

**Via `blocks_add`:**
```json
blocks_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "blocks": [
    {
      "type": "richUrl",
      "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
      "title": "Video Title",
      "description": "A description of the video"
    }
  ]
})
```

Rich URLs require `blocks_add` — they cannot be created via `markdown_add` because they need explicit `type`, `url`, `title`, and `description` fields.

### Inline Links vs Rich URLs

A plain `[text](url)` link inside a text block creates an inline hyperlink. A `richUrl` block creates a standalone visual preview card. Use rich URLs when you want an embed-style preview, not just a clickable link.

---

## 12. Pages (Nested Documents)

A `page` block is a sub-document that lives inside another page. Pages have their own `content` array containing child blocks. **This is how Craft creates document hierarchy.**

### Key Concept: Content Lives Inside Pages

When you see `"type": "page"` with `"content": [...]`, those child blocks are the content of that sub-page. You can click into the page in Craft's UI to see its contents.

### JSON Example — Sub-page with content

```json
{
  "type": "page",
  "textStyle": "page",
  "markdown": "My Sub-Page Title",
  "content": [
    {
      "type": "text",
      "markdown": "This text lives inside the sub-page."
    },
    {
      "type": "text",
      "textStyle": "h2",
      "markdown": "## A heading inside the sub-page"
    },
    {
      "type": "text",
      "listStyle": "bullet",
      "markdown": "- A bullet inside the sub-page"
    }
  ]
}
```

### Creating a Sub-Page via Markdown

```markdown
<page><pageTitle>My Sub-Page</pageTitle><content>Content goes here</content></page>
```

---

## 13. Cards

Cards are a special visual variant of pages. They appear as visual cards in the parent document rather than simple page links. **A card is a page with `textStyle: "card"`.**

### Card Layouts

| cardLayout | Description |
|------------|-------------|
| `small` | Compact card, minimal preview |
| `square` | Square aspect ratio card |
| `regular` | Standard card size (default) |
| `large` | Large card with more preview space |

### Key Concept: Cards Contain Content

Just like pages, cards have a `content` array. The child blocks inside the card are its content — visible when you click into the card. The card's `markdown` field is its title.

### JSON Example — Small card

```json
{
  "type": "page",
  "textStyle": "card",
  "cardLayout": "small",
  "markdown": "Quick Note",
  "content": [
    { "type": "text", "markdown": "Brief content inside the card" }
  ]
}
```

### JSON Example — Regular card

```json
{
  "type": "page",
  "textStyle": "card",
  "cardLayout": "regular",
  "markdown": "Project Overview",
  "content": [
    { "type": "text", "textStyle": "h2", "markdown": "## Goals" },
    { "type": "text", "listStyle": "bullet", "markdown": "- Ship feature X" },
    { "type": "text", "listStyle": "bullet", "markdown": "- Fix bug Y" }
  ]
}
```

### JSON Example — Large card

```json
{
  "type": "page",
  "textStyle": "card",
  "cardLayout": "large",
  "markdown": "Feature Spec",
  "content": [
    { "type": "text", "markdown": "Detailed specification document" },
    { "type": "line", "lineStyle": "regular" },
    { "type": "text", "textStyle": "h3", "markdown": "### Requirements" },
    { "type": "text", "listStyle": "task", "taskInfo": { "state": "todo" }, "markdown": "- [ ] Requirement A" }
  ]
}
```

### How to Create Cards

**Via `blocks_add`:**
```json
blocks_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "blocks": [
    {
      "type": "page",
      "textStyle": "card",
      "cardLayout": "regular",
      "markdown": "Meeting Notes - Jan 30",
      "content": [
        { "type": "text", "textStyle": "h2", "markdown": "## Decisions" },
        { "type": "text", "decorations": ["callout"], "color": "#00ca85", "markdown": "<callout>Ship by Feb 15</callout>" },
        { "type": "line", "lineStyle": "light" },
        { "type": "text", "listStyle": "task", "taskInfo": { "state": "todo" }, "markdown": "- [ ] Write spec" },
        { "type": "text", "listStyle": "task", "taskInfo": { "state": "todo" }, "markdown": "- [ ] Review with team" }
      ]
    }
  ]
})
```

**Via `markdown_add`:**
```json
markdown_add({
  "pageId": "PAGE_ID",
  "position": "end",
  "markdown": "<page textStyle='card' cardLayout='regular'><pageTitle>Meeting Notes</pageTitle><content>## Decisions\n\n<callout>Ship by Feb 15</callout>\n\n- [ ] Write spec\n- [ ] Review with team</content></page>"
})
```

### Styling Inside Cards

Content inside a card supports ALL the same styling as any document — headings, lists, code blocks, dividers, colors, fonts, decorations, images, and even nested pages/cards. There is no limitation.

### Creating Cards via Markdown

```markdown
<page textStyle="card" cardLayout="regular"><pageTitle>Card Title</pageTitle><content>Card body content</content></page>
```

For small card: `<page textStyle="card" cardLayout="small">`
For large card: `<page textStyle="card" cardLayout="large">`
For square card: `<page textStyle="card" cardLayout="square">`

---

## 14. Images

Image blocks reference a URL. They can be uploaded images (hosted by Craft) or external URLs.

### JSON Example

```json
{
  "type": "image",
  "url": "https://r.craft.do/7kRrQjUZV6",
  "altText": "A description of the image",
  "size": "fit",
  "width": "auto"
}
```

### Size and Width Options

| Field | Values | Description |
|-------|--------|-------------|
| `size` | `fit`, `fill` | `fit` = contained, `fill` = cover |
| `width` | `auto`, `fullWidth` | `auto` = natural, `fullWidth` = stretch |

---

## 15. Files

File blocks represent uploaded file attachments.

```json
{
  "type": "file",
  "url": "https://r.craft.do/fileId",
  "fileName": "report.pdf",
  "blockLayout": "regular"
}
```

| blockLayout | Description |
|-------------|-------------|
| `small` | Compact file reference |
| `regular` | Standard file card |
| `card` | Full card-style file preview |

---

## 16. Combining Multiple Styles on One Block

A single text block can combine multiple styling properties simultaneously. Here is the full set of combinable fields:

```json
{
  "type": "text",
  "textStyle": "h2",
  "listStyle": "bullet",
  "decorations": ["callout"],
  "color": "#00ca85",
  "font": "serif",
  "textAlignment": "center",
  "indentationLevel": 1,
  "markdown": "## A green, centered, serif, bulleted, indented callout heading"
}
```

Not all combinations make visual sense, but the API allows them. Common practical combinations:

- **Colored heading:** `textStyle` + `color`
- **Callout with heading:** `decorations: ["callout"]` + `textStyle: "h1"` + `color`
- **Styled list:** `listStyle` + `indentationLevel` + `color`
- **Fancy paragraph:** `font` + `color` + `textAlignment`

---

## 17. Full Document Example

Here is a complete document structure combining many features:

```json
{
  "type": "page",
  "textStyle": "page",
  "markdown": "Project Kickoff Notes",
  "content": [
    { "type": "text", "textStyle": "h1", "markdown": "# Project Kickoff" },
    { "type": "text", "textStyle": "caption", "markdown": "<caption>January 30, 2026</caption>" },
    { "type": "line", "lineStyle": "regular" },
    { "type": "text", "textStyle": "h2", "markdown": "## Attendees" },
    { "type": "text", "listStyle": "bullet", "markdown": "- Alice (PM)" },
    { "type": "text", "listStyle": "bullet", "markdown": "- Bob (Eng)" },
    { "type": "line", "lineStyle": "light" },
    { "type": "text", "textStyle": "h2", "markdown": "## Key Decisions" },
    { "type": "text", "decorations": ["callout"], "color": "#00ca85",
      "markdown": "<callout>We will use React + TypeScript for the frontend.</callout>" },
    { "type": "text", "textStyle": "h3", "markdown": "### Action Items" },
    { "type": "text", "listStyle": "task", "taskInfo": { "state": "todo" }, "markdown": "- [ ] Set up CI pipeline" },
    { "type": "text", "listStyle": "task", "taskInfo": { "state": "todo" }, "markdown": "- [ ] Design database schema" },
    { "type": "line", "lineStyle": "strong" },
    {
      "type": "page",
      "textStyle": "card",
      "cardLayout": "regular",
      "markdown": "Technical Spec",
      "content": [
        { "type": "text", "markdown": "Detailed spec lives in this card." },
        { "type": "code", "language": "typescript", "rawCode": "interface Project {\n  id: string;\n  name: string;\n}" }
      ]
    },
    {
      "type": "richUrl",
      "url": "https://github.com/org/repo",
      "title": "Project Repository",
      "description": "Main codebase"
    }
  ]
}
```

---

## Reference: All Styling Fields

### Block-Level Properties

| Field | Type | Values | Applies To |
|-------|------|--------|------------|
| `type` | string | `text`, `page`, `code`, `line`, `richUrl`, `image`, `file`, `table`, `whiteboard` | All blocks |
| `textStyle` | string | `h1`, `h2`, `h3`, `h4`, `page`, `card`, `caption` | text, page |
| `listStyle` | string | `none`, `bullet`, `numbered`, `task`, `toggle` | text |
| `decorations` | string[] | `callout`, `quote` | text |
| `color` | string | `#RRGGBB` hex | text |
| `font` | string | `system`, `serif`, `mono`, `rounded` | text |
| `textAlignment` | string | `left`, `center`, `right`, `justify` | text |
| `indentationLevel` | int | `0` - `5` | text |
| `cardLayout` | string | `small`, `square`, `regular`, `large` | page (when textStyle=card) |
| `lineStyle` | string | `extraLight`, `light`, `regular`, `strong`, `pageBreak` | line |
| `language` | string | See code block section | code |
| `layout` | string | `small`, `regular`, `card` | richUrl |

### Inline Formatting (within `markdown` field)

| Format | Syntax |
|--------|--------|
| Bold | `**text**` |
| Italic | `*text*` |
| Strikethrough | `~text~` |
| Inline code | `` `text` `` |
| Link | `[label](url)` |
| Highlight (solid) | `<highlight color="COLOR">text</highlight>` where COLOR is: yellow, green, mint, cyan, blue, purple, pink, red, gray |
| Highlight (gradient) | `<highlight color="GRADIENT">text</highlight>` where GRADIENT is: gradient-blue, gradient-purple, gradient-red, gradient-yellow, gradient-brown |
| Caption | `<caption>text</caption>` |
| Callout (block-level) | `<callout>text</callout>` |
| Equation | `$LaTeX$` |
