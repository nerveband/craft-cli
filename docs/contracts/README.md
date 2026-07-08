# Craft Contract Snapshots

Generated/captured contract snapshots used to compare `craft-cli` behavior with
Craft REST and MCP surfaces.

| File | Source | Notes |
|---|---|---|
| `craft-rest-space-openapi.json` | `docs/HHRuPxZZTJ6-openapi.json` | Captured REST OpenAPI for the all-documents test endpoint |
| `craft-rest-beta-openapi.json` | `docs/HHRuPxZZTJ6-openapi.json` | Current beta/space comparison baseline until Craft publishes a distinct beta OpenAPI |
| `craft-mcp-tools.json` | `craft mcp tools --mcp-url https://mcp.craft.do/links/wlYPoWSB9T/mcp` | Public Craft MCP tools snapshot |
| `craft-mcp-resources.json` | `craft mcp resources --mcp-url https://mcp.craft.do/links/wlYPoWSB9T/mcp` | Public Craft MCP resources snapshot |

Do not store private API keys or private endpoint outputs here.
