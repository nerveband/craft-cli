# Local payload comparisons

Private document snapshots must stay outside source control. Use existing profiles and an explicitly selected document. Do not paste working link identifiers or API keys into scripts or reports.

```bash
craft get DOCUMENT_ID --profile work --format structured --deliver file:document-rest.json
craft mcp call craft_read --command "blocks get --id DOCUMENT_ID --format json" --profile work-mcp > document-mcp.json
```

REST document content comes from `GET /blocks?id=DOCUMENT_ID`, not `GET /documents?documentId=...`. Craft MCP exposes `craft_read` with a command string, not a `blocks_get` tool. Confirm current command options through `craft mcp call craft_read --command "blocks get --help"` before fetching private content. Use fixture data for tests.
