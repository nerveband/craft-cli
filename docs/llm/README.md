# Craft CLI -- LLM Reference

Quick index for LLMs, agents, and automation tools.

## Documentation

- `styling-and-markdown.md` -- Complete styling reference with JSON examples for every block type, formatting option, decoration, card layout, divider style, highlight color, and more
- `output-parity.md` -- MCP/API/CLI parity notes + differences chart

## CLI Discovery

LLMs can discover styling documentation directly via the CLI:

```bash
craft llm              # Full command reference as JSON
craft llm styles       # Complete styling and formatting guide
```

## Reference Payloads

- `../craft-everything-in-one.json` -- Full JSON snapshot of a document exercising every Craft feature (headings, dividers, code, cards, pages, highlights, rich URLs, decorations, fonts, colors, etc.)

## Key Concepts for LLMs

1. **Everything is a block.** Paragraphs, headings, dividers, images, cards -- all blocks.
2. **Pages and cards contain content.** A `page` or `card` block has a `content[]` array of child blocks.
3. **Styling is composable.** A single block can have `textStyle` + `color` + `font` + `decorations` + `textAlignment` all at once.
4. **Markdown and JSON coexist.** Use markdown shortcuts for quick content, JSON block properties for precise styling control.
