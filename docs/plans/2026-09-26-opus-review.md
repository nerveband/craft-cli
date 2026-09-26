# Review: craft-cli audit and implementation plan

**Verdict: not ready to implement as written. It will be ready once the corrections below are made.** Most of the audit's findings match the source, the current OpenAPI and the MCP fixtures. The plan has four problems:

- It misses a P0 bug in block move.
- It rates several bugs lower than it should.
- It leaves the output compatibility decisions open.
- It includes work that will slow down the P0 fixes.

The safety fixes (schema flags, dry-run and `isError`) can start now. They don't depend on the other corrections.

## Findings confirmed against source and fixtures

- **Operation count:** the current OpenAPI has 44 operations across 21 paths, including 5 view operations, 4 reminder operations and 5 whiteboard operations.
- **Search parameters:** the client sends `folderIDs` and `documentIDs` at `internal/api/client.go:1459-1464`. The contract uses `folderIds` and `documentIds` at OpenAPI lines 52554 and 52603. `fetchBlocks` exists and `fetchMetadata` is gone from the search query.
- **Page block title:** in page blocks, `title` is a required object (OpenAPI lines 568 and 1372-1376). `Block.Title` is a string at `internal/models/document.go:45`.
- **Repeat on task endpoints:** `repeat` is a sibling of `taskInfo`, not a field inside it (OpenAPI lines 58729, 59131 and 59454). `tasks` `scope` is `required: true`, and `GetDocumentTasks` sends only `documentId`.
- **Folders:** create uses `parentFolderId` (OpenAPI line 46432). The list response nests `folders[]`.
- **No `isError` handling:** the string `isError` appears nowhere in `cmd/` or `internal/`. Fixture `help-20` returned `isError: true` for `collections views`.
- **Schema flags do nothing:** `--request-schema` and `--response-schema` are only declared (`cmd/root.go:44-45,151-152`). The only other use is the presence check in `audit.go:451`.
- **`skill-path`:** it uses `os.Getwd()` (`cmd/agent_support.go:19-23`).
- **Tasks and MCP:** MCP help fixtures 13 and 23 show task add, update and delete. That contradicts `cmd/tasks.go:317-326` and `AGENTS.md`.

## Prioritized corrections

| # | Priority | Correction | Evidence |
|---|---|---|---|
| 1 | P0 (missed) | **`craft blocks move` calls the wrong endpoint with the wrong body.** It sends `PUT /blocks` with `{blocks:[{id, position:{pageId, position}}]}`. The contract is `PUT /blocks/move` with `{blockIds:[...], position:{...}}`. The move either fails or does nothing while the CLI prints "moved". No test covers it, and `prompts/api-upgrade-map.md:26` marks it "Covered". | `internal/api/client.go:876-915`, `cmd/blocks.go:443`, OpenAPI 23828-23990 |
| 2 | P0 (raise from P1) | **`--count` always outputs 0 for `list` and document `search`.** The OpenAPI has no `total` field anywhere. `outputCount(result.Total)` is at `cmd/list.go:71-72` (the audit cites 65-66) and `cmd/search.go` in `runDocumentSearch`. `AGENTS.md` presents `--count` as a guardrail, so agents get wrong answers. JSON payloads also emit `total: 0`. | grep for `"total"` in the OpenAPI finds nothing |
| 3 | P0 (raise from P1) | **Subfolders are invisible, and folder create reports a parent that was never sent.** `Folder` has no `folders` field. `CreateFolder` sends `parentId` but echoes `ParentID: parentID` back as success. | `internal/api/client.go:668-715`, `internal/models/document.go:124-130` |
| 4 | P0 | **`isError` handling must cover every MCP path, not just `mcp call` and `batch`.** Also covered: `runMCPReadCommand` (`cmd/mcp.go:340-350`, used by `images view`, `whiteboards elements get`, `explore-*`), `runMCPCollectionCommand`, and `runMCPWriteCommand`/`completeMCPWrite` (`cmd/mcp_write.go:12-56`). The last one means `--save-revert` runs on a failed write. Do this in one place: have `CallTool` return a typed tool error. | `cmd/mcp_write.go:52-56` |
| 5 | P0 | **Decide the page-title fix from the contract, but confirm it with one read before rating the impact.** Both snapshots require `title` as an object on page blocks. If the server matches, `GetBlock` and `GetBlockWithOptions` (which decode the root into `models.Block`) and any nested page in `get` fail with a decode error. That would already break common commands, so the severity needs a read-only probe on the public test link. The fix is the same either way: a custom unmarshaler or a `json.RawMessage` title that also accepts rich-link string titles. | `internal/api/client.go:788-800,1745-1763` |
| 6 | P0 | **Repeat mapping needs a concrete table, and the block path must stay unchanged.** On `/blocks`, repeat stays under `taskInfo.repeat` (OpenAPI around 440). Only `/tasks` moves it to the item level, so a shared `models.TaskInfo` must not be changed for both. Proposed mapping: `--repeat X` sets `frequency: X` and `type: fixed`. `--repeat-dynamic-days` sets `type: flexible` (not `monthly.dynamicDays`, which is an array of objects). `--repeat-weekdays 0..6` maps to `weekly.days` names. `--repeat-skip-weekends` maps to `daily.skipWeekends` or `monthly.skipWeekends`. `--repeat-reminder HH:MM` maps to `reminder{enabled:true, dateOffset:minutes}`. **`--repeat-end` has no contract field** (only `startDate` exists), so reject it with a migration hint. `--repeat-frequency` duplicates `--repeat`, so reject any conflicting combination. | OpenAPI 59454-59640, `cmd/tasks.go:302-389` |
| 7 | P0 | **Remove tests that pin the wrong contract.** `internal/api/endpoints_test.go:218` asserts `folderIDs`. `internal/api/task_repeat_test.go:17-60` asserts repeat inside `taskInfo`. `cmd/tasks_repeat_test.go:61` only checks that flags are registered, which proves nothing. Replace them with request-shape tests based on the contract. | |
| 8 | P1 | **State the array encoding instead of "verify against examples".** The parameters declare no `style` or `explode`, so the OpenAPI 3.1 default is form with explode: `folderIds=a&folderIds=b`. Use `url.Values.Add`. Treat live acceptance as unverified. | grep for `explode`/`style` finds nothing |
| 9 | P1 | **Fix the pagination claim.** REST `/documents` and search have no `limit` or `cursor`. Only `/reminders` does (OpenAPI 45803-45814). "Use server pagination where Craft offers it" does not apply to `list`. Document that `--limit` truncates on the client, and point to the existing MCP cursor route (`cmd/list.go:57`). | |
| 10 | P1 | **Make output compatibility decisions before step 4.** The `Task` output is flattened (`state`, `scheduleDate` at the top level). Recommended: keep the flattened fields as the CLI contract and add `location` and `repeat`, so the change is additive (minor bump). Emitting the raw server shape would be a breaking change. Also, `Document` always emits `createdAt`/`lastModifiedAt` as `0001-01-01T00:00:00Z` when `--metadata` is absent, plus `hasChildren:false` and an empty `spaceId`. Make these fields omit or pointer (a fidelity fix the plan misses). | `internal/models/document.go:6-18,144-156`, `cmd/output.go:99-101,218` |
| 11 | P1 | **`tasks list` defaults need a decision.** The default `--scope all` now means every task block, including logbook tasks, so results can be very large. `--document` silently ignores `--scope`. The help text leaves out `document`. Recommendation: send `scope=document` when `--document` is set, and document the meaning of `all`. Changing the default would be a breaking change, so only do it at a major version. | `cmd/tasks.go:52-55,280` |
| 12 | P1 | **`AGENTS.md` recommends an MCP wrapper that duplicates a REST command.** `craft whiteboards elements get` goes through MCP, while `craft whiteboards get` uses REST `GetWhiteboardElements`. Under a REST-first design, `AGENTS.md` should recommend the REST command. Keep the MCP wrapper only as an explicit alternative. | `cmd/whiteboards.go:38-51,131` |
| 13 | P1 | **Upload details.** The OpenAPI `requestBody` declares only `application/octet-stream`. Detecting type from `Content-Type` comes from the `fileName` description text, not the schema, so treat MIME handling as unverified until probed. Also: `--sibling` with the default `--position end` sends an invalid combination (validate it), and the 30-second timeout with a whole-file `ReadAll` makes large uploads end with an unknown outcome. | `cmd/upload.go:117-137`, OpenAPI 59884-59980 |
| 14 | P2 | **Smaller correctness items the audit misses:** `include=` is always sent, even for regex-only searches (`client.go:1450`). Block search passes the literal query as an RE2 pattern, so input like `C++` breaks (use `regexp.QuoteMeta` unless `--regex` is set). The 404 handler discards the server's message (`client.go:158`). `AddBlock` and `AddBlockRelative` have no callers and use an invalid position shape; delete them. | |
| 15 | P2 | **MCP transport note.** The tools list includes `execution.taskSupport`, a field from a newer MCP revision than the requested `2025-06-18`. The CLI's own Go client, with its default User-Agent, was never run against the endpoint; only curl was. Given the Python 403, run `craft mcp tools` read-only once before claiming compatibility. | `/tmp/craft-audit-mcp-tools-list.json` |

## Sequencing and scope

- **Step 1 is too heavy to block the P0 fixes.** Split it:
  - Now: commit the fetched OpenAPI as a fixture and add an offline check of the operation inventory.
  - Later: the refresh script, hashing and CI diffing.
- **Recommended order:**
  1. Safety fixes: schema flags, dry-run and `isError`. They are small and prevent unintended writes.
  2. The REST P0s: move, search, tasks, folders and count, with the wrong-contract tests replaced.
  3. Output fidelity, with the compatibility decisions from row 10.
  4. REST collection views.
  5. Documentation.
  6. Reminders, which are experimental and depend on the connection.
- **Defer or cut:**
  - Configurable deadlines and a retry policy. Keep only the `Retry-After` metadata and the uncertain-outcome error for writes.
  - The CLI Spec adapter and the audit-scoring rewrite.
  - Comparing REST style fields one by one.
  - Hardening SSE for multiple events: the server currently returns one event, so only match on the response ID.
  - Discovery pagination.
  - Step 10, the separate `cli-best-practices` repository.
- **Do not add a `blocks learn` or `folders update` wrapper** until someone needs them. The generic `mcp call` covers both once row 4 is fixed.

## Claims I could not verify

- The checked-in snapshot's count of 28 operations. The file is minified to a single line and I had no shell to count. The 44-operation count for the current OpenAPI is verified.
- Whether the live server:
  - fails to decode page titles,
  - returns an undocumented `total`,
  - accepts repeated `folderIds` keys, or
  - honors the upload `Content-Type`.

  All of these come from the contract only. One read-only request to the public test link would settle each.
- Folder delete semantics (the upstream text contradicts itself).
- The `cmd/setup.go` behavior without a TTY, `docs/payloads/README.md`, and the `docs/llm` line references.
- Everything about the `cli-best-practices` repository (the count of 97 checks, the scorecard version), CLI Spec v0.2/v0.3, the MCP 2026 release notes, the Python 403 versus curl result, and the SHA-256 hashes.

## Revised implementation checklist

1. Commit the current OpenAPI as a fixture and add an offline check that the inventory has 44 operations. Remove the dead `docs/HHRuPxZZTJ6-*` references from `docs/contracts/README.md`.
2. Make `--request-schema` and `--response-schema` either output a schema before authentication or network access, or fail. Make `mcp call` respect `--dry-run` and require `--yes` for `craft_write` and `blocks_revert`. Map `isError` to a typed error in `internal/mcp/client.go` so every MCP path gets a nonzero exit, and so do batches. Do not save revert info for failed writes.
3. Point `blocks move` at `PUT /blocks/move` with the `blockIds`/`position` shape. Delete `AddBlock` and `AddBlockRelative`.
4. Search: use `folderIds` and `documentIds` as repeated keys, add `fetchBlocks`, drop `fetchMetadata` from document search, and omit an empty `include`.
5. Tasks: repeat at the item level with the mapping from row 6, and reject `--repeat-end`. Send `scope=document` with `--document`. Add clearing via explicit `null`. Decode `taskInfo`, `location` and `repeat` into the flattened output with `location` and `repeat` added.
6. Folders: send `parentFolderId` and output the nested `folders` tree.
7. `--count` counts the returned items and says it covers only what the API returned. Remove the synthesized `total: 0` and the zero-value document timestamps.
8. Decode the page title union; add fixtures for the root page, nested pages and rich-link titles.
9. Preserve mutation `items` where the CLI throws them away today. Add `Retry-After` and the rate-limit budget headers to `APIError`. Report timeouts on writes as uncertain outcomes.
10. Implement the five REST collection-view endpoints behind the existing commands. Fix the MCP view grammar only if an explicit `--backend mcp` path is kept.
11. Upload: send `fileName`, detect or accept the MIME type, and validate the sibling/position combination.
12. Replace the tests from row 7. Add request-shape and decoding tests from the fixture for rows 1-8, plus MCP `isError` and dry-run tests with a fake transport.
13. Update `README.md`, `SKILL.md`, the project `AGENTS.md` (task writes, REST views, whiteboards get, count semantics), `docs/capabilities/*`, `docs/llm/*` and `prompts/api-upgrade-map.md`.
14. Run `go test ./...`, `go vet ./...` and a build to a temporary path. Run read-only probes on the public test link for the unverified items above. Choose the version bump once the compatibility decision in row 10 is made.