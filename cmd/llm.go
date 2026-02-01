package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type llmFlagSpec struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default"`
	Usage     string `json:"usage"`
}

type llmCommandSpec struct {
	Use     string           `json:"use"`
	Short   string           `json:"short,omitempty"`
	Long    string           `json:"long,omitempty"`
	Aliases []string         `json:"aliases,omitempty"`
	Flags   []llmFlagSpec    `json:"flags,omitempty"`
	Sub     []llmCommandSpec `json:"subcommands,omitempty"`
}

type llmSpec struct {
	Tool      string           `json:"tool"`
	Version   string           `json:"version"`
	Generated string           `json:"generatedAt"`
	Global    []llmFlagSpec    `json:"globalFlags"`
	Commands  []llmCommandSpec `json:"commands"`
	Notes     []string         `json:"notes"`
}

func flagsToSpec(fs *pflag.FlagSet) []llmFlagSpec {
	var specs []llmFlagSpec
	fs.VisitAll(func(f *pflag.Flag) {
		specs = append(specs, llmFlagSpec{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		})
	})
	return specs
}

func commandToSpec(c *cobra.Command) llmCommandSpec {
	spec := llmCommandSpec{
		Use:     c.Use,
		Short:   c.Short,
		Long:    c.Long,
		Aliases: c.Aliases,
		Flags:   flagsToSpec(c.Flags()),
	}

	for _, sc := range c.Commands() {
		if !sc.IsAvailableCommand() || sc.Name() == "help" {
			continue
		}
		spec.Sub = append(spec.Sub, commandToSpec(sc))
	}
	return spec
}

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Machine-readable command reference",
	Long: `Outputs a JSON schema-like description of commands, flags, and semantics.

Intended for LLMs/agents and scripting tools.

Subcommands:
  craft llm           Full command reference as JSON
  craft llm styles    Complete styling and formatting guide`,
	RunE: func(cmd *cobra.Command, args []string) error {
		spec := llmSpec{
			Tool:      "craft",
			Version:   version,
			Generated: time.Now().UTC().Format(time.RFC3339),
			Global:    flagsToSpec(rootCmd.PersistentFlags()),
			Notes: []string{
				"Default output is JSON. Use --format compact (legacy JSON), table, or markdown for human output where supported.",
				"craft delete is a soft-delete to trash (DELETE /documents).",
				"craft clear deletes all content blocks in a document (cannot be undone without a backup).",
				"craft update supports --mode append|replace and auto-chunks large markdown inserts.",
				"Run 'craft llm styles' for complete styling/formatting reference with JSON examples.",
			},
		}

		for _, c := range rootCmd.Commands() {
			if !c.IsAvailableCommand() || c.Name() == "help" {
				continue
			}
			spec.Commands = append(spec.Commands, commandToSpec(c))
		}

		return outputJSON(spec)
	},
}

var llmStylesCmd = &cobra.Command{
	Use:   "styles",
	Short: "Complete styling and formatting reference for Craft documents",
	Long:  "Outputs the full styling reference covering all block types, formatting options, decorations, cards, pages, highlights, dividers, and more. Designed for LLMs to understand how to create richly styled Craft documents.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(llmStylesReference)
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
	llmCmd.AddCommand(llmStylesCmd)
}

// llmStylesReference contains the complete styling guide embedded in the binary.
// This is the single source of truth for how to style Craft documents.
const llmStylesReference = `# Craft Styling and Formatting Reference

Craft documents are block-based. Every piece of content is a block. Blocks can
contain child blocks (pages and cards have content arrays).

## How to Create Content

Two methods to add content:

1. blocks_add — JSON blocks with full styling control:
   blocks_add({"pageId":"PAGE_ID","position":"end","blocks":[
     {"type":"text","textStyle":"h1","markdown":"# Heading"},
     {"type":"line","lineStyle":"regular"},
     {"type":"text","decorations":["callout"],"color":"#00ca85","markdown":"<callout>Note</callout>"}
   ]})

2. markdown_add — markdown string auto-converted to blocks:
   markdown_add({"pageId":"PAGE_ID","position":"end","markdown":"# Heading\n\n---\n\n<callout>Note</callout>"})

3. CLI:
   craft blocks add PAGE_ID --markdown "# Heading"

Positioning:
  pageId + position (start|end)     — add to a page
  siblingId + position (before|after) — add relative to a block
  date + position (start|end)       — add to a daily note

Update existing blocks:
  blocks_update({"blocks":[{"id":"BLOCK_ID","markdown":"New content"}]})

## Block Types

| Type     | Description                          | Key Fields                                           |
|----------|--------------------------------------|------------------------------------------------------|
| text     | Paragraphs, headings, lists, tasks   | textStyle, listStyle, decorations, color, font, textAlignment, indentationLevel |
| page     | Sub-page or card (has child blocks)  | textStyle (page or card), cardLayout, content[]      |
| code     | Code block or math formula           | language, rawCode                                    |
| line     | Divider / separator / page break     | lineStyle                                            |
| richUrl  | Smart link embed (YouTube, Figma)    | url, title, description, layout                      |
| image    | Image block                          | url, altText, size, width                            |
| file     | File attachment                      | url, fileName, blockLayout                           |
| table    | Table block                          | rows                                                 |

## Text Styles (textStyle)

  h1       # Title        (largest heading)
  h2       ## Subtitle
  h3       ### Heading
  h4       #### Strong
  (omit)   Body text      (default)
  caption  <caption>text</caption>

Example:
  {"type":"text","textStyle":"h1","markdown":"# My Title"}

## Inline Formatting (within markdown field)

  **bold**                    Bold
  *italic*                    Italic
  ~strikethrough~             Strikethrough
  ` + "`code`" + `                      Inline code
  [label](url)                Link
  $E=mc^2$                    Inline equation

## Highlighting

Use <highlight> tags in the markdown field:

Solid colors (9):
  <highlight color="yellow">text</highlight>
  <highlight color="green">text</highlight>
  <highlight color="mint">text</highlight>
  <highlight color="cyan">text</highlight>
  <highlight color="blue">text</highlight>
  <highlight color="purple">text</highlight>
  <highlight color="pink">text</highlight>
  <highlight color="red">text</highlight>
  <highlight color="gray">text</highlight>

Gradient colors (5):
  <highlight color="gradient-blue">text</highlight>
  <highlight color="gradient-purple">text</highlight>
  <highlight color="gradient-red">text</highlight>
  <highlight color="gradient-yellow">text</highlight>
  <highlight color="gradient-brown">text</highlight>

Highlights combine with other formatting:
  **<highlight color="gradient-brown">bold highlighted</highlight>**

To create highlighted text:
  blocks_add({"pageId":"ID","position":"end","blocks":[
    {"type":"text","markdown":"Check the <highlight color=\"red\">critical</highlight> items."}
  ]})

## Dividers (type: line)

  lineStyle     Visual                    Markdown Shortcut
  ─────────     ──────                    ─────────────────
  extraLight    Thinnest dotted line      ..-
  light         Thin line                 .--
  regular       Standard divider          ---
  strong        Thick bold separator      =-
  pageBreak     Page break for print      ===

Example:
  {"type":"line","lineStyle":"strong"}

To create dividers:
  blocks_add({"pageId":"ID","position":"end","blocks":[
    {"type":"line","lineStyle":"strong"}
  ]})
  markdown_add only creates "regular" dividers via ---. Use blocks_add for other styles.

## Lists (listStyle on text blocks)

Bullet:
  {"type":"text","listStyle":"bullet","markdown":"- Item"}

Numbered:
  {"type":"text","listStyle":"numbered","markdown":"1. Item"}

Task:
  {"type":"text","listStyle":"task","taskInfo":{"state":"todo"},"markdown":"- [ ] Task"}
  States: todo, done, canceled

Toggle (collapsible):
  {"type":"text","listStyle":"toggle","markdown":"+ Section"}

Nesting: use indentationLevel (0-5):
  {"type":"text","listStyle":"bullet","indentationLevel":1,"markdown":"  - Sub item"}

## Decorations

Quote (Focus):
  {"type":"text","decorations":["quote"],"markdown":"> Quoted text"}

Callout (Block):
  {"type":"text","decorations":["callout"],"markdown":"<callout>Important note</callout>"}

With color:
  {"type":"text","decorations":["callout"],"color":"#00ca85","markdown":"<callout>Green callout</callout>"}

Combine with textStyle:
  {"type":"text","textStyle":"h1","decorations":["callout"],"color":"#00ca85","markdown":"<callout># Green title callout</callout>"}

To create decorations:
  blocks_add({"pageId":"ID","position":"end","blocks":[
    {"type":"text","decorations":["callout"],"color":"#00ca85","markdown":"<callout>Green callout</callout>"},
    {"type":"text","decorations":["quote"],"markdown":"> Focus block"}
  ]})

## Text Alignment (textAlignment)

  left      Left-aligned (default)
  center    Centered
  right     Right-aligned
  justify   Justified

Example:
  {"type":"text","textAlignment":"center","markdown":"Centered text"}

## Block Colors (color)

Any text block accepts a color field with #RRGGBB hex:

  #ef052a   Red
  #ff9200   Orange
  #00ca85   Green
  #0400ff   Blue
  #c400ff   Purple
  #864d00   Brown

Example:
  {"type":"text","color":"#ef052a","markdown":"Red text"}

## Fonts (font)

  system    Default system font
  serif     Serif typeface
  mono      Monospaced typeface
  rounded   Rounded typeface

Example:
  {"type":"text","font":"serif","color":"#0400ff","textAlignment":"justify","markdown":"Blue serif justified"}

## Code Blocks

  {"type":"code","language":"javascript","rawCode":"const x = 1;"}

Languages: ada, bash, cpp, cs, css, dart, dockerfile, go, groovy, haskell,
  html, java, javascript, json, julia, kotlin, lua, markdown, objectivec,
  perl, php, plaintext, python, r, ruby, rust, scala, shell, sql, swift,
  typescript, vbnet, xml, yaml, math_formula, other

Math formula:
  {"type":"code","language":"math_formula","rawCode":"E = mc^2"}

## Rich URLs (Smart Links / Embeds)

  {"type":"richUrl","url":"https://youtube.com/watch?v=ID","title":"Video Title","description":"..."}

Layout options: small, regular, card

Use richUrl for embed-style previews. Use [text](url) in markdown for inline links.

To create rich URLs (requires blocks_add — cannot use markdown_add):
  blocks_add({"pageId":"ID","position":"end","blocks":[
    {"type":"richUrl","url":"https://youtube.com/watch?v=ID","title":"Video","description":"..."}
  ]})

## Pages (Nested Documents)

Pages have a content array of child blocks:

  {
    "type": "page",
    "textStyle": "page",
    "markdown": "Sub-Page Title",
    "content": [
      {"type":"text","markdown":"Content inside the sub-page"},
      {"type":"text","textStyle":"h2","markdown":"## Heading inside"}
    ]
  }

Markdown shortcut:
  <page><pageTitle>Title</pageTitle><content>Body</content></page>

## Cards

A card is a page with textStyle "card". Cards display as visual cards.

Card layouts: small, square, regular, large

  {
    "type": "page",
    "textStyle": "card",
    "cardLayout": "regular",
    "markdown": "Card Title",
    "content": [
      {"type":"text","markdown":"Content inside the card"},
      {"type":"text","listStyle":"bullet","markdown":"- Cards support ALL styling"},
      {"type":"code","language":"python","rawCode":"print('even code blocks')"}
    ]
  }

Content inside cards supports ALL styling: headings, lists, code, dividers,
colors, fonts, decorations, images, nested pages/cards. No limitations.

To create cards:
  blocks_add({"pageId":"ID","position":"end","blocks":[
    {"type":"page","textStyle":"card","cardLayout":"regular","markdown":"Card Title",
     "content":[
       {"type":"text","textStyle":"h2","markdown":"## Section"},
       {"type":"text","listStyle":"task","taskInfo":{"state":"todo"},"markdown":"- [ ] Todo"}
     ]}
  ]})

Markdown shortcut:
  <page textStyle="card" cardLayout="regular"><pageTitle>Title</pageTitle><content>Body</content></page>

## Images

  {"type":"image","url":"https://example.com/img.jpg","altText":"Description"}

  size: fit | fill
  width: auto | fullWidth

## Files

  {"type":"file","url":"https://r.craft.do/fileId","fileName":"report.pdf","blockLayout":"regular"}

  blockLayout: small | regular | card

## Combining Styles

A single block can combine all styling fields:

  {
    "type": "text",
    "textStyle": "h2",
    "decorations": ["callout"],
    "color": "#00ca85",
    "font": "serif",
    "textAlignment": "center",
    "markdown": "## Green centered serif callout heading"
  }

Common combinations:
  - Colored heading: textStyle + color
  - Styled callout: decorations + textStyle + color
  - Fancy list: listStyle + indentationLevel + color
  - Styled paragraph: font + color + textAlignment
`
