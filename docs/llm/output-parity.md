# Output Parity (MCP / API / CLI)

Goal: CLI JSON output matches API/MCP payload shapes by default. Legacy flattened JSON is available via `--format compact`.

## Differences Chart

| Surface | JSON Shape | Notes |
| --- | --- | --- |
| MCP | JSON-RPC envelope | Actual blocks JSON lives in `result.content[].text` (string) |
| API | REST payloads | List endpoints return `{items, total}` |
| CLI (default) | API-shaped JSON | List/search outputs return `{items, total}`; single document returns document JSON |
| CLI (`--format compact`) | Legacy JSON | Flattened arrays for list/search outputs |

## MCP-Backed CLI Output

Commands that wrap MCP-only features return the MCP tool result by default, or
a structured dry-run envelope when `--dry-run` is used. Examples include
`craft documents resolve-link`, `craft blocks revert`, `craft collections views`,
`craft collections active-view set`, and `craft batch`.

## Context-Controlled JSON

`craft list --fields ... --limit ...` and `craft search --fields ... --limit ...`
return projected `{items,total}` payloads. When the CLI truncates a result set
client-side, JSON includes `_metadata.truncated`, `returned`,
`available_in_page`, and a `suggested_command`.

MCP cursor pagination is exposed through `craft list --backend mcp --cursor ...`.

## Payload Snapshots

See `docs/payloads/README.md` for MCP/API/CLI payloads for **Craft Everything In One**.
