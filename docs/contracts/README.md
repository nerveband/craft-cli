# Craft contract snapshots

Captured data, not hand-maintained API definitions. Review source drift and request fixtures before refreshing snapshots.

| File | Source | Capture |
|---|---|---|
| `craft-rest-space-openapi.json` | [Live Craft Connect docs](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1), embedded OpenAPI 3.1.1 | 2026-10-05, 44 operations across 21 paths |
| `craft-rest-beta-openapi.json` | Historical baseline from the same all-documents source | Not an independently verified beta API; retained only for history |
| `craft-mcp-tools.json` | User-supplied Craft MCP, tools/list | 2026-09-26; three workflow tools |
| `craft-mcp-resources.json` | User-supplied Craft MCP, resources/list | 2026-09-26; edit-review UI resource metadata only |

The older `docs/HHRuPxZZTJ6-docs.md` and exported JSON are absent from this checkout. The live docs embed OpenAPI; no separate stable public spec download URL was verified.

Canonical REST SHA256 (sorted compact JSON): `ff9d9c48136aec25c88675319d4cfc42492dfc26e33a27a6cac170f5624de283`.

The 2026-10-05 refresh changes descriptions only; operation inventory and request/response shapes are unchanged, so encoding fixtures remain unchanged. Collection schema updates replace every column and choice. Even retained columns lose unexposed `isHidden`, `defaultValue`, and select `config.limit` settings. Choice themes retain color only, dropping icons and extra metadata. Omitted choice colors reuse the existing palette color for the same column and exact choice name; unrecognized stored colors fall back to gray. Renaming a choice creates a new identity and does not migrate existing selections.

```bash
python3 scripts/craft_contract.py --check
python3 scripts/craft_contract.py  # refresh after reviewing upstream changes
```

The parser accepts data literals and loader references without executing JavaScript. Scheduled CI fails on any contract change. Default Go tests use fixtures and do not fetch Craft. MCP capture negotiated protocol 2025-06-18 and server craft-workflow-links 1.0.0; this does not claim support for all newer MCP revisions. Do not save private content, API keys or working MCP links here.

`clispec-v0.2.schema.json` is pinned from the [official frozen v0.2 schema](https://clispec.dev/schema/v0.2.json). `scripts/validate_clispec.py` validates the generated command manifest with jsonschema 4.26.0 in CI. CLI Spec 0.3.0 scoring requires the newer v0.3 vocabulary and reports that difference separately.
