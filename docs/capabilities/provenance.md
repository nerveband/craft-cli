# Contract provenance and workflow coverage

This CLI is a handwritten Go client. It consumes Craft REST and MCP contracts; it does not generate a competing MCP server. No latency benchmark was run. REST avoids the additional MCP command-dispatch layer, while MCP exposes review and discovery features.

| Source | Use | Verification |
|---|---|---|
| [Craft Connect reference](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1) | REST method, parameter and response definitions | Pinned 2026-10-19, live drift check passed 2026-10-21; scheduled drift check and request fixtures |
| User-supplied Craft MCP server | Alternative interface and augmentation | Discovery/help captured; native Go discovery verified without writes |
| Local cli-best-practices `24b9960` | Audit v3 methodology | All 97 enumerated checks scored before and after; no sibling edits |
| [CLI Spec](https://clispec.dev/spec/v0.2/) | Offline command schema | Frozen v0.2 JSON schema validated; external current-version score recorded separately |
| [clig.dev](https://clig.dev/) | Stream separation, scripting and human usage | JSON, stderr errors, explicit formats and non-TTY checks |
| [Agent Skills](https://agentskills.io/specification) | Bundled skill format and progressive guidance | Root and embedded skill compared in CI |

## Alternative-interface coverage

| Workflow | Direct REST CLI | Craft MCP augmentation | Remaining limitation |
|---|---|---|---|
| Read/write documents and blocks | Native commands and raw payloads | Generic read/write tools | CLI and upstream contracts are separate; MCP dynamic results remain server-defined |
| Collection views | Native REST list/create/update/delete/active-view | Verified command and flag translation | Not every nested REST view field has an MCP translator |
| Reminders | REST pagination, Save for later, completion and clearing | Generic MCP commands available | Connection availability may differ |
| Page styling | Raw styling plus theme/color/cover flags | Discovery, backdrop/washi helpers, review/revert | Task create/update live checks returned no revert payload |
| Large document discovery | REST fetch then bounded display/count | Server cursor pagination | Opt-in private index, modification-based sync and term search; REST still fetches the complete document listing |
| Private document exports | Structured/raw reads and atomic file delivery | Resource/tool reads | Caller must keep exported private content out of source control |

No broader competitor parity claim is made. The compared alternative is Craft's own MCP interface and its captured help surface.

Current implementation and verification are recorded in [the completion plan](../plans/2026-10-21-completion.md). The former beta snapshot referred to the same all-documents test link, not a separately versioned API. The three public connection references were reviewed through the web reader after direct HTML requests returned HTTP 403.
