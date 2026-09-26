# Contract provenance and workflow coverage

This CLI is a handwritten Go client. It consumes Craft REST and MCP contracts; it does not generate a competing MCP server. No latency benchmark was run. REST avoids the additional MCP command-dispatch layer, while MCP exposes review and discovery features.

| Source | Use | Verification |
|---|---|---|
| [Craft Connect reference](https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1) | REST method, parameter and response definitions | Pinned 2026-09-26; scheduled drift check and request fixtures |
| User-supplied Craft MCP server | Alternative interface and augmentation | Discovery/help captured; native Go discovery verified without writes |
| Local cli-best-practices `24b9960` | Audit v3 methodology | All 97 enumerated checks scored before and after; no sibling edits |
| [CLI Spec](https://clispec.dev/spec/v0.2/) | Offline command schema | Frozen v0.2 JSON schema validated; external current-version score recorded separately |
| [clig.dev](https://clig.dev/) | Stream separation, scripting and human usage | JSON, stderr errors, explicit formats and non-TTY checks |
| [Agent Skills](https://agentskills.io/specification) | Bundled skill format and progressive guidance | Root and embedded skill compared in CI |

## Alternative-interface coverage

| Workflow | Direct REST CLI | Craft MCP augmentation | Remaining limitation |
|---|---|---|---|
| Read/write documents and blocks | Native commands and raw payloads | Generic read/write tools | Some convenience output contracts need richer schemas |
| Collection views | Native REST list/create/update/delete/active-view | Verified command and flag translation | Not every nested REST view field has an MCP translator |
| Reminders | REST pagination, Save for later, completion and clearing | Generic MCP commands available | Connection availability may differ |
| Page styling | Raw styling plus theme/color/cover flags | Discovery, backdrop/washi helpers, review/revert | Task-write revert metadata unverified |
| Large document discovery | REST fetch then bounded display/count | Server cursor pagination | No local index, incremental sync or full-text cache |
| Private document exports | Structured/raw reads and atomic file delivery | Resource/tool reads | Caller must keep exported private content out of source control |

No broader competitor parity claim is made. The compared alternative is Craft's own MCP interface and its captured help surface.
