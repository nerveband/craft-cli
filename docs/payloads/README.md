# Craft payloads: "Craft Everything In One"

Document ID: 6DB46B0B-FBE5-4936-88E7-38B91A4F888C

## Commands and timestamps (UTC)

- MCP (JSON-RPC tools/call blocks_get, format=json)
  - Timestamp: 2026-01-31T14:54:10Z
  - Command:
    curl -sS -X POST https://mcp.craft.do/links/548eu6Zqdao/mcp \
      -H 'Content-Type: application/json' \
      -H 'Accept: application/json, text/event-stream' \
      -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"blocks_get","arguments":{"id":"6DB46B0B-FBE5-4936-88E7-38B91A4F888C","format":"json"}}}'
  - Output: docs/payloads/craft-everything-mcp.json

- API (documents?documentId=<id>)
  - Timestamp: 2026-01-31T14:54:10Z
  - Command:
    curl -sS "https://connect.craft.do/links/HHRuPxZZTJ6/api/v1/documents?documentId=6DB46B0B-FBE5-4936-88E7-38B91A4F888C"
  - Output: docs/payloads/craft-everything-api.json

- CLI (craft get --format json)
  - Timestamp: 2026-01-31T14:54:10Z
  - Command:
    craft get "6DB46B0B-FBE5-4936-88E7-38B91A4F888C" --format json
  - Output: docs/payloads/craft-everything-cli.json

## Notes
- MCP response is returned as Server-Sent Events (SSE). The JSON payload was extracted from the "data:" line(s).
- MCP tool inventory captured in docs/payloads/craft-everything-mcp-tools-list.json for reference.
