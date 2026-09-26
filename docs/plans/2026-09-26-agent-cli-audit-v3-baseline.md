# craft-cli Audit v3 baseline

Date: 2026-09-26. CLI version: 1.12.0, HEAD `ac51d02`, working tree with partial implementation changes. Audit source: local `cli-best-practices` commit `24b9960`.

**CLI Spec status: not published.** Existing `craft schema` is a proprietary manifest.

## Scope and scoring

This is the requested pre-continuation baseline, captured after some previously authorized edits, not a pristine v1.12.0 score. Those edits include MCP isError/commit guards, refreshed contracts, initial schema inspection, and model decoding/count corrections. No further code was changed before this audit. Build succeeded. Offline tests failed in repeat serialization tests because the in-progress model changes and old tests disagree.

The supplied Audit v3 enumerates **97 checks**, not 85. Category maxima sum to 97; categories 1-10 sum to 62, not the claimed legacy 50. Every numbered check below is scored once. No twelve checks were silently discarded and no score was scaled to 85. The repository was read locally in the requested order and was not modified.

## Scorecard

| Category | Score | Max |
|---|---:|---:|
| Discoverability | 5 | 7 |
| Structured output | 5 | 6 |
| Input flexibility | 4 | 5 |
| Safety rails | 1 | 6 |
| Error handling | 5 | 7 |
| Context window discipline | 2 | 5 |
| Predictability | 5 | 7 |
| Agent knowledge | 5 | 7 |
| Resilience | 2 | 7 |
| Distribution and lifecycle | 2 | 5 |
| Three-layer introspection | 2 | 5 |
| Persistent identity and configuration | 1 | 5 |
| Two-way I/O and artifacts | 1 | 5 |
| Contract and generation discipline | 0 | 5 |
| Unix composability and agent restraint | 2 | 5 |
| API-native payload ergonomics | 2 | 5 |
| Domain depth and proof gates | 1 | 5 |
| **Total** | **45** | **97** |

## Exemptions

| Check | Declared reason | Evidence |
|---|---|---|
| None | No qualifying exemptions established | Silent non-applicability is scored as failure. |

## Evidence per check

| Check | Result | Evidence |
|---|---|---|
| 1.1 Root help lists subcommands | PASS | Built binary root help lists commands and descriptions. |
| 1.2 Every subcommand has --help | PASS | All 111 paths enumerated from schema return help with exit 0. |
| 1.3 Help includes examples | FAIL | Only 37/111 help pages contain Examples, below 80%. |
| 1.4 Examples use realistic values | PASS | Help examples use document IDs, dates, names and concrete payloads. |
| 1.5 Progressive disclosure works | PASS | Root/group/leaf help works; deeper groups are available through root help paths. |
| 1.6 Machine-readable help available | FAIL | schema returns JSON but lacks positional input types and still initializes config; no CLI Spec document. cmd/schema.go, cmd/root.go. |
| 1.7 Version is queryable | PASS | version exits 0: craft-cli version 1.12.0. |
| 2.1 JSON output available | PASS | list --format json fixture emits JSON. |
| 2.2 JSON is consistent across commands | PASS | list, folders list, tasks list emit items envelopes with totals; get is a single object. |
| 2.3 JSON is the default when piped | PASS | Piped fixture commands default to JSON without ANSI. |
| 2.4 Errors are structured | FAIL | JSON errors require a separate flag, contain error not message, and JSON success selection alone still emits text errors. |
| 2.5 Exit codes are meaningful | PASS | Invalid command exits 1; HTTP 401/404/429/500 and MCP tool failure exit 2. |
| 2.6 Quiet mode suppresses noise | PASS | Read data stays on stdout; search status goes to stderr. |
| 3.1 All inputs via flags (non-interactive) | FAIL | setup requires interactive input; explicit profile alternatives exist but command itself cannot complete via flags. cmd/setup.go. |
| 3.2 Stdin accepted for structured input | PASS | delete --json preview accepts API-shaped structured input; blocks and other writes also support stdin. |
| 3.3 Secrets without argv | PASS | Saved profiles and profiles add-rest --api-key-env provide non-argv credentials. cmd/profiles.go. |
| 3.4 Flags over positional args | PASS | Resource commands use flags with primary IDs; batch IDs are deliberate extensions. |
| 3.5 Raw payload passthrough | PASS | Raw REST body paths exist for documents, tasks, blocks, collections, comments. cmd/raw_payload.go. |
| 4.1 Dry-run exists | PASS | delete fixture --dry-run returns intent; only GET reached mock. |
| 4.2 Dry-run on ALL mutating commands | FAIL | config mutations and feedback ignore dry-run; generic MCP call is already fixed before this baseline. |
| 4.3 Dry-run output describes the action | FAIL | delete preview has target/reversible but no local/server validation provenance. |
| 4.4 Confirmation skip flag | FAIL | delete without --yes performs DELETE in non-TTY fixture; setup lacks TTY guard. |
| 4.5 Re-running is safe or declared unsafe | FAIL | Timeout guidance exists but idempotent writes do not report no-change; contract declarations missing. |
| 4.6 Safety metadata exposed | FAIL | cmd/schema.go inferSafety uses leaf names. |
| 5.1 Errors are actionable | PASS | Missing get argument yields exact missing count plus --help hint. |
| 5.2 Fail fast on missing input | PASS | Missing required arguments exit immediately; separate first-run setup defect recorded under 4.4. |
| 5.3 Network errors are distinct | PASS | API/network errors map to 2 versus validation 1. cmd/root.go. |
| 5.4 Error includes hint for recovery | PASS | Mock 401 includes AUTH_ERROR and recovery hint. |
| 5.5 Errors to stderr, data to stdout | PASS | Failing probes emit errors on stderr and empty stdout. |
| 5.6 Enum errors enumerate valid values | FAIL | Invalid --format reports only unsupported format, omitting accepted values. |
| 5.7 Validation happens before side effects | FAIL | Invalid --format performs a request before failing; --data-source local is ignored. |
| 6.1 Field selection available | PASS | list --fields id returns only IDs in items. |
| 6.2 Pagination or limits | FAIL | REST --limit fetches all then truncates; no declared REST pagination exemption in help. |
| 6.3 ID-only mode | FAIL | list --id-only returns full objects in default JSON mode. |
| 6.4 Count without fetching | FAIL | count is correct (2) after earlier decoder edit, but client-side fetching is not declared. |
| 6.5 Depth control | PASS | get --max-depth 1 sends maxDepth=1 to /blocks. |
| 7.1 Consistent command structure | FAIL | Document verbs are top-level, other resources grouped. Compatibility inconsistency remains. |
| 7.2 Consistent flag names | PASS | Global format, profile, timeout, dry-run and yes names shared. |
| 7.3 No surprises in output shape | PASS | Deterministic fixture responses preserve keys; no volatile response metadata added. |
| 7.4 Documented exit code table | PASS | AGENTS.md documents exit codes 0/1/2/3. |
| 7.5 Canonical verbs match common agent expectations | PASS | Common operations expose list/get/create/update/delete; convenience aliases remain. |
| 7.6 Canonical flags are enforced across commands | PASS | Canonical cross-cutting flags are global Cobra flags. |
| 7.7 Banned aliases are checked mechanically | FAIL | No banned vocabulary test or CI gate found. |
| 8.1 AGENTS.md exists | PASS | Root AGENTS.md present. |
| 8.2 AGENTS.md includes guardrails | PASS | AGENTS.md explicitly requires previews before destructive actions. |
| 8.3 Workflow examples exist | PASS | SKILL.md shows multi-step profile and revert workflows. |
| 8.4 Common mistakes documented | PASS | AGENTS.md documents timeout ambiguity and replacement pitfalls. |
| 8.5 Prompts or skills shipped | PASS | SKILL.md and prompts/ shipped. |
| 8.6 SKILL.md is installable and scoped | FAIL | Skill has valid frontmatter and 108 lines but omits remote-content trust boundary. |
| 8.7 Skill and docs are validated against the live CLI | FAIL | No executable documentation CI; stale views routing guidance. |
| 9.1 Timeout handling | PASS | HTTP client uses bounded timeout; timeout error includes verification advice. internal/api/client.go, cmd/root.go. |
| 9.2 Partial failure reporting | FAIL | DeleteDocuments/DeleteBlocks discard response bytes; MCP batch is improved but REST still loses per-item results. |
| 9.3 Retry guidance | FAIL | Fixture 429 drops Retry-After:12 and scope:space; only generic hint. |
| 9.4 Graceful degradation | PASS | Fixture 401 includes authentication failure and reconfiguration guidance. |
| 9.5 Async commands support --wait | FAIL | No async operations implemented, but no explicit exemption declaration. |
| 9.6 Durable job ledger exists | FAIL | No durable jobs or declared no-async exemption. |
| 9.7 Async retry policy is documented | FAIL | No documented async exemption/policy. |
| 10.1 Single binary or simple install | PASS | go build produces executable without runtime dependencies. |
| 10.2 Self-update | PASS | upgrade uses ChecksumValidator; background notifier never installs. cmd/upgrade.go. Source verified, not executed to avoid installation. |
| 10.3 Version or contract mismatch is detected | FAIL | Contract refreshed before baseline, but no scheduled diff. |
| 10.4 Agent-friendly auth setup | FAIL | Profile setup/status exist but credential precedence is incomplete and first-run may prompt. |
| 10.5 Offline/local mode is explicit when available | FAIL | --data-source local still reaches remote mock; local app commands also exist. |
| 11.1 Human help exists | PASS | Root and group help expose progressive discovery. |
| 11.2 Agent-context is versioned | PASS | schema_version is 2026-07-08. |
| 11.3 Agent-context exposes command metadata | FAIL | Schema omits output/exit/argument contract and marks required flags false. |
| 11.4 Request/response schemas are available | FAIL | Earlier edits enable upstream request schema for some REST commands, but command coverage incomplete and output schema is not actual CLI envelope. |
| 11.5 Skill path is discoverable | FAIL | skill-path outside repo returns nonexistent /private/tmp/SKILL.md. |
| 12.1 Profiles are supported | PASS | Named typed REST/MCP profiles implemented; config unit tests pass. |
| 12.2 Profile precedence is documented | FAIL | Precedence is not fully documented. |
| 12.3 Profiles are exposed through agent-context | FAIL | Manifest omits safe profile metadata. |
| 12.4 Config source can be inspected | FAIL | Config listing does not label sources consistently; URLs may contain link credentials. |
| 12.5 Secrets are separated from non-secret config | FAIL | API key stored with normal config; config save mode 0644. internal/config/config.go. |
| 13.1 Artifact delivery sinks exist | FAIL | --deliver parsed but unused. |
| 13.2 Delivery writes are atomic | FAIL | No delivery implementation or atomicity guarantee. |
| 13.3 Unknown delivery schemes enumerate supported values | FAIL | Invalid --deliver scheme silently succeeds. |
| 13.4 Feedback can be recorded locally | PASS | feedback writes structured local JSONL with 0600; source verified without modifying user feedback. cmd/agent_support.go. |
| 13.5 Optional upstream feedback is discoverable | FAIL | No upstream feedback declaration or explicit exemption. |
| 14.1 One source of truth exists | FAIL | Cobra, docs, manual wrappers and snapshots drift independently. |
| 14.2 Contract validation runs in CI | FAIL | Search sends folderIDs/documentIDs against documented folderIds/documentIds. No scheduled contract CI. |
| 14.3 Generated files are clearly marked | FAIL | Existing generated command reference lacks enforced generation boundary. |
| 14.4 Local/remote scope is part of the contract | FAIL | Scope not declared per command or repeated in outputs. |
| 14.5 Tool/MCP descriptions are token-budgeted | FAIL | No checked wrapper-description budget; upstream tools are passed through. |
| 15.1 Human-readable mode remains available | PASS | list --format table works. |
| 15.2 JSON is concise, not a dump | FAIL | Default list unbounded and get returns entire document. |
| 15.3 Dangerous work requires explicit commitment | FAIL | Non-TTY delete commits without --yes (mock only). |
| 15.4 Flag aliases improve usability without hiding canonical names | FAIL | Canonical names exist but schema does not identify aliases comprehensively. |
| 15.5 Skill guidance favors composition over special agent-only behavior | PASS | Skill teaches ordinary commands and shell composition. |
| 16.1 Resource-based command structure maps to the API | PASS | Resource groups and raw JSON follow Craft resource model. |
| 16.2 Data and error output formats are independently configurable | PASS | --format and --json-errors independently select streams, though error envelope needs work. |
| 16.3 Multiple structured output formats exist when useful | FAIL | list jsonl/yaml/raw all fail after fetching. |
| 16.4 Output transforms are built in | FAIL | list --transform items.0.id silently returns original payload. |
| 16.5 File arguments are first-class and explicit | FAIL | --json-file exists on selected commands; embedded explicit text/base64 file expansion absent. |
| 17.1 Local data layer exists for high-gravity resources | FAIL | No persistence/search index or incremental sync for large document sets. |
| 17.2 Data source is explicit and controllable | FAIL | --data-source local ignored by list. |
| 17.3 Compound domain commands go beyond endpoint mirroring | PASS | Compound update section/chunk workflows and connection diagnostics go beyond simple endpoint mapping. |
| 17.4 Proof-of-behavior checks exist | FAIL | Tests cover useful paths but fail on unfinished repeat change; presence-based self-audit is not proof. |
| 17.5 Provenance and competitor coverage are recorded | FAIL | No unified provenance and competitor/workflow coverage manifest. |

## Top five fixes

1. Declare effects and commitment rules per command, block non-TTY prompts, and honor dry-run on local as well as remote writes.
2. Correct live API encoding and preserve response unions, per-item outcomes and rate-limit metadata.
3. Implement actual format, transform, delivery and output-selection behavior; reject unsupported source modes before requests.
4. Publish an offline command contract, validate documentation and fixture behavior in CI, and schedule upstream contract diffs.
5. Repair skill discovery and trust guidance; declare real capability limits without inventing async or local-index support.

## Evidence notes

- `go build -o /tmp/craft-cli-v3-baseline .` succeeded. Binary probes used a loopback HTTP fixture server, fake API key, and no production mutations. All 111 schema-discovered help paths returned success; only 37 contained examples.
- Raw local evidence: `/tmp/craft-v3-baseline-probes.json`, runner `/tmp/craft-v3-probe.py`, scoring data `/tmp/craft-v3-baseline-score.json`. These temporary artifacts are supporting evidence, not required runtime files.
- `CRAFT_LIVE_TESTS=0 CRAFT_LIVE_MUTATION_TESTS=0 go test ./...` passed cmd/config/mcp and failed api/models repeat tests. These failures are part of the baseline and must be resolved before completion.
- Mock MCP `isError:true` exits 2; generic write preview performs no request, and missing `--yes` refuses with exit 1. These were repaired before this baseline.
- Source review supplements runtime probes for setup, config, upgrade, and local feedback, avoiding changes to user configuration or installed binaries. No `craft audit agent-dx` result was counted as behavioral evidence.
- Live contract spot-check: search uses `folderIds` and `documentIds`, but the wrapper sends `folderIDs`; task response nests `taskInfo`, and document lists omit `total`. Source: [Craft Connect docs](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1), captured in `docs/contracts/craft-rest-space-openapi.json`. The initial model edits correct missing totals but the source-fetch cost remains undocumented.
