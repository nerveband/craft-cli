# Agents

## Quick start

craft-cli is a Go binary for managing Craft.do documents. JSON output by default. 30+ commands covering documents, blocks, folders, tasks, collections, whiteboards, comments, uploads, and search.

## Auth

No interactive login. Set credentials via:
- `craft config add <name> <url>` then `craft config use <name>`
- Or typed profiles: `craft profiles add-rest <name> --api-url URL`, `craft profiles add-mcp <name> --mcp-url URL`
- Equivalent config aliases exist: `craft config add-rest <name> --api-url URL`, `craft config add-mcp <name> --mcp-url URL`
- Use `--profile <name>` to select a REST or MCP profile for one command
- Or per-command: `--api-url URL --api-key KEY`
- For Craft MCP, set `CRAFT_MCP_URL` or pass `--mcp-url URL`
- MCP is its own profile type. Do not add `mcp_url` to a REST profile or as a top-level config key; it will be ignored. A valid MCP profile has `"type": "mcp"` and `"mcp_url": "https://mcp.craft.do/links/<id>/mcp"`.

## Agent guardrails

- Always use `--dry-run` before delete, move, or clear
- Use `--quiet` to suppress status messages (cleaner for parsing)
- Use `--id-only` or `--output-only <field>` to reduce output tokens
- Use `--limit N` to cap list/search results
- Use `--count` when only a result count is needed
- Use `--max-depth N` on get to control block nesting depth
- Use `--json-errors` for machine-readable errors with hints and retry guidance
- Use `craft schema` to discover commands programmatically (JSON manifest with safety metadata)
- Use `craft audit agent-dx --format json` to check agent-readiness before release
- Use `craft mcp tools` to inspect MCP-only capabilities when REST lacks a feature
- Use `craft mcp read-resource URI --metadata-only` before reading resource contents; MCP UI resources can be large
- Use native wrappers before generic MCP calls: `craft mcp edit-review`, `craft images view`, `craft whiteboards elements get`, `craft collections rename`, and `craft collections --property ...`
- Use `craft batch --dry-run` before `craft batch --tool craft_write --yes`
- MCP-only block style/revert flags such as `--theme-id`, `--cover-url`, `--backdrop-*`, `--washi-*`, `--diff`, and `--save-revert` auto-route to MCP under `--backend auto`; `--backend rest` returns `CAPABILITY_UNAVAILABLE`
- For MCP style writes, run `--dry-run` first, then use `--yes --save-revert FILE --diff` on the real write. `--diff` returns the mutation result plus MCP edit-review metadata when available; `--save-revert` captures the undo payload.
- Use `craft list --backend mcp --cursor CURSOR --limit N` for MCP cursor pagination
- Use `--yes` to skip any confirmation prompts

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad input, missing flag) |
| 2 | API error (network, server, rate limit) |
| 3 | Config error (no profile, bad config) |

## Common pitfalls

- `craft list` returns ALL documents (300+). Use `--limit` or `--folder` to filter.
- `craft get` returns full markdown. For large docs, use `--max-depth 1` to limit nesting.
- `craft delete` is soft-delete (moves to trash). Restore with `craft move ID --location unsorted`.
- Whiteboards require element IDs on add. The API rejects elements without an `id` field.
- The search API caps at 20 results. Use `--limit` for fewer, but you can't get more than 20.
- IDs with path traversals (`../`), query params (`?`), or control characters are rejected.
- `craft get --output` refuses paths outside the current working directory unless `--allow-outside-cwd` is explicit.
- Errors include `(retryable)` or `(not retryable)` in the hint. Only retry on rate limits and server errors.
- REST is best for deterministic direct API operations. MCP may expose richer agent-facing capabilities such as link resolution, themes/covers/backdrops, collection views, and block revert metadata.
- Collection view controls (`collections views ...`, `collections active-view set`) are MCP-backed because the captured REST docs do not expose stable view endpoints.
- Whole-doc `craft update --mode replace` is markdown-based: it clears and reinserts content blocks, so block IDs change and block-only styling/state such as `color`, `font`, `textAlignment`, card layout, task state, media/embed fields, comments, and revert anchors is not preserved. Do not use whole-doc replace on styled documents just to make text edits. Prefer `craft blocks update BLOCK_ID --markdown ...` for existing blocks, which preserves omitted styling fields. If new blocks are created, style only those changed/new blocks.
- `craft update --mode replace --section "Heading"` uses a block-boundary delta replacement: only the target section's top-level block range is deleted/reinserted, preserving IDs and styling outside that section.
- Craft accepts `#RRGGBB` color input but may store adjusted palette/readability colors. Keep original brand hexes as source data and resend those, not the adjusted colors returned by `get`.

## MCP escalation flow

When the user asks for a feature that REST does not expose, such as page themes, page backgrounds, covers, washi, richer collection views, link resolution, edit-review metadata, or reversible style writes:

1. Detect it as MCP-only from `craft schema`, `craft profiles capabilities`, or a `CAPABILITY_UNAVAILABLE` error.
2. Check for an existing MCP profile with `craft profiles list` or `craft config list`.
3. If no MCP profile exists, ask the user for a Craft MCP URL or ask them to create one in Craft. The CLI cannot invent a Craft MCP link by itself.
4. Save it as a separate MCP profile: `craft config add-mcp <name>-mcp --mcp-url https://mcp.craft.do/links/<id>/mcp`.
5. Verify it with `craft profiles test <name>-mcp` and, when relevant, `craft mcp tools --profile <name>-mcp`.
6. Retry the requested operation with `--profile <name>-mcp` or `--backend mcp`. For writes, use `--dry-run` first, then `--yes --save-revert FILE --diff` when supported.

## Avoid destructive restyling loops

Agents should avoid any workflow that deletes and recreates blocks unless the user explicitly wants a structural rewrite. Similar styling/token traps:

- Whole-doc `craft update --mode replace`, `craft clear`, delete/re-add, and broad MCP writes all change block IDs and can invalidate saved IDs, comments, revert payloads, and styling assumptions.
- Markdown exports/imports preserve markdown-native things like headings, lists, quotes, callouts, dividers, and code fences, but not all block JSON fields or MCP page/card styling.
- Page-level styling is on the root page block and may survive content replacement, but content-block styling usually does not. Do not rerun a full styling pass automatically; update the specific blocks that changed.

## Debugging

- **Endpoint:** `https://connect.craft.do/links/HHRuPxZZTJ6/api/v1`
- **Auth:** None required (token disabled for this test endpoint)
- **API docs:** `https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1`
- **Public API docs hub:** `https://connect.craft.do/api-docs`
- **MCP endpoint format:** `https://mcp.craft.do/links/<id>/mcp`

## Live tests

- Default `go test ./...` must not hit Craft; live tests skip unless `CRAFT_LIVE_TESTS=1`
- REST live env: `CRAFT_LIVE_REST_URL` plus optional `CRAFT_LIVE_REST_KEY`
- MCP live env: `CRAFT_LIVE_MCP_URL`
- Write-only state env: `CRAFT_LIVE_WRITEONLY_URL`
- Mutating live tests also require `CRAFT_LIVE_MUTATION_TESTS=1`
- Mutating tests must only create/delete temp artifacts named `craft-cli-live-test-YYYYMMDD-HHMMSS...`
