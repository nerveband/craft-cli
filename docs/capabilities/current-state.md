# Craft CLI Capability Current State

This is the implementation-facing capability map for the REST + MCP plan. For
user-facing backend selection guidance, see `docs/capabilities/rest-vs-mcp.md`.

## REST

REST is the default backend for deterministic API operations:

- documents: list, get, create, update, move, delete
- blocks: get, add, update, move, delete, clear
- folders: list, create, move, delete
- tasks: list, add, update, delete
- collections: list, schema, raw create, raw schema update, items add/update/delete
- comments: add
- upload: file upload and raw upload metadata
- whiteboards: create, add, update, delete
- search: document/block search with limit/count controls
- context controls: `--limit`, `--count`, `--fields`, `--max-depth`, `--output-only`, and truncation metadata on JSON list/search payloads
- hardening: document/block/folder IDs reject path traversal, query fragments, encoded path characters, and control characters; output files stay under cwd unless explicitly overridden

## MCP

MCP is used for agent-ahead or MCP-only features:

- generic MCP: initialize, tools, resources, read-resource, call
- batch: top-level `craft batch` and `craft mcp batch`
- documents: resolve-link
- folders: explore-icons
- blocks: explore-themes, explore-washi, search-unsplash, revert
- block styling: MCP-backed `--theme-id`, `--cover-url`, `--backdrop-*`, `--washi-*`, `--save-revert`, and `--diff` on block add/update
- cursor pagination: MCP-backed `craft list --backend mcp --cursor ... --limit ...`
- collections: views list/create/update/delete, active-view set
- resource metadata: `ui://craft/edit-review` via `--metadata-only`

## API States

Profiles model:

- backend type: `rest` or `mcp`
- access mode: `public` or `api-key`
- permission: `read-only`, `write-only`, or `read-write`
- scope: `all-documents`, `selected-documents`, `daily-notes`, or `connection-defined`

Live tests are opt-in through `CRAFT_LIVE_TESTS=1`; mutations require
`CRAFT_LIVE_MUTATION_TESTS=1` and temporary artifact names starting with
`craft-cli-live-test-`.
