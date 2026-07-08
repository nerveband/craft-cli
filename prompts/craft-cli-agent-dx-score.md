# craft-cli Agent DX Score Card

**Date:** 2026-07-08
**Version:** v1.9.0
**Current Runnable Audit:** `craft audit agent-dx --format json`
**Current Score:** 85/85 — Agent-native on the v2 repo-local audit
**Target Floor:** 85/85 — No regression

---

## Axis Scores

### 1. Machine-Readable Output — Score: 2/3
**What exists:** JSON is the default output format. `--format json|compact|table|markdown` across commands. `--json-errors` for structured error output. `--quiet` suppresses status messages.
**Gap:** No NDJSON streaming. No auto-detection of non-TTY context to default to structured output. Errors are not always structured JSON without `--json-errors` flag.

### 2. Raw Payload Input — Score: 1/3
**What exists:** `--json` flag on blocks add/update. `--stdin` for reading from pipe. Some commands accept structured input.
**Gap:** Not all mutating commands accept raw JSON payloads (e.g., `craft create`, `craft tasks add`, `craft folders create` use individual flags). No API-schema-aligned raw input across the board.

### 3. Schema Introspection — Score: 2/3
**What exists:** `craft schema` emits a versioned machine-readable command manifest with flags, safety metadata, backend metadata, and capability metadata. `craft llm` and `craft llm styles` provide human-readable LLM reference docs.
**Gap:** Per-command request/response JSON Schemas are not available yet.

### 4. Context Window Discipline — Score: 1/3
**What exists:** `--output-only <field>`, `--id-only`, `--quiet`, `--raw`, `--no-headers` to reduce output.
**Gap:** No `--fields` flag for arbitrary field selection. No pagination controls. No `--max-results` or `--limit`. Search returns up to 20 results with no control.

### 5. Input Hardening — Score: 0/3
**What exists:** Basic type validation. Structured exit codes (0=success, 1=user, 2=API, 3=config).
**Gap:** No validation for agent hallucination patterns. No rejection of path traversals, percent-encoded segments, or embedded query params in IDs. No security posture statement.

### 6. Safety Rails — Score: 2/3
**What exists:** Global `--dry-run` flag works on: move, blocks update/delete/move, tasks update/delete, folders delete, collections add/update/delete, comments add, clear, delete. Comprehensive coverage.
**Gap:** No response sanitization against prompt injection in API data. Dry-run output is human-oriented prose, not structured JSON.

### 7. Agent Knowledge Packaging — Score: 2/3
**What exists:** `AGENTS.md` with debugging resources. `docs/llm/` folder with styling reference and output parity docs. `prompts/` folder with implement/check/release prompts and agent-DX guidelines.
**Gap:** No per-command structured skill files. No versioned skill library. No explicit agent guardrails in AGENTS.md (e.g., "always use --dry-run", "always use --id-only").

---

## Current 50-Point Audit

Run:

```bash
craft audit agent-dx --format json
```

Current local result: **85/85** on the v2 audit. This includes the legacy defensive baseline plus the compounding checks for introspection, profiles, artifacts, contract discipline, Unix restraint, API-native payload ergonomics, and domain proof gates.

## Legacy 21-Point Estimate: 12/21 — Agent-ready

| Axis | Current | Target | Priority |
|------|---------|--------|----------|
| Machine-Readable Output | 2 | 3 | Medium |
| Raw Payload Input | 1 | 3 | High |
| Schema Introspection | 2 | 3 | Medium |
| Context Window Discipline | 1 | 2 | Medium |
| Input Hardening | 0 | 2 | Medium |
| Safety Rails | 2 | 3 | Low |
| Agent Knowledge Packaging | 2 | 3 | Medium |
| **Total** | **12** | **18** | |

---

## Bonus: Multi-Surface Readiness

- [x] **MCP (stdio JSON-RPC)** — `.mcp.json` configured for project-scoped MCP server
- [ ] **Extension / plugin install** — not available as installable extension
- [x] **Headless auth** — `CRAFT_API_URL` env var support, profile-based config

---

## Comparison: craft-cli vs gustin/craft-cli

| Feature | nerveband/craft-cli | gustin/craft-cli |
|---------|-------------------|-----------------|
| Language | Go (compiled binary) | Shell/npm |
| Distribution | goreleaser, multi-platform | npm install |
| JSON output | Default | `--json` flag |
| Dry-run | Global flag, all mutating cmds | Unknown |
| LLM docs | `docs/llm/`, `prompts/` | None |
| Commands | 30+ including blocks, collections, tasks, comments | Similar coverage via shell wrappers |
| Agent DX score | 8/21 | Estimated 4-6/21 |
| MCP support | Yes (project-scoped) | Unknown |
| Stars | 2 | 0 |

---

## Roadmap to Agent-First (16+/21)

### Phase 1: Quick Wins (+4 points, target 12/21)
1. Keep `craft schema` command outputting JSON command manifest → Schema Introspection 2→3 with request/response schemas
2. Auto-detect non-TTY and default to JSON → Machine-Readable Output 2→3
3. Add `--limit` / `--page` flags → Context Window Discipline 1→2

### Phase 2: Contract Alignment (+4 points, target 16/21)
4. Raw JSON payload input on all mutating commands → Raw Payload Input 1→3
5. Input validation for hallucination patterns → Input Hardening 0→2

### Phase 3: Polish (+2 points, target 18/21)
6. Structured dry-run output + response sanitization → Safety Rails 2→3
7. Per-command skill files with guardrails → Agent Knowledge 2→3
