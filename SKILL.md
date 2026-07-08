---
name: craft-cli
description: Manage Craft.do documents, blocks, folders, tasks, collections, whiteboards, comments, uploads, search, REST profiles, and Craft MCP connections from a JSON-first CLI.
---

# Craft CLI

Use `craft` when an agent needs to read, create, update, move, search, or delete Craft content from the shell.

## Setup

Use non-interactive auth:

```bash
craft profiles add-rest work --api-url "$CRAFT_API_URL" --api-key-env CRAFT_API_KEY
craft profiles add-mcp work-mcp --mcp-url "$CRAFT_MCP_URL"
craft profiles use work
```

Per command, prefer explicit routing:

```bash
craft list --profile work --limit 5
craft mcp tools --profile work-mcp
```

## Safe Defaults

- Use `craft schema` before unfamiliar commands.
- Use `--json-errors` for machine-readable failures.
- Use `--quiet`, `--id-only`, `--output-only`, `--limit`, `--count`, and `--max-depth` to reduce output.
- Use `--dry-run` before delete, move, clear, or broad updates.
- Use `--yes` only after a dry run or when the operation is already scoped to a temporary test artifact.
- Never print or store API keys. Prefer `--api-key-env`.

## REST Vs MCP

Use REST profiles for deterministic API operations such as list, get, create, update, delete, search, uploads, tasks, folders, collections, comments, and whiteboards.

Use MCP profiles for MCP-only or MCP-ahead operations such as:

- `craft documents resolve-link`
- `craft folders explore-icons`
- `craft blocks explore-themes`
- `craft blocks explore-washi`
- `craft blocks search-unsplash`
- generic `craft mcp tools`, `craft mcp resources`, and `craft mcp call`

If REST reports a capability or permission error, inspect profile metadata:

```bash
craft profiles capabilities work
craft profiles test work
```

## Verification

Before release or major automation, run:

```bash
go test ./...
go vet ./...
craft audit agent-dx --format json
```

The agent-DX score should stay at 85/85 on the v2 audit.
