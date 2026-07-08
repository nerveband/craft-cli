# Agents

## Quick start

craft-cli is a Go binary for managing Craft.do documents. JSON output by default. 30+ commands covering documents, blocks, folders, tasks, collections, whiteboards, comments, uploads, and search.

## Auth

No interactive login. Set credentials via:
- `craft config add <name> <url>` then `craft config use <name>`
- Or typed profiles: `craft profiles add-rest <name> --api-url URL`, `craft profiles add-mcp <name> --mcp-url URL`
- Use `--profile <name>` to select a REST or MCP profile for one command
- Or per-command: `--api-url URL --api-key KEY`
- For Craft MCP, set `CRAFT_MCP_URL` or pass `--mcp-url URL`

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
- Use `craft batch --dry-run` before `craft batch --tool craft_write --yes`
- MCP-only block style/revert flags such as `--theme-id`, `--cover-url`, `--backdrop-*`, `--washi-*`, `--diff`, and `--save-revert` auto-route to MCP under `--backend auto`; `--backend rest` returns `CAPABILITY_UNAVAILABLE`
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
