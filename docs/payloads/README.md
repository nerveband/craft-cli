# Craft payloads: "Craft Everything In One"

Payload snapshots are intentionally **not committed** to git because they may contain private document content.

## Regenerate locally

Document ID: 6DB46B0B-FBE5-4936-88E7-38B91A4F888C

### MCP (JSON-RPC tools/call blocks_get, format=json)

```bash
curl -sS -X POST https://mcp.craft.do/links/548eu6Zqdao/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"blocks_get","arguments":{"id":"6DB46B0B-FBE5-4936-88E7-38B91A4F888C","format":"json"}}}' \
  > docs/payloads/craft-everything-mcp.json
```

### API (documents?documentId=<id>)

```bash
curl -sS "https://connect.craft.do/links/HHRuPxZZTJ6/api/v1/documents?documentId=6DB46B0B-FBE5-4936-88E7-38B91A4F888C" \
  > docs/payloads/craft-everything-api.json
```

### CLI (craft get --format json)

```bash
craft get "6DB46B0B-FBE5-4936-88E7-38B91A4F888C" --format json \
  > docs/payloads/craft-everything-cli.json
```
