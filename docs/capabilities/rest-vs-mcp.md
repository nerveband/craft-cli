# REST vs MCP Capabilities

`craft-cli` supports two Craft automation surfaces:

- **REST API:** direct calls to `https://connect.craft.do/links/.../api/v1`
- **Craft MCP:** streamable HTTP MCP calls to `https://mcp.craft.do/links/.../mcp`

Use REST for deterministic API workflows and bulk scripting. Use MCP when an operation is exposed only through Craft's agent-facing command surface or needs MCP review/revert metadata.

## Current MCP Tools

The public Craft MCP endpoint currently exposes:

| Tool | Purpose |
|---|---|
| `craft_read` | Read/search commands, including folders, documents, blocks, tasks, collections, whiteboards, images, and search |
| `craft_write` | Write commands for documents, blocks, tasks, folders, collections, comments, whiteboards, files, and styling |
| `blocks_revert` | Revert prior block mutations when the blocks have not changed since |

## Prefer REST When

- Listing, reading, searching, creating, updating, moving, deleting, uploading, or managing tasks is fully covered by the REST API.
- You need exact REST request/response contracts.
- You want low-overhead script execution.
- You are running CI or contract tests.

## Prefer MCP When

- You need to resolve a Craft app/web URL to a `rootBlockId`.
- You need page styling features such as themes, fonts, covers, backdrops, separators, or washi.
- You need exploration helpers for icons, themes, washi, or Unsplash images.
- You need richer collection view controls.
- You need reversible block mutation metadata or `blocks_revert`.
- You want to inspect Craft's agent-oriented capability surface.

## Generic MCP Commands

```bash
craft mcp initialize --mcp-url https://mcp.craft.do/links/YOUR_LINK/mcp
craft mcp tools --mcp-url https://mcp.craft.do/links/YOUR_LINK/mcp
craft mcp resources --mcp-url https://mcp.craft.do/links/YOUR_LINK/mcp
craft mcp read-resource ui://craft/edit-review --metadata-only
craft mcp edit-review
craft mcp call craft_read --command "connection info"
craft mcp call craft_write --arguments '{"command":"documents create --title Test"}'
craft batch --command "connection info" --dry-run
```

You can also set:

```bash
export CRAFT_MCP_URL=https://mcp.craft.do/links/YOUR_LINK/mcp
```

## Agent Guidance

When a requested operation is unavailable through REST, agents should:

1. Check `craft mcp tools`.
2. Prefer a native CLI command if one exists.
3. Use `craft mcp call` as a fallback for MCP-only functionality.
4. Explain the missing capability and the needed setup if no MCP URL is configured.

When the active profile is REST and the feature is MCP-only, use this escalation path:

1. Detect MCP-only scope from `craft schema`, `craft profiles capabilities`, or `CAPABILITY_UNAVAILABLE`.
2. Check `craft profiles list` or `craft config list` for an existing `mcp` profile.
3. If none exists, ask the user for a Craft MCP URL or ask them to create one in Craft. The CLI cannot generate a new Craft MCP link by itself.
4. Save it as a separate profile with `craft config add-mcp <name>-mcp --mcp-url URL`.
5. Verify it with `craft profiles test <name>-mcp` and `craft mcp tools --profile <name>-mcp`.
6. Retry the operation with `--profile <name>-mcp` or `--backend mcp`. For writes, use `--dry-run` first, then `--yes --save-revert FILE --diff` when supported.

## Native MCP-Backed Commands

```bash
craft documents resolve-link "https://..."
craft folders explore-icons rocket
craft blocks explore-themes --type page
craft blocks explore-washi
craft blocks search-unsplash "mountain lake"
craft blocks update PAGE_ID --theme-id fire-horse --dry-run
craft blocks add PAGE_ID --markdown "test" --backend mcp --save-revert revert.json --dry-run
craft blocks revert --revert-info revert.json --dry-run
craft list --backend mcp --cursor CURSOR --limit 5
craft collections views list COLLECTION_ID
craft collections views create COLLECTION_ID --name Board --dry-run
craft collections views update COLLECTION_ID --view VIEW_ID --name Backlog --dry-run
craft collections views delete COLLECTION_ID --view VIEW_ID --dry-run
craft collections active-view set COLLECTION_ID --view VIEW_ID --dry-run
craft collections rename COLLECTION_ID --name "New Name" --dry-run
craft collections create --backend mcp --name Tasks --property Status=select --dry-run
craft collections add COLLECTION_ID --backend mcp --property Status=Todo --dry-run
craft whiteboards elements get WHITEBOARD_ID
craft images view "https://example.com/image.png"
```
