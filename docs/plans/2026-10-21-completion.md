# Craft CLI completion plan

Date: 2026-10-21. Scope: current published Craft API and available configured REST/MCP connections, plus the remaining Audit v3 implementation backlog. Preserve existing commands and raw upstream outcomes. Do not claim unreleased or inaccessible features are implemented.

## Implementation order

1. Refresh evidence: compare the configured all-documents (previously called beta) contract and the three public connection contracts. Inspect available MCP tools and command help. Inventory every REST operation against adapters. Record source, date, availability, and any unavailable checks.
2. Identity: separate credential material from ordinary profile metadata with atomic writes and legacy read compatibility. Add a redacted effective configuration report and runtime agent context. Keep offline schema independent of user configuration.
3. Input and command ergonomics: add noninteractive setup, document resource aliases, realistic leaf examples, and bounded explicit embedded JSON/text/base64 file expansion with literal escaping. Preserve the destructive confirmation boundary.
4. Local data: add an opt-in per-connection private document cache, incremental refresh based on modification metadata, local list/get/search, explicit freshness/provenance, and safe offline failure. Bound normal list/get output with explicit unlimited options and truncation metadata. This is a breaking default change and requires a major release.
5. Contracts: define CLI input and output contracts from the registered command tree, typed models, REST snapshots and MCP discovery. Keep CLI output distinct from upstream schemas. Cover conditional inputs, alternate outputs, variable upstream results, source/scope metadata and realistic examples. Validate generated artifacts and adapters in CI.
6. Verification: unit and fixture tests for migration failures, cache isolation/freshness/deletion, input expansion, setup, aliases, contract completeness and compatibility. Run Go tests/vet/build, agent subprocess checks, schema validation and live REST/MCP reads. Exercise mutations only on temporary craft-cli-live-test artifacts, preview deletion/moves/clears first, and verify outcomes before retries.
7. Reconcile every historical audit failure with implementation evidence. Document genuine upstream limitations without calling them completed features. Release only after checks pass, then verify the published archives and upgrade this host through craft upgrade.

## Completion record

Implementation is complete. Historical reports remain evidence for their dated snapshots, not current scores.

| Historical gap | Implementation evidence |
|---|---|
| 1.3, examples | Every executable command has a concrete example, checked against its parser and argument contract. |
| 3.1, setup | Flag-based setup supports REST/MCP without a terminal. The optional wizard still requires a terminal. |
| 7.1, vocabulary | Documents resource aliases share the existing handlers and flags. Top-level commands remain available. |
| 11.3, 11.4, contracts | CLI input/output schema flags, model-derived output fields, conditional raw-input modes, dynamic MCP boundaries, output variants and fixture validation. Upstream REST schemas remain separate. |
| 12.3, runtime profiles | `context` exposes safe profile metadata without making offline schema read configuration. |
| 12.4, setting sources | REST resolution and redacted effective source reporting share a resolver; MCP precedence is reported separately. |
| 12.5, credentials | Atomic credential generations separate URLs/keys from metadata, preserve legacy reads, and keep failed metadata publication recoverable. Explicit migration and corrupt-config reset are tested. |
| 14.1, contract discipline | Registered Cobra commands, effect declarations, serialization models, current REST/MCP captures and example fixtures feed generated discovery. CI checks generated parity and behavior. The CLI remains handwritten; MCP owns variable response schemas. |
| 14.4, provenance | Normal JSON objects identify scope, source and backend, plus snapshot age for index reads. Raw streams and scalar projections retain their stream contracts. |
| 15.2, bounded output | Document list defaults to 50, get/block/item depth to 2; other read arrays and text fields have limits with metadata and explicit unlimited flags. This requires 3.0. |
| 16.5, embedded input | Opt-in JSON/text/base64 expansion, literal escaping, regular-file checks, UTF-8 validation, one-value JSON parsing and a total 8 MiB budget. |
| 17.1, 17.2, local data | Private per-connection term index, modification-based refresh, checkpoints, deletion reconciliation, explicit local/live/auto sources and cache inspection/removal. It uses JSON snapshots rather than SQLite. |

### API evidence

- The previously labeled beta all-documents link matches the current pinned 44-operation contract, SHA256 `db9adc89da6e13c5f27943608067fe2a1a84cd46da8856ce83c97edbd1e09dc2`. It is the same link-specific contract, not a separately versioned beta API.
- Reviewed [All Documents](https://connect.craft.do/api-docs/space), [Daily Notes](https://connect.craft.do/api-docs/daily-notes), and [Selected Documents](https://connect.craft.do/api-docs/documents/). Direct extraction of these generic pages returned HTTP 403; the web reader provided their rendered references. Context7 returned unrelated Craft CMS libraries, so no Craft.do claims rely on those results.
- Added explicit date and selected-document search targets and daily-note task placement. Existing raw REST input covers connection-specific fields without pretending every connection has identical availability.
- MCP discovery succeeded on the configured MCP connection. The three tool names and structural input contracts remain compatible; current descriptions were refreshed. Folder metadata update help verified name/icon/tint grammar, now exposed natively.
- Actual MCP task create and update both returned only text outcomes, with no revert payload. The temporary task was deleted after a preview. Task undo remains an explicit upstream limitation and fails before writing.
- Initial expanded live runs found and fixed the root-folder move destination object, typed-nil JSON bodies on bodyless GET and DELETE operations (reminder/view reads and view deletion), collection-item writes that omitted required empty properties and always sent an optional mutation flag, and attachment uploads that selected text/plain instead of the documented binary transport. Text bytes uploaded successfully with application/octet-stream on the same connection, and a second test connection also passed binary upload. Temporary documents/folders were cleaned up after each failed run. One reminder left by an interrupted check was also identified by its exact test naming pattern and deleted after a preview. Transport timeout diagnostics now omit capability URLs while retaining the underlying timeout cause. Batch document creation now stops on failures and reports known partial writes, and malformed structured MCP batches fail locally rather than falling back to command text.

### Boundaries

No async API, upstream feedback service or user-owned MCP server is invented. Experimental reminders and connection-specific endpoint availability remain upstream conditions. The frozen CLI Spec v0.2 discovery format is retained; the current v0.3 standard was inspected, but migrating that standard is separate from API parity and must not be claimed by relabeling the manifest. Exact arbitrary MCP payload fields cannot be fixed statically; they are declared dynamic and preserved.

## Final pre-release verification

| Check | Result |
|---|---|
| Go tests, vet and build | Passed with Go 1.27.3 |
| Offline behavior gate | 520 subprocess checks passed; no production mutations |
| CLI contracts and examples | Passed for all 116 registered executable commands |
| Frozen CLI Spec v0.2 | Schema validation passed |
| Live REST coverage | 67 CLI checks passed across all 44 pinned operations; zero unavailable checks and zero cleanup errors |
| Native MCP folder metadata | Name/icon/tint write passed, verified through REST, then temporary folder deleted after preview |
| MCP discovery | Current tools, resource metadata, protocol 2025-06-18 and server 1.0.0 verified |
| MCP task undo boundary | Live create/update returned no revert payload; temporary task deleted |
| Upgrader | Go 1.27.3 throwaway binary at 2.1.1 upgraded itself to published 2.1.2 |
| Dependency and tool versions | No direct module updates; latest Go stable pinned; current Actions majors and jsonschema confirmed |
| Release configuration | GoReleaser check passed |
| Active documentation | Relative links, skill parity and punctuation checks passed |

Release target: 3.0.0. Publication, archive checksums, GitHub CI, local `craft upgrade`, credential migration and final connectivity checks run after this release commit. Their results belong to the published release record.


The project toolchain is pinned to Go 1.27.3, verified against the [official stable download inventory](https://go.dev/dl/?mode=json). The Go language minimum remains 1.26.0. No newer direct module dependencies were reported. The pinned CI validator jsonschema 4.26.0 is also the latest published release. Profile connection checks now preserve failure exit codes and respect the requested timeout.
