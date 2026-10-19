# Craft contract snapshots

Captured data, not hand-maintained API definitions. Review upstream drift and update request fixtures before refreshing a snapshot. Record what changed in the refresh commit message.

| File | Source | Capture |
|---|---|---|
| `craft-rest-space-openapi.json` | [Live Craft Connect docs](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1), embedded OpenAPI 3.1.1 | 2026-10-19, 44 operations across 21 paths |
| `craft-mcp-tools.json` | User-supplied Craft MCP, tools/list | 2026-09-26; three workflow tools |
| `craft-mcp-resources.json` | User-supplied Craft MCP, resources/list | 2026-09-26; edit-review UI resource metadata only |
| `clispec-v0.2.schema.json` | [Official frozen CLI Spec v0.2 schema](https://clispec.dev/schema/v0.2.json) | Validates `docs/command-reference.json` in CI |

```bash
python3 scripts/craft_contract.py --check   # exits 1 on drift, prints the canonical SHA256
python3 scripts/craft_contract.py           # refresh after reviewing upstream changes
```

The live docs embed OpenAPI; no separate stable public spec download URL was verified. The parser reads data literals and loader references without executing JavaScript. Scheduled CI fails on any contract change. Default Go tests use fixtures and do not fetch Craft.

MCP capture negotiated protocol 2025-06-18 and server craft-workflow-links 1.0.0; this does not claim support for newer MCP revisions. Do not save private content, API keys or working MCP links here.
