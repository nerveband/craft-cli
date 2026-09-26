# Craft API and agent CLI audit

Date: 2026-09-26. Commit: `ac51d0232120fa6f4c6a8501876536d2cb63fb38` (`v1.12.0`). Scope: assessment of `craft-cli` v1.12.0, the live [Craft Connect documentation](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1), and `nerveband/cli-best-practices`. No Craft mutations were made. Architecture preference: keep the CLI as the primary user/agent interface, use direct REST where it covers the operation, and use MCP to augment missing functionality. No latency benchmark was run; direct REST avoids the additional MCP command-dispatch layer.

## Evidence and limits

- Fetched the live Craft page on this date and extracted its embedded OpenAPI 3.1.1 document from the page loader data. Compared method/path inventories and parameter definitions with the saved contract, then inspected request/response schemas against the client. The page describes this particular all-documents link, so availability of experimental features may differ for another connection.
- Compared `cmd/`, `internal/api/client.go`, `internal/models/document.go`, `docs/contracts/craft-rest-space-openapi.json`, `docs/capabilities/`, and `docs/llm/`. The older `docs/HHRuPxZZTJ6-docs.md` and `docs/HHRuPxZZTJ6-openapi.json` named in plans and `docs/contracts/README.md` are **not present in this checkout**. The checked-in REST OpenAPI contract contains 28 method/path operations; the current contract has 44 operations across 21 paths. Additions are 5 collection-view operations, 4 reminders operations, 5 whiteboard operations, collection creation, and collection schema update. This is a snapshot gap, not proof those features never existed.
- Initial assessment encountered an unreadable sibling checkout and used published files. The user subsequently supplied a repaired clean local clone at `24b9960`. The Audit v3 baseline, plan integration and final audit use that local source exclusively and supersede the initial policy comparison.
- REST findings are static contract comparisons; no document-content requests or writes were made. MCP assessment used the endpoint supplied in this conversation for `initialize`, `tools/list`, `resources/list`, and command help only. The URL is intentionally not copied into this report. Unknown-command results are live confirmations; request-shape findings are contract mismatches without mutation testing.

## Priority recommendations

| Priority | Recommendation | Reason |
|---|---|---|
| P0 | Honor MCP tool errors and dry-run/confirmation in generic calls | `isError: true` is currently returned as successful CLI output, and `mcp call` does not enforce the inherited dry-run flag. |
| P0 | Repair task repeat shapes, task reads, nested page decoding, and folder parent fields | These are contract correctness defects already present in the older snapshot, not just missing new features. |
| P0 | Repair `GET /documents/search` parameter names and array encoding, then add request contract fixtures | The CLI emits `folderIDs` and `documentIDs` while the live docs specify `folderIds` and `documentIds`. Multi-scope search can silently lose its filter. |
| P0 | Preserve the actual Craft response shape for block reads and write results | The typed models omit major block variants and attributes; several write methods discard `items`, making partial outcomes and server-assigned IDs unavailable. |
| P1 | Refresh the checked-in REST contract from the current downloadable Craft OpenAPI, then diff it in CI | The published contract is behind the live docs and its source file references are absent. A repeatable diff makes later drift visible. |
| P1 | Add REST wrappers for collection views and, when available for the connection, reminders | Current REST docs expose these operations. The CLI currently requires MCP for views and has no reminder commands. |
| P1 | Expose rate-limit and retry metadata without automatic mutation retries | Craft documents public-IP and space budgets, `Retry-After`, and block budgets; `APIError` drops headers. Safe callers need the delay and scope. |
| P2 | Correct stale capability guidance and the best-practices scorecard | Documentation currently labels views MCP-only and the scorecard rates v1.9.0. |

## Craft Connect gap table

“New” means absent from the checked-in snapshot or CLI, not a claim about Craft's release date. Source links point to the [live link-specific reference](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1) unless otherwise noted.

| Area | Live contract or behavior | v1.12.0 evidence and gap | Status / action |
|---|---|---|---|
| Document search scopes | `GET /documents/search` uses `folderIds` and `documentIds`, each string or array, plus `include` and `regexps` as string or array. | `internal/api/client.go:1459-1462` sends `folderIDs` and `documentIDs`; `cmd/search.go:33-47` joins IDs with commas. `SearchDocumentsAdvanced` sets one scalar value for `include` and `regexps`. | **Wrong parameter casing and potentially wrong multi-value encoding.** Fix against fresh OpenAPI examples and test query serialization. |
| Search options and results | Document search supports `fetchBlocks`; block search supports `fetchBlocks`, before/after counts, and structured context. | `internal/api/client.go:1405,1433-1487` and `cmd/search.go:108-170` do not expose `fetchBlocks`; the CLI offers one `--context` for both directions and `--metadata` for document search, although the current query schema replaces `fetchMetadata` with `fetchBlocks`. Response-field descriptions still mention `fetchMetadata`, an upstream documentation inconsistency. | Changed query contract and missing options. Add `fetchBlocks`, retain `blockIds` and `blocks` in responses, and clarify metadata compatibility with Craft. |
| Collection views | REST now documents list/create on `/collections/{collectionId}/views`, update/delete on `/collections/{collectionId}/views/{viewId}`, and `PUT /collections/{collectionId}/active-view`. View definitions cover table, gallery, kanban, filters, sorts, grouping, and active view. | `docs/contracts/craft-rest-space-openapi.json` has no view paths. `cmd/collections.go:455-543` and `cmd/schema.go:165-173` mark views MCP-only. `docs/capabilities/rest-vs-mcp.md` repeats this. | **New REST surface / stale backend map.** Keep MCP for features REST still lacks, but implement and advertise REST view controls where equivalent. |
| Reminders | Experimental `GET/POST/PUT/DELETE /reminders`; GET has `status`, `limit` (1 to 200), and `cursor`; reminders are block-scoped and connection availability depends on ownership/auth conditions. | No reminder path in checked-in contract, API client, models, or command tree. Task date flags are not timed reminder support. | **New conditional feature.** Add capability detection and cursor-safe CLI commands only after validating availability on a suitable test connection. |
| Whiteboards | REST has `POST /whiteboards` and CRUD on `/whiteboards/{whiteboardBlockId}/elements`, with element limits of 500 per request. | `internal/api/client.go:1778-1848` and `cmd/whiteboards.go` cover them, but the checked-in REST OpenAPI omits all paths; `docs/contracts/README.md` calls that file the current baseline. | Implemented but contract snapshot missing. Add to refreshed contract and fixtures. |
| Upload | `POST /upload` takes binary body plus optional `fileName` query parameter and a required placement. File type derives from `Content-Type`; filename is ignored for image/video. | `internal/api/client.go:1677-1704` always uses `application/octet-stream` and never sends `fileName`. `cmd/upload.go:61-83,154-190` derives a name locally but does not pass it to the client. | Missing parameter and media type fidelity. Preserve filename and detected or explicit MIME type. |
| Block read representation | `GET /blocks` returns a union with page `title` object and `styling`, collection item `properties`, image/video dimensions and MIME data, separator style, and other media fields. It also accepts `Accept: text/markdown`. | `internal/models/document.go:16-61,107-122` flattens the union into a generic `Block`, omitting many fields; `BlocksResponse` omits page title/styling. `internal/api/client.go:234-263` builds markdown locally from JSON instead of requesting the server-rendered markdown representation. | **Wrong type plus lossy fields.** `Block.Title` is a string, so a page title object causes a JSON decode error in direct or nested block reads, not merely a dropped field. Model discriminated variants or retain `json.RawMessage` for unknown fields; test markdown negotiation separately. |
| Block insert/update shape | `POST /blocks` supports `blocks` plus position or `markdown` plus position; writes return `items` with assigned IDs. `PUT /blocks` applies only supplied fields. | Raw JSON path exists (`internal/api/client.go:1715-1748`), but typed update fields use `omitempty` for zero/empty values (`internal/models/document.go:169-204`), so clearing a value through typed input can be impossible. Several helpers return only error or the first item. Raw map payloads can preserve explicit empty values; distinguish that supported escape hatch from typed flags. | Partial update hazard. Use presence-aware fields for clear operations and surface full result arrays. |
| Batch write outcomes | Document, folder, task, block, and collection writes return `items` arrays; some contain IDs or updated objects. | Methods such as `DeleteDocuments`, `DeleteBlocks`, `MoveDocuments`, and `UpdateTasksRaw` discard response bodies (`internal/api/client.go:462-519,767-790,1119-1154`). | Missing response fidelity and partial-result reporting. Retain server results before adding retries or batching further. |
| Rate limits | 50 requests/10 s per public IP, 100 requests/60 s per space, and 20,000 blocks/60 s per space. Responses may include `X-RateLimit-*`, `X-BlockBudget-*`, and `Retry-After` on 429. | `internal/api/client.go:75-169` reads only status/body and rewrites 429 to “Retry later”; `APIError` has no retry duration or headers. | Changed operational contract not reflected in client. Expose retry delay and rate budgets in structured errors, especially for read-only callers. |
| Task semantics | `GET /tasks?scope=all` means all task blocks, not a union of inbox/active/upcoming/logbook. Task writes can include location changes. | CLI has task CRUD in `cmd/tasks.go` and raw API support, but typed `UpdateTask` only takes state/dates/repeat (`internal/api/client.go:1090-1118`); command help does not explain the special `all` semantics. | Missing typed move/location capability and documentation precision. |
| Folders | Live docs say delete moves descendants to parent/Unsorted, but the delete request description also says “Folders must be empty.” | `cmd/folders.go:206-275` states contents will be moved to parent. | **Ambiguous upstream reference.** Confirm with a reversible fixture before changing behavior or help. |
| Recurring task request shape | For task endpoints, `repeat` is a sibling of `taskInfo`. `repeat.type` is `fixed` or `flexible`; `frequency` is daily/weekly/monthly/yearly. Options are nested under `daily`, `weekly`, `monthly`, `yearly`, and `reminder` is an object with `enabled` and `dateOffset`. | `internal/api/client.go:981-982,1101-1102` puts repeat inside `taskInfo`. `internal/models/document.go:75-87` and `cmd/tasks.go:342-390` use frequency names as `type`, numeric `weekdays`, boolean `dynamicDays`, and a string reminder. | **P0 wrong request shape, already inconsistent with the older contract.** Repair typed flags/models and their tests. Raw task JSON can send the proper top-level repeat shape. Add explicit null support for removing repeat/dates. |
| Task list response and scope | Task state/dates are under `items[].taskInfo`; placement is under `location`. Document reads require `scope=document` and `documentId`. | `Task` puts state/dates at top level and lacks `location` (`internal/models/document.go:144-156`). `GetTasks` decodes straight into it; `GetDocumentTasks` sends only `documentId` (`internal/api/client.go:915-950`). | **P0 wrong response shape and missing required query scope.** Task status/dates/location can disappear, and document task reads do not meet the query contract. |
| Folder parent and tree | Create uses `folders[].parentFolderId`; list returns nested `folders[]` recursively. | `CreateFolder` sends `parentId` (`internal/api/client.go:668-692`); `Folder` lacks nested `folders` (`internal/models/document.go:124-130`). | **P1 wrong input field and lossy output.** These mismatches also predate the current live snapshot. |
| Count fidelity | Document, task, folder, and search list schemas return `items` without a server `total`. | Typed lists add an integer `Total` that defaults to zero. `craft list --count` uses it directly (`cmd/list.go:65-66`), and `GetDocumentsFiltered` / `GetDocumentsAdvanced` do not compute it (`internal/api/client.go:180-206,1517-1566`). | **P1 incorrect count for conforming responses.** Derive count from items only when the result is complete; distinguish returned count from unknown global count on paginated APIs. |
| REST page styles | The current block-insert schema includes page `styling`; the block-update union and read schemas expose richer styling fields than the typed models. | Raw block JSON is already available, while native style flags auto-route to MCP (`cmd/blocks.go`, `internal/models/document.go:169-204`). | **Partial REST capability hidden by broad MCP-only guidance.** Compare each style field before adding a REST adapter; retain MCP-only theme exploration and revert/review features. |
| Deprecated features | The current reference does not clearly mark any existing REST operation as deprecated. | No defensible deprecation finding from this comparison. | Do not remove commands based on absence from one snapshot. |

The checked-in snapshots under `docs/contracts/` use the same 28-operation REST baseline for beta and space. The original older Markdown snapshot requested in the brief is unavailable here, so historical comparisons use the checked-in OpenAPI instead. No operations were removed in the current 44-operation inventory. Collection creation and schema update already have CLI methods despite being absent from the old snapshot. The live page also includes Craft-specific markdown tags such as `<page>`, `<card>`, `<highlight>`, and output-only collection tags; `docs/llm/styling-and-markdown.md` should be checked against those after the contract refresh.

## Current Craft MCP assessment

Queried the user-supplied endpoint on 2026-09-26 with discovery and help requests. Initialization negotiated `2025-06-18` and reported server `craft-workflow-links` version `1.0.0`. This establishes compatibility with that revision; it does not establish support for every newer MCP revision. The initial Python HTTP client received 403; curl with the same JSON-RPC requests succeeded. No profile was saved.

| Discovery item | Current result | Difference from saved contract |
|---|---|---|
| Tools | `craft_read`, `craft_write`, `blocks_revert` | Same three tool names and input schemas. |
| Read/write schema | One required string, `command`, using Draft-07 JSON Schema | Full command grammar is exposed through help, not the outer tool input schema. None of the three tools advertises `outputSchema`. |
| Annotations | Read tool is read-only; write and revert are destructive. `openWorldHint` is true on read/write and false on revert. | Read/write previously had `openWorldHint: false` in `docs/contracts/craft-mcp-tools.json`. |
| Tool descriptions | Read/write now mention block reminders; write task examples use `--task`. | Saved write description uses `--id` for task update/delete. `blocks_revert` is unchanged. |
| Resources | One resource, `ui://craft/edit-review` | Same URI as `docs/contracts/craft-mcp-resources.json`. Only the resource listing was requested; no UI resource body or private document content was fetched. |

| Priority | Gap | Evidence and recommendation |
|---|---|---|
| P0 | MCP application errors can exit successfully | The harmless `craft_read` command `collections views list --help` returned a successful JSON-RPC result containing `isError: true` and an unknown-command message. `internal/mcp/client.go:133-143` only checks JSON-RPC `error`; `cmd/mcp.go:112-117,324-343` prints the result. Batch marks `ok` from Go error alone (`cmd/mcp.go:264-294`). Treat tool errors as failures and preserve their content; test batch stopping and final exit status. |
| P0 | Generic MCP calls bypass the advertised dry-run contract | `cmd/mcp.go:86-118` calls the tool without checking `isDryRun()` or requiring commitment for writes. The global flag and inferred schema imply dry-run is supported. Fix this before recommending generic calls as a safe fallback. |
| P1 | Collection-view wrappers generate invalid MCP grammar | `cmd/collections.go:466-543` emits `collections views list/create/update/delete` and `collections active-view set` with positional IDs. Current help specifies `collections views-list/views-create/views-update/views-delete/views-set-active --collection ID`, with view-specific flags. The read form was confirmed invalid live. Prefer direct REST implementations for these operations and preserve the public CLI syntax. |
| P1 | View settings are largely absent from native flags | Current `views-create --help` documents filters, sort/group rules, hidden/order/visible fields, widths, calculations, gallery previews, and kanban columns. Wrappers only assemble name/type/JSON and view ID. A generic `--json` flag is not advertised by that MCP help. Map the REST view object directly or explicitly translate supported fields. |
| P1 | Task-write capability is incorrectly denied | Live `craft_write tasks --help` exposes add/update/delete. `tasks update --help` supports markdown, state, dates, location, repeat JSON, and `--no-repeat`. `cmd/tasks.go:318-326` and `AGENTS.md` say MCP exposes no task writes, which is false for this endpoint and contradicted even by the saved tool description. Keep REST task CRUD as the primary route. Correct the capability claim, but do not infer that task writes supply reversible metadata: that remains untested. |
| P1 | Reminders lack native CLI access | Both read and write help expose reminders. Create accepts a block ID and an ISO timestamp with an explicit offset, or a REST-shaped batch array; omitting time means Save for later. Implement the REST surface first, with MCP augmentation only if needed for connection-specific features. Preserve timezone/DST semantics from help and connection metadata. |
| P2 | Useful MCP discovery has no native wrapper | `blocks learn` returns topic-specific block JSON guidance; `folders update` appears in write help. Consider a CLI help adapter for `blocks learn` and a folder-update wrapper if REST remains unavailable. Generic calls can reach them once error and safety handling is corrected. |
| P2 | MCP transport handles only a narrow response pattern | `internal/mcp/client.go:65-77` hard-codes client version `1.0.0` and protocol `2025-06-18`. The client does not retain negotiated sessions/version or paginate discovery. Its SSE helper joins every data line into one JSON document (`:146-168`), which fails for multiple JSON-RPC events. Current discovery succeeded with one event and no pagination. Harden only the necessary transport behavior with fixtures before claiming broader compatibility. |

The existing `images view --url`, `whiteboards elements get --whiteboard`, collection rename, and collection item add/update wrappers match the discovered command names. Their actual data operations were not executed. Do not classify every MCP adapter as broken because the collection-view grammar is wrong.

The requested architecture can remain simple: agents call `craft`; the CLI uses REST for normal documents, blocks, tasks, folders, collections, uploads, whiteboards, views, and reminders where supported. MCP supplies review/revert metadata, discovery helpers, link resolution, and verified gaps. Backend selection should be explicit in schema/results, and failed writes must never be retried on the other backend automatically.

## `cli-best-practices` audit of craft-cli

The published [audit advertised as 85 points](https://github.com/nerveband/cli-best-practices/blob/main/scorecards/agent-cli-audit.md) and [skill](https://github.com/nerveband/cli-best-practices/blob/main/SKILL.md) are useful checklists, but this assessment does not issue a numerical score. The published checklist actually contains 97 numbered checks, which also needs correction. `craft audit agent-dx` in `cmd/audit.go` mostly checks command and flag presence and is not a behavioral certification. These are the material gaps found against those rules.

| Priority | Convention and finding | Evidence |
|---|---|---|
| P0 | **Declared dry-run support is not reliable.** Generic `mcp call` inherits `--dry-run` but invokes the tool directly; `feedback` writes a local file without checking it. `inferSafety` defaults both leaf names to `supports_dry_run: true`. | `cmd/mcp.go:86-118`, `cmd/agent_support.go:34-73`, `cmd/schema.go:237-250`. Enforce or explicitly reject the flag, and declare safety per full command path. |
| P0 | **Schema inspection flags do not prevent execution.** `--request-schema` and `--response-schema` are registered globally but never read, so an invocation intended to inspect a write schema can execute the command instead. | `cmd/root.go:44-45,151-152`; a repository-wide symbol search finds only declarations and flag binding. Implement early schema-only dispatch before auth, validation requiring resource IDs, or network calls. |
| P1 | **Bundled skill discovery points at the caller’s directory.** `craft skill-path` returns `$PWD/SKILL.md` without checking whether it exists. An installed binary invoked outside the repo can advertise an unrelated or nonexistent skill. | `cmd/agent_support.go:13-30`; use an installed asset path or a supported embedded export. |
| P1 | **The machine schema is incomplete for agent invocation.** The general command manifest has flag names/types but no positional argument schema, enums, output fields, error kinds/exit codes, or complete command-path filtering. The registered `--request-schema` and `--response-schema` booleans are never read anywhere in the Go source, so those flags do not implement schema output. `schema --command` searches only root children. | `cmd/schema.go:13-43,61-82,98-150`; compare the [CLI Spec v0.2 schema guidance](https://clispec.dev/spec/v0.2/). |
| P1 | **Large REST reads remain unbounded upstream.** `craft list --limit` fetches all documents and truncates locally; `--count` also fetches the list. This reduces displayed tokens but not network/cost/rate exposure. | `cmd/list.go:43-91`, `internal/api/client.go:1517-1566`. Disclose client-side limit and use server pagination where Craft offers it. |
| P1 | **Safe retries are undocumented in machine results and non-idempotent creates have no token.** The CLI warns about ambiguous write timeouts, but the API client has only a fixed 30-second timeout and no request correlation or idempotency key. | `internal/api/client.go:18,60-73,75-114`; `cmd/llm.go:90`; [safety guidance](https://github.com/nerveband/cli-best-practices/blob/main/patterns/safety-rails.md). If Craft lacks keys, return an explicit uncertain-outcome error and a narrow verification recipe. |
| P1 | **Several advertised output controls are inert or unsupported.** JSONL/YAML are named in root help but absent from output dispatch. `--deliver`, `--transform`, and `--data-source` are registered but their variables have no execution path. | `cmd/root.go:29-31,137-140`, `cmd/output.go:129-133`, `cmd/output_blocks.go:498-499`. Implement the documented supported behavior or reject unsupported options before any mutation. |
| P1 | **The success envelope is not stable across commands.** Some commands return resource JSON, others print only an ID in quiet mode, and write helpers discard `items`. This makes the generic schema's absent output fields more costly. | `cmd/upload.go:189-203`; `internal/api/client.go:462-519,1119-1154`; `cmd/output.go`. Specify per-command response schema without forcing a breaking envelope change. |
| P2 | **Agent knowledge and scorecards drift.** `docs/llm/README.md` and capability maps describe views as MCP-only; `scorecards/craft-cli.md` in the published best-practices repo rates v1.9.0 and says `--count` is missing, though v1.12.0 has it. | `docs/llm/README.md:14-25`; `docs/capabilities/rest-vs-mcp.md:31-39`; `cmd/list.go:141-163`; [published scorecard](https://github.com/nerveband/cli-best-practices/blob/main/scorecards/craft-cli.md). |
| P2 | **Interactive setup needs a non-TTY guard.** Normal profile commands exist, but `craft setup` still reads from stdin and prints prompts. A pipeline can hang or consume piped data unintentionally. | `cmd/setup.go:35-138`. Fail early without a TTY and point to `craft config add` or `craft profiles add-rest`. |

Positive coverage: default JSON, `--json-errors`, exit codes, dry-run, `--yes`, typed profiles, batch IDs, field projection, `--count`, MCP escalation, and LLM docs already satisfy many checklist items. Keep their existing contracts stable while addressing the gaps.

## Is `cli-best-practices` current?

The published repository is directionally useful but contains stale craft-cli observations and a few recommendations that should be narrowed. These are concrete edits for **that repository**, not changes made here.

| Priority | Addition or correction | Basis |
|---|---|---|
| P1 | Correct `principles/contract-first.md`, which says craft-cli has no schema and scores 0/3 on introspection, while the same repository’s scorecard awards 2/3. Preserve the principle that types can still encode the wrong external contract unless fixtures validate them. | Compared `principles/contract-first.md` with `scorecards/craft-cli.md` and current `cmd/schema.go`. |
| P1 | Fix the audit denominator in `SKILL.md`, `scorecards/agent-cli-audit.md`, README references, and scoring templates. The 17 category maxima and numbered checks sum to 97, not the advertised 85. Version the checklist and allow documented not-applicable results. | Counted all `### N.N` headings in published `scorecards/agent-cli-audit.md`: 97. |
| P1 | Update `scorecards/craft-cli.md` to v1.12.0 and rerun behavior checks; update the `--count` finding and verify JSONL behavior instead of scoring its help text. Separate observed results from self-reported `craft audit agent-dx` checks. | The scorecard is dated 2026-04-01 and v1.9.0. Current `cmd/list.go` has `--count`. `cmd/root.go` advertises JSONL, but `cmd/output.go` does not implement that format, so registration alone must not earn a passing score. |
| P1 | Add a versioned CLI contract section. Recommend the frozen [CLI Spec v0.2](https://clispec.dev/spec/v0.2/) for present conformance, and describe [v0.3](https://clispec.dev/) as an August 2026 candidate. Require command effects, output cardinality, error kinds, and schema validation. | CLI Spec now formalizes introspection and bounded output; candidate status matters for compatibility claims. |
| P1 | Make retries an outcome matrix: read-only, idempotent write, non-idempotent create, and ambiguous timeout. Do not imply every create can safely be retried or made idempotent locally with check-then-create. | [CLI Spec safe retries](https://clispec.dev/spec/v0.2/) and Craft's lack of an advertised idempotency-key contract. A preflight lookup has a race. |
| P1 | Correct the MCP comparison. Describe CLI and MCP as complementary surfaces; use MCP tool `inputSchema`, optional `outputSchema`, `structuredContent`, `isError`, and tool safety annotations where applicable. Avoid saying MCP necessarily loads all schemas into context or that CLI is inherently cheaper. | [MCP tools specification and SDK guidance](https://modelcontextprotocol.io/specification/2025-06-18/server/tools), [MCP 2026 release notes](https://blog.modelcontextprotocol.io/posts/2026-07-28/). Discovery and schema selection depend on the host/client. |
| P2 | Ground output rules in [clig.dev](https://clig.dev/): data on stdout, diagnostics on stderr, explicit machine format, no ANSI in pipes, and compatibility with established human defaults. Do not require JSON default or `--quiet` when clean structured output and explicit format already solve the problem. | [Command Line Interface Guidelines](https://clig.dev/) and [CLI Spec v0.2](https://clispec.dev/spec/v0.2/) allow declared existing defaults. |
| P2 | Replace “sanitize known prompt injection patterns” with a trust-boundary rule: preserve remote data faithfully, label provenance, escape control sequences in human displays, and never execute returned text as instructions. | `patterns/safety-rails.md:62-68` risks destructive alteration of legitimate data; schema and provenance give a clearer boundary. |
| P2 | Narrow generic advice for `--dry-run`, `--yes`, `--fields`, local sync, webhooks, feedback, and self-update to commands where the behavior is meaningful. State that dry-run previews local intent unless the service offers a true validation endpoint. | The published skill and scorecard make several features universal; [CLI Spec v0.3](https://clispec.dev/) explicitly distinguishes command effects and output cardinality. |
| P2 | Update skill guidance to cite the [Agent Skills specification](https://agentskills.io/specification): concise frontmatter, task triggers, progressive disclosure, and workflow checks. Keep repository `AGENTS.md` for local project rules; do not require an agent-only CLI mode or `AI_AGENT` environment switching without user opt-in and a stable documented contract. | [Agent Skills](https://agentskills.io/specification); `patterns/agent-knowledge.md:17-97` treats several emerging conventions as stronger than their evidence supports. |

## Other useful inputs and verification plan

1. Use the live page's embedded OpenAPI document as a reproducible contract source. An OpenAPI document is demonstrably published in loader data; a stable standalone export URL was not verified. The checked-in `craft-rest-space-openapi.json` should carry fetch date, source URL, content hash, and a generated operation/field diff. Do not overwrite it in this assessment.
2. Add fixture-backed contract tests for query casing and array serialization, response decoding of every block variant, upload query and MIME, empty-value updates, and write `items` handling. Keep default tests offline. Use opt-in read-only live probes on the public test link only when a fixture cannot settle behavior.
3. Continue using the CLI for all normal workflows, with MCP adapters for review/revert metadata, exploration helpers, link resolution, and other verified REST gaps. Do not assume every page styling field is MCP-only: the new REST block schema includes `styling` and separator fields, so compare individual fields before routing. Recheck the current `craft_read`/`craft_write` tool schemas before hard-coding an MCP-only classification. The CLI's `craft mcp tools` and `docs/contracts/craft-mcp-tools.json` provide comparison inputs.
4. Add a capability table generated from the REST OpenAPI plus MCP tool schemas, with explicit “REST”, “MCP”, “both”, and “conditional” entries. Use it to check `docs/capabilities/`, `docs/llm/`, and `craft schema` in CI.
5. For rate-limit behavior, use mocked `429` responses containing and omitting `Retry-After`, plus `X-RateLimit-*` and `X-BlockBudget-*`. Test that the CLI conveys a safe retry delay without duplicating mutations. Live probes should remain read-only.

## Comprehensive implementation plan

This plan incorporates the follow-up request for fixes and README/related documentation. The user subsequently authorized comprehensive implementation after an OMP Opus 5.5 review at high reasoning. That authorization supersedes the initial assessment-only file restriction for craft-cli changes. Commit, push, release, and production mutation tests remain separate actions. Work can be done by one agent; no delegation is needed.

### 1. Establish the contract baseline

Files: `docs/contracts/README.md`, the two REST snapshots, MCP tool/resource snapshots, and a small contract refresh/check script if implementation is authorized.

- Replace the space baseline with the fetched 44-operation OpenAPI; identify beta as an alias of the same source rather than implying an independently verified API.
- Record source, capture date, protocol/server versions, and hash. Use sanitized tool/help fixtures without private content or a working MCP link.
- Record exact deltas: collection views, reminders, collection create/schema update, whiteboards, search `fetchBlocks`, removed search query `fetchMetadata`, upload filename, and response field additions.
- Add a deterministic offline inventory check. Network refresh is an explicit maintenance command, not part of default tests.

Acceptance: fixtures cover all 44 method/path pairs; refresh output distinguishes additions, removals, and changed schema fields; no production data appears in snapshots.

### 2. Fix dangerous CLI control-flow gaps first

Files: `cmd/root.go`, `cmd/schema.go`, `cmd/mcp.go`, `cmd/mcp_write.go`, `cmd/agent_support.go`, `internal/mcp/client.go`, and their existing tests.

- Intercept request/response schema flags before command execution. Unsupported schema requests must fail explicitly rather than falling through to a mutation.
- Ensure `mcp call --dry-run` returns only a local preview. Require the existing explicit commitment convention for `craft_write` and `blocks_revert`; handle unknown tool effects conservatively without pretending every read needs write confirmation.
- Decode MCP `isError` and distinguish it from transport and JSON-RPC errors. Return nonzero status for failed calls and batches, with partial result records where work already succeeded.
- Replace leaf-name safety inference with explicit metadata on executable commands. Mark local feedback writes and schema reads accurately.

Acceptance: fixture-server request counters stay zero for schema-only and dry-run invocations; malformed or unsupported schema requests cannot trigger writes; `isError: true` produces a structured error and failure exit; batch stops or continues according to one documented policy.

### 3. Repair existing REST request contracts

Files: `internal/api/client.go`, `internal/models/document.go`, `cmd/search.go`, `cmd/folders.go`, `cmd/tasks.go`, `cmd/upload.go`, and focused tests.

- Correct search query casing. Encode arrays according to the refreshed OpenAPI serialization rules and verify against documented examples; do not assume comma joining is accepted.
- Add `fetchBlocks` to both search modes and expose separate context-before/context-after options while retaining the existing `--context` shortcut. Expose search daily-note filters already represented in the API option struct.
- Send `parentFolderId` for folder creation and `scope=document` for document task reads.
- Replace repeat modeling with `type`, `frequency`, nested frequency options, and reminder object; place it at task item level. Map existing unambiguous flags to valid payloads and reject obsolete ambiguous inputs with actionable migration guidance.
- Support task date/repeat clearing with explicit null semantics, task content updates, and location changes. Keep raw JSON as the complete escape hatch.
- Send upload filename and a correct MIME type, allowing explicit override for stdin. Preserve page/date/sibling placement validation.

Acceptance: table-driven request fixtures validate exact names, types, nesting, null values, and target placement. Repeat tests validate the external contract instead of merely matching current implementation structs.

### 4. Preserve response data and truthful outcomes

Files: `internal/models/document.go`, `internal/api/client.go`, `cmd/output.go`, `cmd/output_blocks.go`, `cmd/list.go`, `cmd/search.go`, `cmd/tasks.go`, `cmd/folders.go`.

- Handle page title objects without breaking rich-link string titles. Preserve page styling, media fields, collection properties, comments, search block IDs, and nested folder trees.
- Decode task metadata/location correctly and retain recurrence objects returned by reads and writes.
- Derive counts only for complete collections. Report returned count, truncation, and unknown total separately where pagination applies.
- Preserve mutation `items` and expose successes/failures rather than synthesizing complete success from the requested IDs.
- Keep legacy output mode stable. If correcting the default JSON shape breaks consumers, document a compatibility migration and choose a version boundary before release rather than silently changing contracts.

Acceptance: fixtures include nested pages, every block variant, folder nesting, task locations, metadata, partial mutation outcomes, zero results, and nonempty lists without `total`. Existing compact output remains stable where promised.

### 5. Add the missing REST capabilities

Files: `cmd/collections.go`, `internal/api/client.go`, `internal/models/document.go`, and a focused reminder command file if needed.

- Implement the five collection-view endpoints behind the existing `craft collections views` / active-view commands. Use REST by default when selected, with raw JSON/file/stdin support for complete view configurations.
- Validate kanban grouping and preserve omitted settings; allow explicit empty settings where the service defines clearing.
- Add reminder list/create/update/delete using the four REST endpoints. Support `status`, `limit`, and cursor pagination; keep the status consistent across pages.
- Explain connection-dependent reminder availability. Accept explicit timezone-offset timestamps and use connection timezone only when offering relative-time conversion; cover daylight-saving ambiguity.
- Expose full placement for whiteboard creation and the existing 500-element bounds. Compare REST style fields individually and enable direct REST where the contract supports the same operation.

Acceptance: all new commands have local dry-runs, request/response fixtures, accurate capability metadata, and bounded output. A missing reminder capability produces a useful error without silently switching accounts/backends.

### 6. Make MCP augmentation accurate and resilient

Files: `cmd/collections.go`, `cmd/tasks.go`, `cmd/blocks.go`, `cmd/mcp.go`, `internal/mcp/client.go`, and MCP tests.

- Correct the collection view grammar and named IDs for explicit MCP execution. Translate supported flags, avoiding undocumented generic JSON forwarding.
- Remove the false claim that task writes do not exist. Keep task CRUD on REST by default; describe task revert/diff support as unverified until actual result metadata is tested in an authorized temporary fixture.
- Add a small `blocks learn` wrapper if useful and folder update where REST has no equivalent. Keep image and whiteboard wrappers that already match current help.
- Parse SSE events independently and match the requested JSON-RPC ID. Preserve structured content and error metadata. Handle negotiated protocol/session headers where required, and report the actual CLI version in `clientInfo`.
- Add discovery pagination only where needed; current three-tool discovery is not paginated. Treat annotations as hints, not authorization.

Acceptance: help-contract fixtures match current commands; tests include tool errors, multiple SSE events, JSON responses, malformed responses, and session requirements. No automatic failover of a possibly applied write.

### 7. Improve retry and context controls

Files: `internal/api/client.go`, `internal/mcp/client.go`, `cmd/root.go`, `cmd/list.go`, `cmd/output.go`, and error tests.

- Retain HTTP status, `Retry-After`, rate budgets, and a request identifier when supplied. Surface timeout as uncertain outcome for writes.
- Allow a configurable request deadline/cancellation. Retry only clearly safe operations under a documented bounded policy; do not add a general write retry loop.
- Distinguish server pagination from client-side truncation in output/help. NDJSON should not be described as streaming if it is formatted only after the whole response arrives.
- Implement the advertised JSONL/YAML formatting, structured transforms, and atomic file delivery, or reject combinations that cannot honor them before execution. Reject an unsupported local data source explicitly rather than silently using live data.
- Make `--quiet` suppress diagnostics without unexpectedly changing the data shape; retain existing ID-only behavior only through a compatibility decision or explicit `--id-only` contract.

Acceptance: mocked 429 and timeout tests verify metadata and no duplicate mutation; output tests cover JSON, compact, JSONL, projection, and truncation according to the documented contract.

### 8. Fix discovery, setup, and audit quality

Files: `cmd/schema.go`, `cmd/audit.go`, `cmd/agent_support.go`, `cmd/setup.go`, `SKILL.md`, and generated command reference.

- Support nested command-path schema queries, positional arguments, enums, required inputs, effects, backends, errors, and real output schemas. Ensure schema works without valid configuration or network access.
- Generate/check `docs/command-reference.json` from command metadata. Consider an optional CLI Spec adapter after stabilizing the existing schema, without claiming candidate-spec conformance prematurely.
- Resolve `skill-path` to a shipped asset or offer a supported export. Verify from outside the repository.
- Refuse interactive setup without a TTY and print the noninteractive configuration alternative. Correct its contradictory `--quiet` help.
- Replace presence-only audit checks with targeted behavior checks where they matter. Report skipped/not-applicable checks and derive denominator from the actual versioned checklist.

Acceptance: schema/dry-run tests run with a fake failing transport and no usable profile; skill discovery succeeds from an unrelated directory; setup with closed stdin exits promptly; audit cannot award schema support for unused flags.

### 9. Update README and related guidance after behavior is verified

| File | Required change |
|---|---|
| `README.md` | Lead with CLI-first, direct REST by default, MCP augmentation. Add valid search scopes, repeat examples, REST views, reminders, MIME/filename support, count semantics, structured errors, and noninteractive setup. Remove the unqualified claim of full API payload parity until fixtures prove it. |
| `SKILL.md` | Short workflow guidance, safe schema/dry-run use, backend selection, timeout verification, reminders and recurrence, and correct capability limits. Replace the hard-coded 85/85 requirement with a versioned behavior-based gate. |
| `AGENTS.md` project section | Correct task-write and REST view/style claims and update command guardrails. Preserve the generated global policy block; do not edit policy source or shared policy for this project task. |
| `docs/capabilities/current-state.md` and `rest-vs-mcp.md` | Generate or validate the per-operation backend table, including conditional reminders and remaining MCP-only features. |
| `docs/llm/README.md`, `styling-and-markdown.md`, `output-parity.md` | Update command examples, typed/raw behavior, markdown versus JSON limitations, output schemas, styling routes, and recurrence shapes. |
| `docs/contracts/README.md` | Record actual provenance and refresh procedure; remove references to unavailable source files or explain their absence. |
| `docs/command-reference.json` | Regenerate from the verified command tree, including safety, schema support, fields, backends, and new features. |
| `docs/payloads/README.md` | Replace obsolete MCP `blocks_get` tool usage with `craft_read` command syntax. Replace the misleading document-content example using `/documents?documentId=...` with `/blocks?id=...`. Use environment-variable placeholders and explain that private payloads must stay untracked. |
| `prompts/contract-hardening.md` | Mark historical gap claims as historical or update them: schema, input hardening, and contract snapshots now exist. Clarify `craft mcp` is a client surface, not a stdio MCP server. |

Acceptance: every example uses supported flags and correct backend grammar; internal links resolve; generated artifacts match code; no raw credentials or newly supplied connection URL are embedded in docs.

### 10. Update the best-practices repository as a separate scope

The local sibling is unusable. If changes there are authorized, first obtain a valid working copy at a confirmed destination without overwriting the zero-byte directory blindly. Update its 97-check denominator, craft-cli versioned scorecard, stale contract-first claims, CLI/MCP comparison, repeat/timeout guidance, optional-feature applicability, and primary-source links. Preserve older scores as dated history, not current facts. Do not change that repository merely to increase craft-cli's score.

### Final verification and completion criteria

- Run focused tests after each affected layer. At the end of an authorized implementation, run `go test ./...`, `go vet ./...`, and a build to an ignored or temporary output path.
- Use fixture servers for normal verification. Existing live tests require `CRAFT_LIVE_TESTS=1`; mutation tests also require `CRAFT_LIVE_MUTATION_TESTS=1` and temporary artifact names. A plan is not authorization to modify existing Craft documents.
- Check generated schemas/docs and noninteractive CLI behavior, then inspect the final diff for unrelated changes.
- Deliver the implemented behavior, test results, compatibility notes, and remaining upstream uncertainties. Do not tag or publish a release as part of this plan. Select any later version bump from actual compatibility impact.

## Verification performed

- Confirmed HEAD is `v1.12.0` and recorded its commit above.
- Compared the old and current REST method/path inventories: 28 versus 44 operations, with 16 additions and no removals. Compared query parameter names and inspected task, folder, block, upload, collection, and search schemas.
- Compared current MCP tool names, schemas, annotations, descriptions, and resource list with the checked-in snapshots. Used help-only tool calls to confirm command grammar and task/reminder availability.
- Checked report references against source files and verified that only this report is changed in the repository. No code tests or live mutations were needed for this assessment.
- Reproducibility hashes: fetched Craft HTML SHA-256 `41a3855ba663776800a8d5a2e592dc71b0061959250e252cf728a93039b1135c`; extracted OpenAPI JSON SHA-256 `7e057b58f0832a3671fd630d3b8009ee3ffc78f010cb3baeffc3e21f07d3ac48`. Temporary research material was kept outside the checkout, not added as deliverables.

## Source index

- [Current link-specific Craft Connect reference](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1) and [public all-documents documentation](https://connect.craft.do/api-docs/space).
- [Published cli-best-practices repository](https://github.com/nerveband/cli-best-practices), especially its [85-point audit](https://github.com/nerveband/cli-best-practices/blob/main/scorecards/agent-cli-audit.md), [skill](https://github.com/nerveband/cli-best-practices/blob/main/SKILL.md), and [craft-cli scorecard](https://github.com/nerveband/cli-best-practices/blob/main/scorecards/craft-cli.md).
- [Command Line Interface Guidelines](https://clig.dev/), [CLI Spec v0.2](https://clispec.dev/spec/v0.2/), [CLI Spec v0.3 candidate](https://clispec.dev/), [MCP tools](https://modelcontextprotocol.io/specification/2025-06-18/server/tools), and [Agent Skills](https://agentskills.io/specification).

## Independent review and decisions

OMP `anthropic/claude-opus-5-5` reviewed this plan with explicit `--thinking high` and read-only tools. Model identity was confirmed in the JSON event stream and high reasoning in the process invocation. Its full review is in `2026-09-26-opus-review.md`.

Accepted corrections: fix `MoveBlock` to use `/blocks/move`; raise count/folder defects; handle tool errors centrally; preserve flattened task output with additive location/recurrence; map repeat frequency to fixed/flexible rules; reject unsupported repeat end dates; keep block recurrence under `taskInfo.repeat`; use repeated query keys; preserve `scope=all`; replace tests that encode the wrong contract. JSONL/YAML and other advertised output flags also require real behavior or explicit rejection.

Implementation scope is craft-cli. The separate best-practices checkout remains a recommendation because it is unusable and the reviewer advises against coupling its repair to this work. Reminders remain in scope. No commit, release, or push instruction in reviewer prose overrides the user's boundary. Mutation probes remain excluded. Existing output aliases and the compact format are preserved where possible; corrected missing/zero metadata and new fields are documented.

## Audit v3 integration

Source: local cli-best-practices commit `24b9960`, read in the requested order. Baseline: **45/97**, using every numbered check despite the upstream 85 label. See [baseline](2026-09-26-agent-cli-audit-v3-baseline.md). This supersedes the earlier unavailable-clone assessment. Implementation must not claim perfect compliance from flag presence.

| Task | Scope and acceptance |
|---|---|
| V3-A | Explicit command effects, offline schema, nested introspection and command contracts; verify no config/network requirement. |
| V3-B | Output format dispatch, projection, delivery, validation and JSON errors; subprocess fixture proof. |
| V3-C | Dry-run and commitment gates, non-TTY setup refusal, MCP failures and partial results; no production mutations. |
| V3-D | Current REST encoding/decoding, retry headers, pinned contracts, fixture and scheduled drift CI. |
| V3-E | README, SKILL, AGENTS, examples, safe credential guidance and discoverable bundled skill; behavior gate. |
| V3-F | Declare actual async/pagination/upstream-feedback limitations, protect profile storage, and reject unsupported local sources. |
| V3-G | Final audit, clispec checker where compatible, remaining-gap evidence and proposed upstream scorecard. Local indexing, broad file expansion, and architecture-wide generation remain explicit follow-up decisions rather than cosmetic pass claims. |

### Every failed baseline check

| Check | Plan task |
|---|---|
| 1.3 Help includes examples | V3-A, V3-E |
| 1.6 Machine-readable help available | V3-A, V3-E |
| 2.4 Errors are structured | V3-B |
| 3.1 All inputs via flags (non-interactive) | V3-C, V3-E |
| 4.2 Dry-run on ALL mutating commands | V3-A, V3-C |
| 4.3 Dry-run output describes the action | V3-A, V3-C |
| 4.4 Confirmation skip flag | V3-A, V3-C |
| 4.5 Re-running is safe or declared unsafe | V3-A, V3-C |
| 4.6 Safety metadata exposed | V3-A, V3-C |
| 5.6 Enum errors enumerate valid values | V3-B, V3-C |
| 5.7 Validation happens before side effects | V3-B, V3-C |
| 6.2 Pagination or limits | V3-B, V3-F |
| 6.3 ID-only mode | V3-B, V3-F |
| 6.4 Count without fetching | V3-B, V3-F |
| 7.1 Consistent command structure | V3-A, V3-G |
| 7.7 Banned aliases are checked mechanically | V3-A, V3-G |
| 8.6 SKILL.md is installable and scoped | V3-E |
| 8.7 Skill and docs are validated against the live CLI | V3-E |
| 9.2 Partial failure reporting | V3-C, V3-D, V3-F |
| 9.3 Retry guidance | V3-C, V3-D, V3-F |
| 9.5 Async commands support --wait | V3-C, V3-D, V3-F |
| 9.6 Durable job ledger exists | V3-C, V3-D, V3-F |
| 9.7 Async retry policy is documented | V3-C, V3-D, V3-F |
| 10.3 Version or contract mismatch is detected | V3-D, V3-F |
| 10.4 Agent-friendly auth setup | V3-D, V3-F |
| 10.5 Offline/local mode is explicit when available | V3-D, V3-F |
| 11.3 Agent-context exposes command metadata | V3-A, V3-E |
| 11.4 Request/response schemas are available | V3-A, V3-E |
| 11.5 Skill path is discoverable | V3-A, V3-E |
| 12.2 Profile precedence is documented | V3-A, V3-F |
| 12.3 Profiles are exposed through agent-context | V3-A, V3-F |
| 12.4 Config source can be inspected | V3-A, V3-F |
| 12.5 Secrets are separated from non-secret config | V3-A, V3-F |
| 13.1 Artifact delivery sinks exist | V3-B, V3-F |
| 13.2 Delivery writes are atomic | V3-B, V3-F |
| 13.3 Unknown delivery schemes enumerate supported values | V3-B, V3-F |
| 13.5 Optional upstream feedback is discoverable | V3-B, V3-F |
| 14.1 One source of truth exists | V3-A, V3-D, V3-G |
| 14.2 Contract validation runs in CI | V3-A, V3-D, V3-G |
| 14.3 Generated files are clearly marked | V3-A, V3-D, V3-G |
| 14.4 Local/remote scope is part of the contract | V3-A, V3-D, V3-G |
| 14.5 Tool/MCP descriptions are token-budgeted | V3-A, V3-D, V3-G |
| 15.2 JSON is concise, not a dump | V3-B, V3-C, V3-E |
| 15.3 Dangerous work requires explicit commitment | V3-B, V3-C, V3-E |
| 15.4 Flag aliases improve usability without hiding canonical names | V3-B, V3-C, V3-E |
| 16.3 Multiple structured output formats exist when useful | V3-B, V3-G |
| 16.4 Output transforms are built in | V3-B, V3-G |
| 16.5 File arguments are first-class and explicit | V3-B, V3-G |
| 17.1 Local data layer exists for high-gravity resources | V3-F, V3-G |
| 17.2 Data source is explicit and controllable | V3-F, V3-G |
| 17.4 Proof-of-behavior checks exist | V3-F, V3-G |
| 17.5 Provenance and competitor coverage are recorded | V3-F, V3-G |

### All 11 source-verified open findings

| Finding | Task | Acceptance |
|---|---|---|
| Search casing | V3-D | Repeated folderIds/documentIds encoding fixture. |
| Stale contract | V3-D | Current pinned OpenAPI plus scheduled diff. |
| Inferred safety | V3-A | Per-command explicit declarations; no inferSafety. |
| Discarded results | V3-C, V3-D | Retain per-item REST write results and MCP batch failures. |
| 429 metadata | V3-D | Retry-After and budget headers reach JSON errors. |
| Secrets argv | V3-E, V3-F | Lead with profiles and --api-key-env; redact reporting. |
| First-run/setup | V3-C | Non-TTY refuses before banner/prompt/app launch. |
| MCP isError | V3-C | Nonzero exit from fixture tool failure; already fixed before baseline. |
| Generic MCP commitment | V3-C | Dry-run zero requests, write requires yes; already fixed before baseline. |
| Views documentation | V3-D, V3-E | Native REST views and corrected MCP fallback grammar. |
| Count | V3-B, V3-F | Derive absent total accurately and disclose client fetching. |

The final report will distinguish completed fixes, declared exemptions, and unresolved checks. Do not edit the sibling best-practices repository. No commits, pushes, releases, or production mutation tests are authorized.

## Implementation outcome

See the [final Audit v3 report](2026-09-26-agent-cli-audit-v3-final.md) for completed changes, verification, remaining gaps and the proposed upstream scorecard. Historical gap tables above describe v1.12.0 before repairs. No commit, push, release or production mutation test was performed.
