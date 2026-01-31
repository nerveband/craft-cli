# Output Parity (MCP / API / CLI)

Goal: CLI JSON output matches API/MCP payload shapes by default. Legacy flattened JSON is available via `--format compact`.

## Differences Chart

| Surface | JSON Shape | Notes |
| --- | --- | --- |
| MCP | JSON-RPC envelope | Actual blocks JSON lives in `result.content[].text` (string) |
| API | REST payloads | List endpoints return `{items, total}` |
| CLI (default) | API-shaped JSON | List/search outputs return `{items, total}`; single document returns document JSON |
| CLI (`--format compact`) | Legacy JSON | Flattened arrays for list/search outputs |

## Payload Snapshots

See `docs/payloads/README.md` for MCP/API/CLI payloads for **Craft Everything In One**.
