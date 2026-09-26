# craft-cli Audit v3 final

Date: 2026-09-26. Audit source: clean local `cli-best-practices` at `24b9960`, left unchanged. Target: working tree based on craft-cli v1.12.0 (`ac51d02`). No version bump, commit, push or release.

**Result: 45/97 before, 83/97 after, an improvement of 38 checks.** The after total includes seven declared exemptions. Fourteen checks remain failed.

**CLI Spec status: v0.2 schema validation passed; full behavioral conformance is not claimed.** `clispec` 0.3.0 scored the safe `version` probe **22/24**. Its remaining failures require v0.3 and cardinality declarations; the emitted document deliberately targets frozen v0.2. Neither that tool score nor `craft audit agent-dx` substitutes for this audit.

## Scope and scoring

The supplied v3 checklist has **97 numbered checks**, despite its 85 label. All 97 are scored once, without normalization or silently dropping twelve checks. Categories 1-10 contain 62 checks, so its stated legacy-50 comparison is also inconsistent.

The baseline was taken before continuing implementation after the user's v3 addition, but after earlier authorized edits had started. It is not a pristine released-v1.12.0 score. In particular, MCP tool-error/commit handling and initial decoding/count fixes were already present. See the [baseline report](2026-09-26-agent-cli-audit-v3-baseline.md).

The original [audit and plan](2026-09-26-api-and-best-practices-audit.md) contains the API gap table and an Audit v3 mapping for every failed check and all 11 upstream scorecard findings. The requested OMP Opus 5.5 review was completed earlier with the explicitly requested high setting; see the [review](2026-09-26-opus-review.md). No further agents were launched after the low-reasoning constraint.

## Before and after by category

| Category | Before | After | Max |
|---|---:|---:|---:|
| Discoverability | 5 | 6 | 7 |
| Structured output | 5 | 6 | 6 |
| Input flexibility | 4 | 4 | 5 |
| Safety rails | 1 | 6 | 6 |
| Error handling | 5 | 7 | 7 |
| Context window discipline | 2 | 5 | 5 |
| Predictability | 5 | 6 | 7 |
| Agent knowledge | 5 | 7 | 7 |
| Resilience | 2 | 7 | 7 |
| Distribution and lifecycle | 2 | 5 | 5 |
| Three-layer introspection | 2 | 3 | 5 |
| Persistent identity and configuration | 1 | 2 | 5 |
| Two-way I/O and artifacts | 1 | 5 | 5 |
| Contract and generation discipline | 0 | 3 | 5 |
| Unix composability and agent restraint | 2 | 4 | 5 |
| API-native payload ergonomics | 2 | 4 | 5 |
| Domain depth and proof gates | 1 | 3 | 5 |
| **Total** | **45** | **83** | **97** |

## Exemptions

Declared exemptions count as passes under v3. No exemption is inferred solely to raise the score.

| Check | Declared reason | Evidence |
|---|---|---|
| 6.2 | REST collection pagination unavailable upstream | SKILL.md Capability limits and list --help explicitly distinguish local truncation from MCP server cursors. Reminders have server paging. |
| 9.5 | No async submission API exposed | SKILL.md Capability limits; schema extensions.async. |
| 9.6 | No async jobs to persist | SKILL.md Capability limits; schema extensions.async. |
| 9.7 | No async polling or resume workflow | SKILL.md Capability limits; synchronous timeout verification remains documented. |
| 10.5 | No local simulation/data query backend | SKILL.md explicitly distinguishes local Craft URL-scheme actions from a data cache. Unsupported data-source modes fail before requests. |
| 13.5 | No upstream feedback transport | SKILL.md and schema extensions.feedback declare local-only feedback. |
| 14.5 | Consumes upstream MCP tools; does not publish a tool server | SKILL.md and docs/capabilities/provenance.md declare that Craft owns tool descriptions; no generated MCP server surface is claimed. |

## Evidence per check

| Check | Before | After | Evidence / remaining gap |
|---|---|---|---|
| 1.1 Root help lists subcommands | PASS | PASS | Built binary root help lists subcommands and descriptions. |
| 1.2 Every subcommand has --help | PASS | PASS | Every executable manifest path was exercised with --help by scripts/verify_agent_contract.py; all succeed. |
| 1.3 Help includes examples | FAIL | FAIL | Concrete usage examples remain below the required 80% across the full command tree. Do not count generated help probes as workflow examples. |
| 1.4 Examples use realistic values | PASS | PASS | Existing examples use IDs, dates, names and actual payload syntax; new reminders/learn examples use concrete values. |
| 1.5 Progressive disclosure works | PASS | PASS | Root/group/leaf help remains available; nested schema filtering works through --command or positional command paths. |
| 1.6 Machine-readable help available | FAIL | PASS | cmd/clispec.go emits typed args and declared effects. Malformed isolated config does not affect schema; frozen v0.2 schema validation passes. |
| 1.7 Version is queryable | PASS | PASS | version returns a JSON version record; table mode keeps human text. |
| 2.1 JSON output available | PASS | PASS | Loopback list --format json emits valid JSON. |
| 2.2 JSON is consistent across commands | PASS | PASS | Document, folder and task list probes emit items envelopes with totals. Single get is a record; profile lists now use items. |
| 2.3 JSON is the default when piped | PASS | PASS | Default piped data is JSON; human output strips terminal controls in cmd/execution.go. |
| 2.4 Errors are structured | FAIL | PASS | Errors have error.kind/message plus legacy code/message aliases; validation exits 1, API exits 2, config exits 3. |
| 2.5 Exit codes are meaningful | PASS | PASS | Mock validation/config/API failures produce 1/3/2 respectively. Specific API kinds share code 2 as documented. |
| 2.6 Quiet mode suppresses noise | PASS | PASS | Quiet suppresses diagnostics; JSON write outcomes survive quiet mode. Use explicit id-only for IDs. |
| 3.1 All inputs via flags (non-interactive) | FAIL | FAIL | The optional setup wizard itself still requires a TTY and interactive answers. Equivalent profile setup is non-interactive, but this check says every command. |
| 3.2 Stdin accepted for structured input | PASS | PASS | Raw JSON/stdin mutation paths and previews remain available; fixture tests use them without real writes. |
| 3.3 Secrets without argv | PASS | PASS | Isolated profile setup reads a fake key through --api-key-env. Global CRAFT_API_KEY and saved profiles avoid secret argv. |
| 3.4 Flags over positional args | PASS | PASS | Primary identifiers are positional; optional settings use flags. Intentional multi-ID batch extensions remain. |
| 3.5 Raw payload passthrough | PASS | PASS | Raw payload paths preserved, including full folders create, view objects, reminders and task null values. |
| 4.1 Dry-run exists | PASS | PASS | Delete preview returns an intent and only performs fixture reads. |
| 4.2 Dry-run on ALL mutating commands | FAIL | PASS | REST BeforeWrite prevents network writes in dry-run; local mutations are intercepted in commandPreflight; MCP call/batch honor preview. Task/folder creation gaps fixed. |
| 4.3 Dry-run output describes the action | FAIL | PASS | dryRunOutput labels validated=local and includes operation/targets. Delete preview declares reversibility; it is not server permission validation. |
| 4.4 Confirmation skip flag | FAIL | PASS | Non-TTY deletes refuse without --yes; setup uses term.IsTerminal and refuses before banner/app launch. Generic MCP writes require --yes. |
| 4.5 Re-running is safe or declared unsafe | FAIL | PASS | Writes are conservatively declared non_idempotent with read/search verification guidance after timeout. No server idempotency or blind mutation retry is promised. |
| 4.6 Safety metadata exposed | FAIL | PASS | Complete command paths have explicit declarations in cmd/effects.go; unknown executable commands fail closed. TestCommandEffectCoverage enforces coverage. |
| 5.1 Errors are actionable | PASS | PASS | Missing argument/flag errors include requirements and a help hint. |
| 5.2 Fail fast on missing input | PASS | PASS | Missing args and first-run missing configuration return promptly; isolated-config subprocess probe confirms no banner or read loop. |
| 5.3 Network errors are distinct | PASS | PASS | Network/API failures map to exit 2; validation is exit 1. Request timeout is configurable. |
| 5.4 Error includes hint for recovery | PASS | PASS | Auth errors suggest profiles and environment references. Mock 401 produces a recovery hint. |
| 5.5 Errors to stderr, data to stdout | PASS | PASS | Errors remain on stderr; known partial write outcomes may intentionally remain on stdout with nonzero exit. |
| 5.6 Enum errors enumerate valid values | FAIL | PASS | Format/backend/source validation enumerates accepted values. Task state and reminder status flags report accepted enums. |
| 5.7 Validation happens before side effects | FAIL | PASS | Invalid format/delivery/source and missing commitment produce zero fixture requests; Cobra rejects unknown flags before RunE. |
| 6.1 Field selection available | PASS | PASS | list --fields id projects items; shared output-only and dotted transforms work. |
| 6.2 Pagination or limits | FAIL | EXEMPT | SKILL.md Capability limits and list --help explicitly distinguish local truncation from MCP server cursors. Reminders have server paging. |
| 6.3 ID-only mode | FAIL | PASS | list --id-only returns exactly fixture-a and fixture-b on separate lines, not full records. |
| 6.4 Count without fetching | FAIL | PASS | Counts derive from absent upstream total correctly. list help discloses full fetch; search marks counts as returned results. |
| 6.5 Depth control | PASS | PASS | get --max-depth 1 now reaches the REST query even in ordinary JSON mode. |
| 7.1 Consistent command structure | FAIL | FAIL | Document verbs remain top-level while most resources use resource + verb. Avoid a breaking rename solely to gain a point. |
| 7.2 Consistent flag names | PASS | PASS | Cross-cutting flags use shared names, including the new --timeout. |
| 7.3 No surprises in output shape | PASS | PASS | Deterministic fixture responses preserve output structure. Raw mode explicitly returns the final upstream response. |
| 7.4 Documented exit code table | PASS | PASS | AGENTS.md and README document 0/1/2/3; schema maps each kind to its exit code. |
| 7.5 Canonical verbs match common agent expectations | PASS | PASS | List/get/create/update/delete vocabulary remains available; reminders follow it. |
| 7.6 Canonical flags are enforced across commands | PASS | PASS | Shared Cobra flags enforce format, dry-run, yes, timeout and profile conventions. |
| 7.7 Banned aliases are checked mechanically | FAIL | PASS | TestCanonicalVocabulary rejects ls/rm as canonical verbs and checks shared flag names. Compatibility aliases remain allowed. |
| 8.1 AGENTS.md exists | PASS | PASS | Root AGENTS.md updated with current API and audit guidance. |
| 8.2 AGENTS.md includes guardrails | PASS | PASS | Guardrails require previews, scoped commitment, uncertainty verification and untrusted-content treatment. |
| 8.3 Workflow examples exist | PASS | PASS | README and SKILL provide multi-step profile and reversible-write workflows. |
| 8.4 Common mistakes documented | PASS | PASS | Replacement/style loss, timeout ambiguity, pagination and task-revert limitations are documented. |
| 8.5 Prompts or skills shipped | PASS | PASS | SKILL.md and prompts remain shipped; misleading perfect-score prompts were corrected. |
| 8.6 SKILL.md is installable and scoped | FAIL | PASS | Valid frontmatter, scoped trigger, setup, safe defaults and explicit untrusted-content rule; under 500 lines. |
| 8.7 Skill and docs are validated against the live CLI | FAIL | PASS | CI compares generated command-reference.json to live schema, checks root/bundled skill parity and runs command help plus representative examples. Prose coverage is not exhaustive. |
| 9.1 Timeout handling | PASS | PASS | --timeout sets REST and MCP HTTP deadlines. Timeout errors explain uncertain write outcomes. |
| 9.2 Partial failure reporting | FAIL | PASS | CLI captures upstream write responses before typed helpers discard them. Fixture mixed success/error retains both items and exits 2; multi-request writes expose operations. |
| 9.3 Retry guidance | FAIL | PASS | Mock Retry-After=12 and scope=space reach JSON details. REST/MCP rate-limit headers are retained; no automatic write retry. |
| 9.4 Graceful degradation | PASS | PASS | Mock auth failures return explicit auth kind and profile/key guidance. |
| 9.5 Async commands support --wait | FAIL | EXEMPT | SKILL.md Capability limits; schema extensions.async. |
| 9.6 Durable job ledger exists | FAIL | EXEMPT | SKILL.md Capability limits; schema extensions.async. |
| 9.7 Async retry policy is documented | FAIL | EXEMPT | SKILL.md Capability limits; synchronous timeout verification remains documented. |
| 10.1 Single binary or simple install | PASS | PASS | Single Go binary builds successfully. Python/jsonschema are development verification tools only. |
| 10.2 Self-update | PASS | PASS | Existing updater uses checksum validation and requires an explicit upgrade command. Automatic update notifications were removed from normal command execution. |
| 10.3 Version or contract mismatch is detected | FAIL | PASS | Current 44-operation contract is pinned. Scheduled GitHub workflow checks live drift; the actual live check passed. |
| 10.4 Agent-friendly auth setup | FAIL | PASS | Non-interactive profile setup, read-only inspection and credential precedence documented. Isolated profile behavior and 0600 storage checked. |
| 10.5 Offline/local mode is explicit when available | FAIL | EXEMPT | SKILL.md explicitly distinguishes local Craft URL-scheme actions from a data cache. Unsupported data-source modes fail before requests. |
| 11.1 Human help exists | PASS | PASS | Human root/group/leaf help remains available. |
| 11.2 Agent-context is versioned | PASS | PASS | Manifest declares clispec=0.2 and CLI version. Legacy schema remains available explicitly. |
| 11.3 Agent-context exposes command metadata | FAIL | FAIL | Types/effects/flags/errors are exposed, but per-command stdout schemas remain permissive and conditional input requirements are incomplete. |
| 11.4 Request/response schemas are available | FAIL | FAIL | Pinned REST schemas work offline for mapped operations. They are upstream schemas, not complete CLI convenience-output schemas; MCP and multi-operation commands do not all provide request/response contracts. |
| 11.5 Skill path is discoverable | FAIL | PASS | skill-path emits embedded skill metadata/content from an unrelated temporary directory; no guessed nonexistent path. |
| 12.1 Profiles are supported | PASS | PASS | Isolated named profile creation/show/list works without interactive auth. |
| 12.2 Profile precedence is documented | FAIL | PASS | README documents flag/env/profile precedence, separate MCP selection and explicit URL credential behavior. |
| 12.3 Profiles are exposed through agent-context | FAIL | FAIL | Offline schema intentionally does not load available user profile names. It points to profiles list instead, which does not meet the literal check. |
| 12.4 Config source can be inspected | FAIL | FAIL | Profile/config lists label config_file and redact values, but there is no unified effective-setting/source report for defaults, environment and per-command flags. |
| 12.5 Secrets are separated from non-secret config | FAIL | FAIL | Config is mode 0600 and reporting redacts keys/link identifiers, but credentials still share the profile JSON rather than a separate credential store. |
| 13.1 Artifact delivery sinks exist | FAIL | PASS | Structured and human output can use stdout or file delivery; loopback export needs no prose scraping. |
| 13.2 Delivery writes are atomic | FAIL | PASS | File publication uses temp file + sync + atomic link/rename; existing destination requires --yes before network calls. |
| 13.3 Unknown delivery schemes enumerate supported values | FAIL | PASS | Unknown delivery schemes enumerate stdout and file:<path>, with zero requests. |
| 13.4 Feedback can be recorded locally | PASS | PASS | Local structured feedback remains available; dry-run prevents log append. |
| 13.5 Optional upstream feedback is discoverable | FAIL | EXEMPT | SKILL.md and schema extensions.feedback declare local-only feedback. |
| 14.1 One source of truth exists | FAIL | FAIL | Cobra/effects drive the manifest and pin/fixtures validate selected wrappers. The entire CLI, all prose and MCP translators do not yet derive from one complete contract. |
| 14.2 Contract validation runs in CI | FAIL | PASS | CI runs request/response fixtures, live-schema documentation parity and a scheduled upstream diff. TestPinnedContractMatchesEncodingFixtures binds exact names/move fields to the pin. |
| 14.3 Generated files are clearly marked | FAIL | PASS | Manifest includes generated_from metadata; docs/contracts/README.md states capture/regeneration boundaries. Bundled skill copy is checked against source. |
| 14.4 Local/remote scope is part of the contract | FAIL | FAIL | Effects declare local/remote scope, but all normal responses do not repeat scope consistently, especially raw upstream data. |
| 14.5 Tool/MCP descriptions are token-budgeted | FAIL | EXEMPT | SKILL.md and docs/capabilities/provenance.md declare that Craft owns tool descriptions; no generated MCP server surface is claimed. |
| 15.1 Human-readable mode remains available | PASS | PASS | Table/markdown/rich modes remain; human output is terminal-control safe. |
| 15.2 JSON is concise, not a dump | FAIL | FAIL | Default REST lists still fetch and display all records, and get remains potentially large. Explicit limits/depth help but do not satisfy bounded defaults. |
| 15.3 Dangerous work requires explicit commitment | FAIL | PASS | Destructive mutations require explicit yes; generic MCP unknown/write tools require commitment. Dry-run blocks REST writes at the transport boundary. |
| 15.4 Flag aliases improve usability without hiding canonical names | FAIL | PASS | Canonical long flags and short aliases are exposed in schema. README distinguishes --raw markdown from --format raw REST bytes. |
| 15.5 Skill guidance favors composition over special agent-only behavior | PASS | PASS | Skill teaches normal shell commands, profiles, previews and composition, not an incompatible agent-only interface. |
| 16.1 Resource-based command structure maps to the API | PASS | PASS | Resource groups, REST views/reminders and exact payload escape hatches map to the current API. |
| 16.2 Data and error output formats are independently configurable | PASS | PASS | Success format is independent of --json-errors; JSON errors are also the default in machine-oriented modes. |
| 16.3 Multiple structured output formats exist when useful | FAIL | PASS | JSON, YAML and buffered JSONL work in fixture probes. Raw returns final REST bytes; markdown/table remain available. |
| 16.4 Output transforms are built in | FAIL | PASS | list --transform items.0.id returns the selected string; dotted object/array lookup is implemented. |
| 16.5 File arguments are first-class and explicit | FAIL | FAIL | Existing file/stdin inputs work, but general embedded @file://, @data:// and literal-@ escaping across payloads remain unimplemented. |
| 17.1 Local data layer exists for high-gravity resources | FAIL | FAIL | No SQLite/FTS cache, incremental sync or local index for high-volume Craft resources. This is not exempted. |
| 17.2 Data source is explicit and controllable | FAIL | FAIL | Unsupported local/auto modes now refuse, but there is no selectable local/live/refresh pipeline or universal response provenance. |
| 17.3 Compound domain commands go beyond endpoint mirroring | PASS | PASS | Section replacement, batching, chunked document operations and connection diagnostics provide compound workflows. |
| 17.4 Proof-of-behavior checks exist | FAIL | PASS | 141 subprocess checks, real fixture servers, pinned-contract assertions, schema validation, unit tests and vet prove behavior. Legacy presence score is not used. |
| 17.5 Provenance and competitor coverage are recorded | FAIL | PASS | docs/capabilities/provenance.md records sources and a workflow comparison against Craft MCP, including missing local search and unverified task reverts. |

## Verification performed

| Verification | Result | Boundary |
|---|---|---|
| Build | Passed, `/tmp/craft-cli-v3-final` | Local binary, version unchanged |
| `go test ./...` | Passed | `CRAFT_LIVE_TESTS=0`, `CRAFT_LIVE_MUTATION_TESTS=0` |
| `go vet ./...` | Passed | Offline static analysis |
| `scripts/verify_agent_contract.py` | 141 subprocess checks passed | Loopback REST/MCP fixtures and isolated config only |
| Pinned request-contract test | Passed | Exact search casing/arrays, block move body and 44-operation inventory |
| `scripts/craft_contract.py --check` | Passed against live docs | Read-only docs fetch; no JavaScript execution |
| Frozen CLI Spec v0.2 schema | Passed with jsonschema 4.26.0 | Pinned schema, generated command manifest |
| `cargo install clispec --locked` | Installed clispec 0.3.0 | Requested external checker |
| `clispec score /tmp/craft-cli-v3-final version --output json` | 22/24 | Safe version probe, not an API operation |
| Native Go Craft MCP discovery | Passed, three expected tools | Initialize, initialized notification and tools/list only |
| Sibling repository | Clean at 24b9960 | No edits, commits, pulls or pushes |

Raw temporary evidence includes `/tmp/craft-v3-proof.json`, `/tmp/craft-clispec-score-final.json`, `/tmp/craft-native-mcp-probe.json`, and the baseline probe files. Reproducible checks live in `scripts/` and the Go tests; temporary files are not runtime dependencies. The CI workflow was added locally but has not been pushed or run on GitHub.

No real Craft data was mutated. Live availability of reminders and actual mutation success under a particular user's permissions remain untested. Local dry-run is not server validation.

## Implemented corrections

- Correct search filter casing and repeated keys, literal block-search escaping, fetchBlocks/context controls, folder parent encoding, document task scope, block moves, recurrence payload placement, task clearing/content/location flags, upload filename/MIME and placement validation.
- Preserve page-title variants, unknown block fields, nested task information, folder children, collection schemas and search context. Avoid fabricated zero-date metadata and clarify returned search counts.
- Implement REST collection views/active view and reminders, with explicit MCP augmentation and corrected view grammar. Add block learning and direct REST mappings for verified page-style flags.
- Preserve per-item and multi-request write outcomes, distinguish partial failure, retain rate-limit metadata, enforce previews/commitment, and handle MCP application errors, sessions, notifications and separate SSE events.
- Implement structured format/projection/delivery behavior, offline CLI Spec discovery, bundled skill discovery, non-TTY refusal, private config permissions, redacted profile reporting and explicit unsupported-source errors.
- Add contract capture/drift automation, request fixtures, behavior proof, generated-reference checks, updated agent guidance and migration notes. The legacy presence diagnostic now warns that it is not behavioral evidence.

## Top remaining work, prioritized

| Priority | Checks | Concrete next task |
|---|---|---|
| P1 | 11.3, 11.4, 14.1 | Define exact CLI output schemas and conditional input requirements per command, then validate adapters, documentation and examples against them. Keep upstream REST schemas distinct from convenience output. |
| P1 | 12.4, 12.5 | Add effective configuration provenance and a separate credential store or keychain migration, preserving existing profiles safely. Current 0600 storage/redaction is an improvement, not separation. |
| P1 | 15.2, 17.1, 17.2 | Design opt-in local persistence and incremental search for large spaces. Decide bounded defaults and compatibility before changing list/get behavior. Do not pretend client truncation reduces API fetch cost. |
| P2 | 1.3, 3.1, 7.1, 12.3, 14.4 | Finish realistic examples, distinguish the optional interactive wizard from automation workflows, plan vocabulary compatibility aliases, and define runtime profile/scope metadata without making offline schema load config. |
| P2 | 16.5 | Add explicit embedded file/text/base64 expansion with literal-@ escaping and input-size limits. Existing json-file/stdin support is narrower. |
| P2 | MCP / release compatibility | Keep generic MCP access for unwrapped folder-update options; task revert/diff remains unverified. Review output/schema/confirmation changes before choosing a release version. |

A perfect score is not the objective. These gaps are recorded instead of inventing async jobs, claiming local indexing, or publishing unsupported pagination guarantees. The full checklist is not a substitute for product-specific compatibility decisions.

## All 11 prior open findings

| Prior finding | Result | Evidence |
|---|---|---|
| Search folderIDs/documentIDs | Fixed | Exact request encoding fixture and pinned-contract assertion |
| Stale REST contract/no scheduled diff | Fixed locally | 44-operation snapshot and scheduled workflow; live diff passed |
| inferSafety | Replaced | Explicit full-path effects and coverage test |
| Discarded per-item writes | Fixed at CLI boundary | Observer retains upstream outcomes; partial failure fixture exits 2 |
| Retry-After lost | Fixed | REST/MCP headers exposed; fixture delay/scope asserted |
| Safer secret documentation | Fixed | Profiles/api-key-env lead README/SKILL; isolated profile probe |
| Setup/first-run prompts | Fixed | Real TTY detection and closed-stdin refusal |
| MCP isError success | Fixed | Typed tool failure, exit 2 and preserved details |
| Generic MCP dry-run/yes | Fixed | No requests on preview or uncommitted write |
| Views documented as MCP-only | Fixed | REST defaults and corrected capability/LLM guidance |
| Count assumes total | Fixed | Missing totals derived; fetching cost and search count scope documented |

## Proposed replacement scorecard text for cli-best-practices

The sibling repository was not edited. Suggested status section:

> craft-cli v1.12.0 working tree, audited 2026-09-26 against local Audit v3 at 24b9960. Baseline: 45/97. After implementation: 83/97, including seven declared exemptions, with fourteen remaining failures. The checklist currently labels itself 85 checks but enumerates 97. This is not a released-version score and the baseline already included a few initial implementation edits. All eleven source-verified open findings are addressed in the working tree. Go tests/vet and 141 offline subprocess checks passed. CLI Spec v0.2 schema validation passed; clispec 0.3.0 scored a safe version probe 22/24 because it requires v0.3/cardinality metadata. Remaining gaps include full command/output contracts, examples, effective config provenance and credential separation, bounded defaults/local indexing, and general embedded file expansion. No production mutation tests were run.

### Concrete corrections/additions for the updated best-practices repo

| Priority | Proposed change | Why |
|---|---|---|
| P0 | Fix the v3 denominator and legacy comparison in SKILL.md, checklist, bands and scorecards: the enumerated totals are 97 and 62, not 85 and 50. | Scores cannot be compared honestly until the counted checks and labels agree. |
| P1 | Validate every CLI Spec example against its declared version. Document that current clispec 0.3.0 scoring rejects a frozen v0.2 document even if it validates against v0.2. | Separate schema validity, spec-version support and observed behavior. See the [frozen v0.2 spec](https://clispec.dev/spec/v0.2/) and [official schemas](https://github.com/rvben/clispec/tree/main/docs/schema). |
| P1 | Require denominator/check-ID inventories and executable evidence manifests alongside scores. | Prevent self-audit inflation and detect missing checks mechanically. |
| P1 | Distinguish immutable offline discovery from runtime profile/config inspection. | Requiring both no-config schema and current profile names in that same schema creates conflicting goals. A separate runtime context command can satisfy the latter. |
| P1 | Separate upstream schemas from CLI output schemas and allow documented convenience projections. | Preserving raw/unknown fields and truthful outcomes matters; forbidding all flattening is too broad. |
| P1 | Add transport fixture cases for MCP tool isError, negotiated sessions, notifications, SSE event separation, resource metadata-only behavior and partial batches. | Tool annotations are hints; correct tool error handling and request safety remain client obligations. See [MCP tools](https://modelcontextprotocol.io/specification/2026-07-28/server/tools). |
| P2 | Clarify that client limits/counts do not imply server pagination or reduced fetch cost, and that buffered JSONL is not streaming. | These were real silent behavior gaps in craft-cli. |
| P2 | Treat local indexing, arbitrary embedded file expansion and an optional interactive setup wizard as product-specific requirements, with explicit applicability rules. | Avoid forcing unrelated architecture merely to reach a universal score. |
| P2 | Prefer a native CLI for verified deterministic operations and MCP for capability gaps/review workflows; benchmark actual workloads before asserting speed/token superiority. | Transport choice should follow the current provider surface, not a universal CLI-versus-MCP rule. |

The updated repository already covers effects, retry headers, non-TTY refusal, schemas, skills and upstream drift. Those are not proposed as new omissions. The additions above target inconsistencies or implementation lessons observed in this audit.
