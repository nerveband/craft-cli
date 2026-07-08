# Craft CLI REST + MCP Agent-First Implementation Plan

**Date:** 2026-07-08  
**Status:** Implemented and fully verified  
**Goal:** Make `craft-cli` an agent-first Craft automation surface that handles multiple Craft API connection states, exposes REST and MCP-only capabilities clearly, and gives agents enough runtime metadata to choose the right backend for each operation.

## Implementation Verification

Current verification result:

```bash
go test ./...
go test ./... -run Live
go vet ./...
go build -o /tmp/craft-cli-live-tests-check .
/tmp/craft-cli-live-tests-check audit agent-dx --format json
```

Public live verification has also been run with:

```bash
CRAFT_LIVE_TESTS=1 \
CRAFT_LIVE_REST_URL=https://connect.craft.do/links/HHRuPxZZTJ6/api/v1 \
CRAFT_LIVE_MCP_URL=https://mcp.craft.do/links/wlYPoWSB9T/mcp \
CRAFT_LIVE_WRITEONLY_URL=https://connect.craft.do/links/Dfl9gELEXWY/api/v1 \
go test ./... -run Live
```

The built-in audit reports:

```json
{
  "score": 85,
  "max": 85,
  "grade": "agent-native",
  "failures": []
}
```

Live tests skip unless `CRAFT_LIVE_TESTS=1` is set. Mutating live tests also require `CRAFT_LIVE_MUTATION_TESTS=1` and only create temporary artifacts with the `craft-cli-live-test-YYYYMMDD-HHMMSS` prefix.

Keyed live mutation execution has been verified with the approved test fixture
through environment variables only. Do not inline API keys in docs, shell
history, committed files, or test logs.

```bash
CRAFT_LIVE_TESTS=1 CRAFT_LIVE_MUTATION_TESTS=1 \
CRAFT_LIVE_REST_URL=https://connect.craft.do/links/5VruASgpXo0/api/v1 \
CRAFT_LIVE_REST_KEY=<redacted> \
go test ./internal/api -run LiveRESTMutation
```

## Executive Summary

Craft now exposes overlapping but non-identical surfaces:

- **Craft Connect REST API** for direct, scriptable HTTP operations.
- **Craft MCP** for richer agent-oriented commands, reversible block edits, style/theme helpers, and MCP Apps review UI metadata.
- **Craft app API settings** that vary by document scope, permission level, and access mode.

`craft-cli` should not pretend all connections are equal. It should model the active connection, detect capabilities, and tell users and agents when an operation requires a different setup.

Example target behavior:

```bash
craft blocks update PAGE_ID --theme-id fire-horse
```

If the active REST profile cannot perform that style operation but an MCP profile can, the CLI should return:

```json
{
  "error": "operation requires MCP capabilities",
  "code": "CAPABILITY_UNAVAILABLE",
  "required": ["mcp", "blocks.style"],
  "hint": "Configure an MCP profile with `craft profiles add-mcp ...` or use a REST-supported block update."
}
```

## Inputs And Test Fixtures

Use these as live test fixtures. Do not store secrets in source-controlled files.

| Fixture | Type | Access | Scope | Purpose |
|---|---|---|---|---|
| `keyed-all-docs-rw` | REST | API key | All Documents, Read and Write | Full read/write/delete live test fixture |
| `public-selected-writeonly` | REST | Public URL | Selected Documents, Write Only | Permission and write-only behavior fixture |
| `public-mcp` | MCP | Public URL | Craft MCP server | MCP parity and MCP-only capability fixture |
| `beta-all-docs-noauth` | REST docs/test endpoint | Public/no key | All Documents test endpoint | Beta REST docs and endpoint behavior comparison |

Live mutation testing is approved for temporary documents/tasks. Test artifacts should use this prefix:

```text
craft-cli-live-test-YYYYMMDD-HHMMSS
```

Live destructive tests must clean up after themselves and must be gated behind:

```bash
CRAFT_LIVE_TESTS=1
CRAFT_LIVE_MUTATION_TESTS=1
```

## Current Findings

### REST API Drift

Known local client drift against the live beta docs:

- `MoveDocument` should call `PUT /documents/move`, not `PUT /documents`.
- `MoveFolder` should call `PUT /folders/move`, not `PUT /folders`.
- Task payloads appear stale:
  - add should use nested `taskInfo` and `location`;
  - update should use `tasksToUpdate`;
  - delete should use `idsToDelete`.
- `craft schema --command collections` hides nested `collections schema` because the schema generator excludes every command named `schema`.
- Root help points to an outdated Craft API docs URL.

### MCP Surface

The public Craft MCP endpoint at `https://mcp.craft.do/links/wlYPoWSB9T/mcp` exposes:

- `craft_read`
- `craft_write`
- `blocks_revert`
- resource `ui://craft/edit-review`

This is not the older one-tool-per-operation MCP shape. It is a command mini-language behind two primary tools plus a revert tool.

### MCP-Only Or MCP-Ahead Capabilities

MCP currently exposes capabilities that are missing or only partially represented in the CLI/REST client:

- `documents resolve-link <url>` to convert Craft app/web links to `rootBlockId`.
- Cursor pagination for document listing.
- Exploration helpers:
  - `folders explore-icons`
  - `blocks explore-themes --type page|fonts|code`
  - `blocks explore-washi`
  - `blocks search-unsplash`
- Page/block style write flags:
  - `--theme-id`
  - `--font`
  - `--text-color`
  - `--bg-color`
  - `--theme-color`
  - `--cover-url`
  - `--cover-crop`
  - `--cover-attribution`
  - `--backdrop-type`
  - `--backdrop-color`
  - `--backdrop-colors`
  - `--backdrop-direction`
  - `--backdrop-url`
  - `--separator`
  - `--washi-pattern`
  - `--washi-color`
- Collection creation with property definition flags.
- Collection views:
  - list
  - create
  - update
  - delete
  - set active
- Batch command execution via semicolon-separated commands.
- Reversible block mutations via `blocks_revert`.
- Edit-review UI metadata that can be translated into CLI dry-run/diff/revert output.

## Architecture Principles

1. **Connection state is explicit.** Every profile has scope, access mode, permission, and capability metadata.
2. **REST and MCP are peers.** REST is faster and simpler for direct API calls; MCP is richer for agent-style operations and some styling/revert flows.
3. **Agents get guidance, not mystery failures.** Unsupported operations return structured errors with exact setup hints.
4. **Raw API/MCP payloads are first-class.** Convenience flags wrap raw payloads; they do not replace them.
5. **Testing includes grading.** The CLI best-practices grading system becomes a runnable audit and CI signal.
6. **Mutations are reversible when possible.** Mutating commands emit enough metadata for dry-run, review, and revert.

## Profile And Capability Model

### Profile Types

Support both REST and MCP profiles.

```json
{
  "name": "keyed-all-docs-rw",
  "type": "rest",
  "api_url": "https://connect.craft.do/links/.../api/v1",
  "api_key_env": "CRAFT_KEYED_ALL_DOCS_KEY",
  "access_mode": "api_key",
  "permission": "read_write",
  "document_scope": "all_documents",
  "docs_url": "https://connect.craft.do/api-docs/space",
  "capabilities": {
    "read": true,
    "write": true,
    "delete": true,
    "collections": true,
    "whiteboards": true,
    "tasks": true,
    "upload": true,
    "mcp": false
  }
}
```

```json
{
  "name": "public-mcp",
  "type": "mcp",
  "mcp_url": "https://mcp.craft.do/links/.../mcp",
  "access_mode": "public",
  "permission": "read_write",
  "document_scope": "connection_defined",
  "capabilities": {
    "read": true,
    "write": true,
    "blocks.revert": true,
    "blocks.style": true,
    "documents.resolve_link": true,
    "explore.icons": true,
    "explore.themes": true,
    "explore.unsplash": true,
    "collections.views": true,
    "mcp": true
  }
}
```

### Profile Commands

Add a first-class `profiles` command group while preserving existing `config` commands as compatibility aliases.

```bash
craft profiles add-rest NAME --api-url URL [--api-key KEY | --api-key-env ENV] \
  [--scope all-documents|selected-documents|daily-notes] \
  [--permission read-only|write-only|read-write] \
  [--access public|api-key]

craft profiles add-mcp NAME --mcp-url URL \
  [--scope connection-defined] \
  [--permission read-only|write-only|read-write]

craft profiles list
craft profiles show NAME
craft profiles use NAME
craft profiles test NAME
craft profiles test --all
craft profiles capabilities NAME
craft profiles remove NAME --dry-run
```

### Precedence

Profile/config resolution order:

1. Explicit flags: `--api-url`, `--api-key`, `--mcp-url`.
2. Explicit profile: `--profile NAME`.
3. Environment: `CRAFT_PROFILE`, `CRAFT_API_URL`, `CRAFT_API_KEY`, `CRAFT_MCP_URL`.
4. Active profile from config.
5. Legacy config active URL.

## Capability Routing

Each command should declare required capabilities in `craft schema`.

Examples:

```json
{
  "command": "blocks update",
  "required_capabilities": ["write", "blocks.update"],
  "optional_capabilities": ["blocks.style", "blocks.revert"],
  "backends": ["rest", "mcp"]
}
```

```json
{
  "command": "documents resolve-link",
  "required_capabilities": ["documents.resolve_link"],
  "backends": ["mcp"]
}
```

Routing policy:

- Prefer REST for operations fully supported by REST.
- Prefer MCP for MCP-only capabilities.
- Allow explicit backend selection:

```bash
craft blocks update ID --markdown "Text" --backend rest
craft blocks update ID --theme-id fire-horse --backend mcp
```

If backend is `auto`, select the first active profile/backend that satisfies the declared capabilities.

## Implementation Phases

## Phase 0: Baseline, Safety, And Contract Capture

**Goal:** Establish a clean baseline and capture current docs/MCP capabilities without mutating user data.

Tasks:

- Run `go test ./...`.
- Capture live REST docs pages and embedded OpenAPI where available.
- Capture MCP `initialize`, `tools/list`, `resources/list`.
- Add `docs/contracts/` for generated or captured contracts:
  - `craft-rest-space-openapi.json`
  - `craft-rest-beta-openapi.json`
  - `craft-mcp-tools.json`
  - `craft-mcp-resources.json`
- Add `docs/capabilities/current-state.md` summarizing REST vs MCP coverage.

Verification:

```bash
go test ./...
craft schema --commands-only
```

Deliverables:

- Contract snapshots.
- Capability map.
- No functional code changes unless required for snapshot tooling.

## Phase 1: REST Drift Fixes

**Goal:** Fix known API mismatches and keep existing CLI behavior stable.

Tasks:

- Change `MoveDocument` endpoint to `PUT /documents/move`.
- Change `MoveFolder` endpoint to `PUT /folders/move`.
- Update folder/document move request payloads to match docs.
- Update task add/update/delete payloads:
  - `taskInfo`
  - `location`
  - `tasksToUpdate`
  - `idsToDelete`
- Add `tasks list --scope all`.
- Add `tasks add --location dailyNote --date today|tomorrow|yesterday|YYYY-MM-DD`.
- Fix `craft schema` nested `schema` filtering.
- Update root help links.

Verification:

```bash
go test ./...
go run . schema --command collections
go run . move DOC_ID --folder FOLDER_ID --dry-run --json-errors
go run . tasks add "test" --location dailyNote --date today --dry-run
```

Deliverables:

- Unit tests for endpoint paths and payload shapes.
- Schema snapshot tests for nested commands.

## Phase 2: Profile System And Multi-Connection UX

**Goal:** Make multiple Craft connections first-class and scriptable.

Tasks:

- Add profile model with REST/MCP profile types.
- Add `profiles` command group.
- Preserve `config add/use/list/remove/reset` compatibility.
- Add `--profile` global flag.
- Add `CRAFT_PROFILE` support.
- Add `CRAFT_MCP_URL` support.
- Add profile test/capability detection.
- Add config migration from legacy profiles.

Verification:

```bash
craft profiles add-rest keyed-all-docs-rw --api-url "$CRAFT_RW_URL" --api-key-env CRAFT_RW_KEY
craft profiles add-rest public-selected-writeonly --api-url "$CRAFT_WRITEONLY_URL" --access public --permission write-only
craft profiles add-mcp public-mcp --mcp-url "$CRAFT_MCP_URL"
craft profiles list --format json
craft profiles test --all --format json
```

Deliverables:

- Multiple profiles can coexist.
- Commands can run against a specific profile.
- Capability results are machine-readable.

## Phase 3: Permission And API State Handling

**Goal:** Handle Read Only, Write Only, Read and Write, Public, API Key, All Documents, and Selected Documents states cleanly.

Tasks:

- Add command capability declarations.
- Add permission preflight where profile metadata is known.
- Add runtime classification for `401`, `403`, `404`, `405`, and scope errors.
- Add structured errors:
  - `AUTH_REQUIRED`
  - `PERMISSION_DENIED`
  - `READ_UNAVAILABLE`
  - `WRITE_UNAVAILABLE`
  - `SCOPE_UNAVAILABLE`
  - `CAPABILITY_UNAVAILABLE`
  - `BACKEND_REQUIRED`
- Add hints that recommend profile/backend changes.

Examples:

```json
{
  "error": "read operation not available for this connection",
  "code": "READ_UNAVAILABLE",
  "profile": "public-selected-writeonly",
  "hint": "This profile appears write-only. Use create/update commands or switch to a read-capable profile."
}
```

Verification:

```bash
craft list --profile public-selected-writeonly --json-errors
craft create --profile public-selected-writeonly --title "craft-cli-live-test" --dry-run
craft connection --profile keyed-all-docs-rw
craft connection --profile public-mcp
```

Deliverables:

- Agents can distinguish bad credentials from insufficient permissions from unsupported backend.

## Phase 4: REST Endpoint Parity

**Goal:** Implement missing REST endpoints and flags from current regular/beta docs.

Tasks:

- Collections:
  - `collections create`
  - `collections schema update`
  - `collections views list`
  - `collections views create`
  - `collections views update`
  - `collections views delete`
  - `collections active-view set`
- Documents:
  - cursor/pagination if available via REST.
  - date filters currently exposed by docs.
  - selected-doc scope behavior documentation.
- Search:
  - `fetchBlocks`
  - `folderIds`
  - `documentIds`
  - daily note date filters.
- Tasks:
  - move task via location updates if supported.
  - update markdown.
  - repeat config if accepted by API.
- Upload:
  - full placement support for `pageId`, `date`, `siblingId`, `position`.
- Whiteboards:
  - appState/assets preservation where supported.

Verification:

```bash
go test ./...
craft collections create --dry-run ...
craft collections views create --dry-run ...
craft search "test" --fetch-blocks --limit 5
```

Deliverables:

- REST feature parity table updated.
- All new endpoints covered by endpoint/payload tests.

## Phase 5: MCP Client Backend

**Goal:** Allow `craft-cli` to call Craft MCP directly for MCP-only features.

Tasks:

- Add minimal streamable HTTP MCP client:
  - `initialize`
  - `tools/list`
  - `resources/list`
  - `tools/call`
  - `resources/read` only for metadata/inspection, not for normal huge UI dumps.
- Add `internal/mcp` package or equivalent.
- Add MCP profile support.
- Add `craft mcp tools`.
- Add `craft mcp call TOOL --arguments JSON`.
- Add `craft mcp read-resource URI --metadata-only`.

Verification:

```bash
craft mcp tools --profile public-mcp
craft mcp call craft_read --arguments '{"command":"connection info"}' --profile public-mcp
craft mcp call craft_read --arguments '{"command":"documents list --help"}' --profile public-mcp
```

Deliverables:

- Generic MCP invocation available for advanced users and debugging.
- MCP tool metadata stored in capabilities.

## Phase 6: MCP Feature Integration Into Native Commands

**Goal:** Expose high-value MCP-only/ahead features as normal CLI commands.

Tasks:

- Documents:
  - `documents resolve-link <url>`
  - or root alias `resolve-link <url>`.
- Folders:
  - `folders explore-icons <term>`.
- Blocks:
  - `blocks explore-themes --type page|fonts|code`
  - `blocks explore-washi`
  - `blocks search-unsplash <query>`
  - style flags on `blocks add`
  - style flags on `blocks update`
- Collections:
  - MCP-compatible property definition flags for `collections create`.
  - full view flags for create/update.
- Batch:
  - `craft batch --backend rest|mcp|auto --file ops.json`
  - support NDJSON stdin.
  - optionally support semicolon command strings for MCP parity.
- Revert:
  - mutation commands can emit `revertInfo`.
  - `craft blocks revert --revert-info FILE_OR_JSON`.
  - `--save-revert path` on mutating block commands.

Verification:

```bash
craft documents resolve-link "https://..."
craft folders explore-icons rocket --profile public-mcp
craft blocks explore-themes --type page --profile public-mcp
craft blocks update PAGE_ID --theme-id fire-horse --dry-run --backend mcp
craft blocks revert --revert-info revert.json --dry-run
```

Deliverables:

- Users do not have to know MCP internals for common MCP-only features.
- Agents still get clear backend/capability metadata.

## Phase 7: Raw Payload Input Everywhere

**Goal:** Align with the CLI best-practices requirement that mutating commands accept raw structured payloads.

Tasks:

- Add `--json` and `--stdin` to all mutating commands:
  - create
  - update
  - delete
  - clear
  - move
  - blocks add/update/delete/move
  - folders create/move/delete
  - tasks add/update/delete
  - collections create/schema/items/views
  - comments add
  - upload metadata where applicable
  - whiteboards create/add/update/delete
- Raw REST payloads should match REST docs.
- Raw MCP payloads should match MCP command arguments.
- Add local validation before API calls.

Verification:

```bash
echo '{"documents":[{"title":"Test"}]}' | craft create --stdin --dry-run
craft tasks add --json '{"tasks":[...]}' --dry-run
craft blocks update --json '[{"id":"...","markdown":"..."}]' --dry-run
```

Deliverables:

- Agents can use API docs directly without bespoke flag translation.

## Phase 8: Structured Dry-Run, Diff, And Revert

**Goal:** Make mutating commands reviewable and reversible where possible.

Tasks:

- Make dry-run JSON structured:

```json
{
  "dry_run": true,
  "operation": "blocks.update",
  "backend": "mcp",
  "profile": "public-mcp",
  "request": {
    "method": "tools/call",
    "tool": "craft_write",
    "arguments": {
      "command": "blocks update --id ..."
    }
  },
  "capabilities": ["write", "blocks.style"]
}
```

- For real MCP mutations, capture structured result when available.
- Save `revertInfo` when available.
- Add `--save-revert`.
- Add `--diff` for operations that can preview existing and new block content.
- Add `craft history` only if local persistence is useful; otherwise keep file-based revert.

Verification:

```bash
craft blocks add --id PAGE_ID --markdown "test" --backend mcp --save-revert revert.json
craft blocks revert --revert-info revert.json --dry-run
```

Deliverables:

- CLI equivalent of MCP edit-review safety without requiring a graphical MCP App.

## Phase 9: Context Window Discipline

**Goal:** Prevent accidental huge outputs and improve agent ergonomics.

Tasks:

- Add `--fields id,title,lastModifiedAt`.
- Add cursor/page support where available.
- Add response metadata:

```json
{
  "items": [],
  "_metadata": {
    "truncated": true,
    "returned": 20,
    "next_cursor": "...",
    "suggested_command": "craft list --cursor ..."
  }
}
```

- Add consistent `--limit`.
- Add `--max-depth` aliases where relevant.
- Add `--full` for raw complete payloads when compact output is default.

Verification:

```bash
craft list --fields id,title --limit 5
craft get DOC_ID --max-depth 1
craft search "api" --limit 5 --fields documentId,markdown
```

Deliverables:

- Agent-safe output defaults.

## Phase 10: Input Hardening

**Goal:** Defend against common agent mistakes and unsafe IDs/paths.

Tasks:

- Reject IDs with:
  - `../`
  - query strings
  - fragments
  - control characters
  - malformed percent encoding
  - obvious URL passed where block ID is expected.
- Add helpful error suggesting `documents resolve-link`.
- Validate enums locally.
- Validate dates locally.
- Validate mutually exclusive flags.
- Validate output paths stay under cwd unless `--allow-outside-cwd`.
- Sanitize API-originated text in any agent-facing guidance.

Verification:

```bash
craft get 'https://www.craft.do/s/...' --json-errors
craft get '../bad' --json-errors
craft tasks add "x" --location dailyNote --date not-a-date --json-errors
```

Deliverables:

- Better self-correction for agents.

## Phase 11: Agent DX Grading System

**Goal:** Use the CLI best-practices grading system as a real test/audit.

Tasks:

- Convert `prompts/agent-cli-audit.md` into a runnable script or Go command.
- Add:

```bash
craft audit agent-dx --format json
```

or:

```bash
scripts/agent-dx-audit.sh ./craft
```

- Score 50 checks across:
  - discoverability
  - structured output
  - input flexibility
  - safety rails
  - error handling
  - context control
  - idempotency
  - auth/profile handling
  - docs/examples
  - agent packaging
- Update stale `prompts/craft-cli-agent-dx-score.md` to reflect that `craft schema` now exists.
- Add CI threshold:
  - minimum before merge: no regression
  - target: 36+/50 Agent-first
  - stretch: 46+/50 Best-in-class

Example output:

```json
{
  "score": 41,
  "max": 50,
  "grade": "agent-first",
  "failures": [
    {
      "check": "raw payload passthrough",
      "command": "craft folders create",
      "fix": "add --json/--stdin"
    }
  ]
}
```

Verification:

```bash
craft audit agent-dx --format json
```

Deliverables:

- Grading becomes part of the development loop, not a static note.

## Phase 12: Live Test Matrix

**Goal:** Verify real API behavior across the states users can configure in Craft.

### REST Keyed All Documents Read/Write

Test:

- `connection`
- `list`
- `get`
- `search`
- create temp doc
- add block
- update block
- create task
- update task
- delete temp task/doc
- collection create/item/view if available
- upload small test asset if safe

### REST Public Selected Documents Write Only

Test:

- read commands fail cleanly or return limited data.
- write commands work only within selected scope.
- unsupported operations return `READ_UNAVAILABLE`, `SCOPE_UNAVAILABLE`, or `CAPABILITY_UNAVAILABLE`.
- dry-run works without mutation.

### MCP Public

Test:

- initialize
- tools/list
- resources/list
- `craft_read connection info`
- `craft_read documents list --help`
- `craft_read blocks explore-themes --help`
- `craft_write documents create --help`
- actual temp document create/delete after mutation gate is enabled.
- block mutation with revert info.
- `blocks_revert` dry-run/real revert if possible.

Verification command:

```bash
CRAFT_LIVE_TESTS=1 CRAFT_LIVE_MUTATION_TESTS=1 go test ./... -run Live
```

Deliverables:

- Live tests document actual Craft behavior rather than assumptions.
- Default local test runs do not call Craft and safely skip live coverage.
- REST keyed read/write, REST write-only state, and MCP read-matrix tests are represented in `internal/api/live_test.go` and `internal/mcp/live_test.go`.

## Phase 13: Documentation And Agent Packaging

**Goal:** Make user and agent setup decisions obvious.

Tasks:

- Update `AGENTS.md`:
  - current REST docs links
  - current MCP endpoint behavior
  - profile usage
  - when to use REST vs MCP
  - always use `--dry-run` before mutations unless explicitly allowed
- Update `README.md`:
  - multi-profile setup
  - REST vs MCP comparison
  - permission modes
  - access modes
  - selected-doc limitations
  - write-only behavior
- Update `docs/llm/*`.
- Add generated command reference from `craft schema`.
- Add `docs/capabilities/rest-vs-mcp.md`.

Example docs table:

| Need | Recommended Backend | Why |
|---|---|---|
| Fast list/get/search | REST | Direct API, low overhead |
| Resolve Craft URL to block ID | MCP | MCP exposes `documents resolve-link` |
| Page themes/covers/backdrops | MCP | MCP exposes style command flags |
| Reversible block edits | MCP | MCP exposes `blocks_revert` |
| Bulk raw API payloads | REST | Best for deterministic scripts |
| Agent review UI | MCP | MCP App resource exposes edit-review UI |

Deliverables:

- Users can configure the correct connection for the work they want.
- Agents can explain missing capabilities and setup steps.

## Phase 14: Release Hardening

**Goal:** Ship without regressing current users.

Tasks:

- Keep backward-compatible aliases for existing commands.
- Add deprecation warnings only when necessary and suppress under `--quiet`.
- Ensure old `config` commands still work.
- Add changelog with migration notes.
- Add release checklist:
  - `go test ./...`
  - agent DX audit
  - live read tests
  - live mutation tests on approved fixtures
  - docs generated
  - contract snapshots updated

Deliverables:

- Stable release with clear migration path.

## Command Surface Target

### Profiles

```bash
craft profiles add-rest
craft profiles add-mcp
craft profiles list
craft profiles show
craft profiles use
craft profiles test
craft profiles capabilities
craft profiles remove
```

### MCP

```bash
craft mcp tools
craft mcp resources
craft mcp call
craft mcp read-resource
```

### Documents

```bash
craft documents resolve-link
craft list
craft get
craft create
craft update
craft delete
craft move
```

### Blocks

```bash
craft blocks get
craft blocks add
craft blocks update
craft blocks delete
craft blocks move
craft blocks revert
craft blocks explore-themes
craft blocks explore-washi
craft blocks search-unsplash
```

### Collections

```bash
craft collections list
craft collections create
craft collections schema
craft collections schema update
craft collections items
craft collections add
craft collections update
craft collections delete
craft collections views list
craft collections views create
craft collections views update
craft collections views delete
craft collections active-view set
```

### Batch And Audit

```bash
craft batch
craft audit agent-dx
```

## REST Vs MCP Decision Rules

Use REST when:

- The operation is fully covered by REST.
- The user wants deterministic raw payloads.
- The task is high-volume listing/searching.
- The agent needs minimal overhead.

Use MCP when:

- The operation needs link resolution.
- The operation needs page theme, cover, backdrop, washi, or theme exploration.
- The operation needs reversible block edits.
- The operation needs collection view controls ahead of REST implementation.
- The operation benefits from MCP App edit-review metadata.
- REST returns capability errors but MCP profile has the feature.

Use both when:

- Verifying parity.
- Performing create/read/update/revert workflows.
- Capturing live behavior for contracts.

## Risks And Mitigations

| Risk | Mitigation |
|---|---|
| MCP command language changes | Capture `tools/list` snapshots and expose generic `craft mcp call` fallback |
| REST docs and MCP diverge | Maintain separate capability maps and test both |
| Write-only APIs confuse agents | Permission-aware errors and profile capability tests |
| Live tests mutate real data | Gate with env vars, use temp prefix, cleanup in defer/finalizers |
| Secrets leak into repo | Use env vars only, redact key values in logs |
| Raw payload support bypasses validation | Validate dangerous IDs/paths and mutually exclusive targeting before sending |
| Output becomes too large | `--fields`, `--limit`, cursor metadata, truncation metadata |
| Backward compatibility breaks | Preserve old commands and flags, add aliases where needed |

## Open Questions For Implementation

- Should `documents resolve-link` be top-level `craft resolve-link` too?
- Should `craft batch` accept semicolon command strings, JSON operation arrays, NDJSON, or all three?
- Should MCP profiles be allowed as fallback automatically, or require explicit `--backend mcp` for mutating operations?
- Should live mutation tests create a dedicated folder for test artifacts?
- Should `craft audit agent-dx` be built into the binary or live only as `scripts/agent-dx-audit.sh`?

## Recommended First Implementation Slice

Start with the smallest slice that removes current correctness bugs and establishes the architecture:

1. Fix REST move/task/schema/help drift.
2. Add profile model with `--profile`.
3. Add generic MCP client and `craft mcp tools/call`.
4. Add capability metadata to `craft schema`.
5. Add `documents resolve-link` through MCP.
6. Add agent DX audit script baseline.

This creates the foundation for the richer MCP-native features without forcing every command to be rewritten at once.
