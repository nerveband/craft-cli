# Craft CLI -- LLM Reference

Quick index for LLMs, agents, and automation tools.

## Documentation

- `styling-and-markdown.md` -- Complete styling reference with JSON examples for every block type, formatting option, decoration, card layout, divider style, highlight color, and more
- `output-parity.md` -- MCP/API/CLI parity notes + differences chart
- `../command-reference.json` -- Generated `craft schema` command/capability/safety manifest

## CLI Discovery

LLMs can discover styling documentation directly via the CLI:

```bash
craft llm              # Full command reference as JSON
craft llm styles       # Complete styling and formatting guide
craft schema           # Machine-readable command/capability/safety manifest
craft audit agent-dx   # Agent-native CLI scorecard
```

## Backend Selection

- Use REST profiles for deterministic list/get/search/create/update/delete payloads.
- Use MCP profiles for link resolution, block revert metadata, style/theme exploration, collection views, and `ui://craft/edit-review` metadata.
- Use `craft mcp read-resource URI --metadata-only` before requesting full MCP resource contents.
- Use `craft batch --dry-run` to inspect multi-command MCP execution before running writes.
- Use `--fields` with `list` and `search` to keep JSON payloads small, for example `craft list --fields id,title --limit 5`.
- Use MCP for page style flags and reversible block edits; explicit `--backend rest` with MCP-only style flags returns `CAPABILITY_UNAVAILABLE`.
- Use native wrappers instead of generic MCP calls for edit review, image view, collection rename/dynamic properties, and whiteboard element reads.

## Reference Payloads

- `../craft-everything-in-one.json` -- Full JSON snapshot of a document exercising every Craft feature (headings, dividers, code, cards, pages, highlights, rich URLs, decorations, fonts, colors, etc.)

## Key Concepts for LLMs

1. **Everything is a block.** Paragraphs, headings, dividers, images, cards -- all blocks.
2. **Pages and cards contain content.** A `page` or `card` block has a `content[]` array of child blocks.
3. **Styling is composable.** A single block can have `textStyle` + `color` + `font` + `decorations` + `textAlignment` all at once.
4. **Markdown and JSON coexist.** Use markdown shortcuts for quick content, JSON block properties for precise styling control.
