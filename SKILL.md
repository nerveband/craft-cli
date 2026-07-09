---
name: craft-cli
description: Manage Craft.do documents, blocks, folders, tasks, collections, whiteboards, comments, uploads, search, REST profiles, and Craft MCP connections from a JSON-first CLI.
---

# Craft CLI

Use `craft` when an agent needs to read, create, update, move, search, or delete Craft content from the shell.

## Setup

Use non-interactive auth:

```bash
craft profiles add-rest work --api-url "$CRAFT_API_URL" --api-key-env CRAFT_API_KEY
craft profiles add-mcp work-mcp --mcp-url "$CRAFT_MCP_URL"
craft profiles use work
```

Equivalent config aliases are available:

```bash
craft config add-rest work --api-url "$CRAFT_API_URL" --api-key-env CRAFT_API_KEY
craft config add-mcp work-mcp --mcp-url "$CRAFT_MCP_URL"
```

MCP is a separate profile type. Do not add `mcp_url` to a REST profile or as a top-level config key. A valid MCP profile has `"type": "mcp"` and `"mcp_url"` and is best created with `craft profiles add-mcp` or `craft config add-mcp`.

Per command, prefer explicit routing:

```bash
craft list --profile work --limit 5
craft mcp tools --profile work-mcp
```

## Safe Defaults

- Use `craft schema` before unfamiliar commands.
- Use `--json-errors` for machine-readable failures.
- Use `--quiet`, `--id-only`, `--output-only`, `--limit`, `--count`, and `--max-depth` to reduce output.
- Use `--dry-run` before delete, move, clear, or broad updates.
- Use `--yes` only after a dry run or when the operation is already scoped to a temporary test artifact.
- Never print or store API keys. Prefer `--api-key-env`.

## REST Vs MCP

Use REST profiles for deterministic API operations such as list, get, create, update, delete, search, uploads, tasks, folders, collections, comments, and whiteboards.

Use MCP profiles for MCP-only or MCP-ahead operations such as:

- `craft documents resolve-link`
- `craft folders explore-icons`
- `craft blocks explore-themes`
- `craft blocks explore-washi`
- `craft blocks search-unsplash`
- generic `craft mcp tools`, `craft mcp resources`, and `craft mcp call`

For MCP style writes, use a reversible verification flow:

```bash
craft blocks update PAGE_ID --theme-id soil-and-clay --backend mcp --profile work-mcp --dry-run
craft blocks update PAGE_ID --theme-id soil-and-clay --backend mcp --profile work-mcp --yes --save-revert revert.json --diff
craft blocks revert --revert-info-file revert.json --profile work-mcp --dry-run
```

`--diff` returns the mutation result plus MCP edit-review metadata when available. REST may not expose page-level theme/background properties, so verify MCP-only styling through MCP responses, edit-review metadata, and saved revert payloads.

If REST reports a capability or permission error, inspect profile metadata:

```bash
craft profiles capabilities work
craft profiles test work
```

## MCP Escalation Flow

When a user asks for page themes, page backgrounds, covers, washi, richer collection views, link resolution, edit-review metadata, reversible style writes, or another MCP-only feature, do not keep trying REST. Do this:

1. Confirm the feature is MCP-only with `craft schema`, `craft profiles capabilities`, or the `CAPABILITY_UNAVAILABLE` error.
2. Check for an existing MCP profile with `craft profiles list` or `craft config list`.
3. If none exists, ask the user for a Craft MCP URL or ask them to create one in Craft. The CLI cannot generate a new Craft MCP link on its own.
4. Save it as a separate profile: `craft config add-mcp <name>-mcp --mcp-url "$CRAFT_MCP_URL"`.
5. Verify it: `craft profiles test <name>-mcp` and, if useful, `craft mcp tools --profile <name>-mcp`.
6. Retry the requested operation with `--profile <name>-mcp` or `--backend mcp`. For writes, dry-run first, then use `--yes --save-revert FILE --diff` when supported.

## Styling And Replacement

Whole-doc `craft update --mode replace` is markdown-based. It clears and reinserts content blocks, so block IDs change and block-only styling/state such as `color`, `font`, `textAlignment`, card layout, task state, media/embed fields, comments, and revert anchors is not preserved.

Do not use whole-doc replace on styled documents just to make text edits. Prefer `craft blocks update BLOCK_ID --markdown ...` for existing blocks; omitted fields are preserved on that block. If a content edit creates new blocks, style only the changed/new blocks rather than rerunning a full styling pass.

`craft update --mode replace --section "Heading"` uses a block-boundary delta replacement: only that section's top-level block range is deleted/reinserted, preserving IDs and styling outside the section. Styling inside the replaced section can still be lost because those section blocks are recreated.

Similar destructive restyling traps: `craft clear`, delete/re-add, whole-doc markdown replacement, and broad MCP writes can invalidate saved block IDs, comments, revert payloads, and styling assumptions. Markdown round trips preserve markdown-native structures such as headings, lists, quotes, callouts, dividers, and code fences, but not all block JSON fields or MCP page/card styling. Page-level styling may survive because it is attached to the root page block, but content-block styling generally does not.

Craft accepts `#RRGGBB` color input but may store adjusted palette/readability colors. Keep original brand hexes as source data and resend those. Do not reuse Craft-returned adjusted colors as the source of truth.

## Verification

Before release or major automation, run:

```bash
go test ./...
go vet ./...
craft audit agent-dx --format json
```

The agent-DX score should stay at 85/85 on the v2 audit.
