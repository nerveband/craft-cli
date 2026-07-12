# Craft CLI

A powerful command-line interface for interacting with Craft Documents. Built for speed, automation, and seamless integration with LLMs and scripting workflows.

![Craft CLI Demo](demo.gif)

## Features

- **Multi-Profile Support** - Store multiple Craft API connections and switch between them
- **API Key Authentication** - Support for API keys with secure storage per profile
- **Craft MCP Support** - Inspect and call Craft MCP tools with `craft mcp`
- **Agent DX Audit** - Score CLI agent-readiness with `craft audit agent-dx`
- **Multiple Output Formats** - JSON (default, full API payloads), Compact (legacy), Table, and Markdown outputs
- **LLM/Script Friendly** - Quiet mode, JSON errors, field extraction, stdin support
- **Local Craft Integration** - Open documents, create new docs, search directly in Craft app (macOS)
- **Auto-Chunking** - Automatically splits large documents to avoid API limits
- **Section Replacement** - Update specific sections by heading name, including Craft-decorated headings
- **Self-Updating** - Built-in upgrade command to stay up to date
- **Interactive Setup** - Guided first-time configuration wizard
- **Shell Completions** - Tab completion for Bash, Zsh, Fish, and PowerShell
- **Cross-Platform** - Works on macOS, Linux, and Windows
- **Dry-Run Mode** - Preview changes before making them

## Quick Start

### Installation

#### One-Line Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/nerveband/craft-cli/main/install.sh | bash
```

#### Download Pre-compiled Binaries

Download from the [releases page](https://github.com/nerveband/craft-cli/releases):

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/nerveband/craft-cli/releases/latest/download/craft-cli_Darwin_arm64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/nerveband/craft-cli/releases/latest/download/craft-cli_Darwin_x86_64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

**Linux (x64)**
```bash
curl -L https://github.com/nerveband/craft-cli/releases/latest/download/craft-cli_Linux_x86_64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

**Windows (x64)**
Download `craft-cli_Windows_x86_64.zip` from releases and add to your PATH.

#### Build from Source

```bash
git clone https://github.com/nerveband/craft-cli.git
cd craft-cli
go build -o craft .
```

### First-Time Setup

Run the interactive setup wizard:

```bash
craft setup
```

This will guide you through:
1. Getting your API URL from the Craft app
2. Creating your first profile
3. Verifying the connection

Or configure manually:

```bash
craft config add work https://connect.craft.do/links/YOUR_LINK/api/v1
```

## Usage

### Document Operations

```bash
# List all documents
craft list

# List with table format
craft list --format table

# Get a specific document
craft get <document-id>

# Get as markdown
craft get <document-id> --format markdown

# Search documents
craft search "meeting notes"

# Create a document
craft create --title "New Document" --markdown "# Hello World"

# Create from file
craft create --title "From File" --file content.md

# Create from stdin
echo "# My Content" | craft create --title "From Stdin" --stdin

# Update a document
craft update <document-id> --title "Updated Title"           # Rename (updates root page block)
craft update <document-id> --file content.md                  # Append content
craft update <document-id> --mode replace --file content.md   # Replace all content blocks
craft update <document-id> --mode replace --section "Intro" --file intro.md

# Delete a document
craft delete <document-id>            # Move to trash (soft-delete)

# Clear document content (does not delete the document)
craft clear <document-id>

# Preview delete without executing
craft delete <document-id> --dry-run
```

`update --mode replace` is a markdown replacement path: it deletes and recreates content blocks. Use it for structural rewrites, not routine text edits on styled documents. For existing styled blocks, use `craft blocks update BLOCK_ID --markdown ...` so omitted styling fields stay attached to the block. If new blocks are created, style only those changed/new blocks rather than rerunning a full document styling pass.

Section replacement matches headings even when Craft markdown wraps them in lightweight styling tags such as `<callout>## Pricing</callout>`. Replacement content can also start with a decorated heading without duplicating the original heading. Use `--dry-run` before real section replacements when editing styled documents.

### Multi-Profile Management

Store and switch between multiple Craft API connections:

```bash
# Add typed REST and MCP profiles
craft profiles add-rest work --api-url https://connect.craft.do/links/WORK_LINK/api/v1
craft profiles add-mcp work-mcp --mcp-url https://mcp.craft.do/links/WORK_LINK/mcp
craft profiles list
craft profiles use work

# Equivalent config aliases are also available
craft config add-rest work --api-url https://connect.craft.do/links/WORK_LINK/api/v1
craft config add-mcp work-mcp --mcp-url https://mcp.craft.do/links/WORK_LINK/mcp

# Legacy REST config commands remain supported
# Add profiles
craft config add work https://connect.craft.do/links/WORK_LINK/api/v1
craft config add personal https://connect.craft.do/links/PERSONAL_LINK/api/v1

# Add profile with API key for authentication
craft config add secure https://connect.craft.do/links/LINK/api/v1 --key pdk_your_key_here

# List all profiles (* = active, [key] = has API key)
craft config list

# Switch active profile
craft config use personal

# Remove a profile
craft config remove old-profile

# Reset all configuration
craft config reset

# Override profile for single command
craft list --profile work
craft mcp tools --profile work-mcp
craft list --api-url https://connect.craft.do/links/OTHER_LINK/api/v1

# Use API key for single command (without saving to profile)
craft list --api-url https://connect.craft.do/.../api/v1 --api-key pdk_your_key
```

### Craft MCP

Some Craft capabilities are exposed through Craft MCP before they are available in the direct REST workflow, including link resolution, theme/style exploration, richer collection view controls, and reversible block edits. Configure MCP per command with `--mcp-url` or through `CRAFT_MCP_URL`.

```bash
# Save a reusable MCP profile
craft profiles add-mcp work-mcp --mcp-url https://mcp.craft.do/links/YOUR_LINK/mcp

# List available MCP tools
craft mcp tools --mcp-url https://mcp.craft.do/links/YOUR_LINK/mcp
craft mcp tools --profile work-mcp

# Call a Craft MCP read command
craft mcp call craft_read --command "connection info"

# Call with raw JSON arguments
craft mcp call craft_write --arguments '{"command":"documents create --title Test"}'
```

Use REST/API profiles for deterministic direct API calls. Use MCP when an operation needs MCP-only capabilities such as `documents resolve-link`, page themes/covers/backdrops, edit-review/revert metadata, or richer collection view controls.

MCP profiles are separate profiles, not an `mcp_url` attribute on a REST profile. The canonical config shape is `"type": "mcp"` plus `"mcp_url"` in its own profile entry, created with `craft profiles add-mcp` or `craft config add-mcp`. A top-level `mcp_url` or an `mcp_url` added to a REST profile is ignored by MCP commands.

If an agent is using REST and the requested feature is MCP-only, it should stop retrying REST, check for an existing MCP profile, ask the user for a Craft MCP URL if none exists, save it with `craft config add-mcp`, verify it with `craft profiles test`, then rerun the operation with that MCP profile. The CLI cannot generate a new Craft MCP link by itself; the user must provide or create the URL in Craft.

For MCP style writes, verify with MCP instead of REST when the fields are MCP-only:

```bash
craft blocks update PAGE_ID --theme-id soil-and-clay --backend mcp --profile work-mcp --dry-run
craft blocks update PAGE_ID --theme-id soil-and-clay --backend mcp --profile work-mcp --yes --save-revert revert.json --diff
craft blocks revert --revert-info-file revert.json --profile work-mcp --dry-run
```

`--diff` returns the mutation result plus MCP edit-review metadata when available. `--save-revert` stores the undo payload needed by `craft blocks revert`.

### Local Craft App Commands (macOS)

Interact directly with the Craft app on your Mac:

```bash
# Open a document in Craft
craft local open <document-id>

# Create a new document in Craft
craft local new

# Create with title
craft local new --title "Quick Note"

# Append to daily notes
craft local today "Remember to call John"
craft local yesterday "What I did yesterday"
craft local tomorrow "Tasks for tomorrow"

# Search in Craft
craft local search "project ideas"
```

### LLM & Scripting Features

Optimized for automation and LLM integration:

```bash
# Quiet mode - suppress status messages
craft list -q

# JSON error output for parsing
craft list --json-errors

# Extract specific fields
craft list --output-only id
craft list --id-only
craft list --fields id,title,lastModifiedAt --limit 5
craft search "api" --fields documentId,markdown --limit 5
craft list --count

# Raw content output
craft get <doc-id> --raw

# No table headers
craft list --format table --no-headers

# Dry-run mode
craft create --title "Test" --dry-run

# Agent-readiness scorecard
craft audit agent-dx --format json

# MCP inspection and batch execution
craft mcp tools
craft mcp edit-review
craft batch --command "connection info" --dry-run

# MCP-ahead collection view controls
craft collections rename <collection-id> --name "New Name" --dry-run
craft collections create --backend mcp --name Tasks --property Status=select --dry-run
craft collections views list <collection-id>
craft collections active-view set <collection-id> --view <view-id> --dry-run

# MCP page styling and reversible block mutations
craft blocks update <page-id> --theme-id fire-horse --dry-run
craft blocks add <page-id> --markdown "Test" --backend mcp --save-revert revert.json --dry-run
craft list --backend mcp --cursor <cursor> --limit 5
craft whiteboards elements get <whiteboard-id>
craft images view https://example.com/image.jpg

# Read content from stdin
cat document.md | craft create --title "Imported" --stdin
echo "New content" | craft update <doc-id> --stdin
```

### Live Test Matrix

Live tests are opt-in and skip by default:

```bash
# Read-only REST and MCP live checks
CRAFT_LIVE_TESTS=1 \
CRAFT_LIVE_REST_URL="https://connect.craft.do/links/<id>/api/v1" \
CRAFT_LIVE_REST_KEY="$CRAFT_API_KEY" \
CRAFT_LIVE_MCP_URL="https://mcp.craft.do/links/<id>/mcp" \
go test ./... -run Live

# Optional write-only state check
CRAFT_LIVE_TESTS=1 \
CRAFT_LIVE_WRITEONLY_URL="https://connect.craft.do/links/<id>/api/v1" \
go test ./internal/api -run LiveRESTWriteOnlyState

# Mutations require an explicit second gate and only create temp artifacts
CRAFT_LIVE_TESTS=1 CRAFT_LIVE_MUTATION_TESTS=1 \
CRAFT_LIVE_REST_URL="https://connect.craft.do/links/<id>/api/v1" \
go test ./internal/api -run LiveRESTMutation
```

### Output Formats

```bash
# JSON (default) - full API/MCP-shaped payloads (best for scripts and LLMs)
craft list --format json

# Compact - legacy flattened JSON output
craft list --format compact

# Table - human readable
craft list --format table

# Markdown - documentation friendly
craft get <doc-id> --format markdown
```

### LLM & Styling Docs

LLM-friendly docs live in `docs/llm/`:
- `docs/llm/README.md` - quick index
- `docs/llm/styling-and-markdown.md` - markdown shortcuts + styling JSON examples
- `docs/llm/output-parity.md` - MCP/API/CLI parity notes + differences chart

### MCP/API/CLI Differences (Summary)

| Surface | JSON shape | Notes |
| --- | --- | --- |
| MCP | JSON-RPC envelope | Blocks payload returned inside `result.content[].text` |
| API | REST payload | List endpoints return `{items, total}` |
| CLI | Default JSON | Mirrors API shapes; `--format compact` keeps legacy flattened arrays |

### Shell Completions

Enable tab completion for your shell:

**Bash**
```bash
craft completion bash > /etc/bash_completion.d/craft
# Or on macOS with Homebrew:
craft completion bash > $(brew --prefix)/etc/bash_completion.d/craft
```

**Zsh**
```bash
craft completion zsh > "${fpath[1]}/_craft"
```

**Fish**
```bash
craft completion fish > ~/.config/fish/completions/craft.fish
```

**PowerShell**
```powershell
craft completion powershell > craft.ps1
# Then source from your profile
```

### Self-Update

Keep Craft CLI up to date:

```bash
# Check for and install updates
craft upgrade

# Check current version
craft version
```

After installing an update, `craft upgrade` prints recent release notes and marks both the version just installed and the version you upgraded from.

**For LLMs and Automated Agents:**

When encountering issues, missing features, or errors that might be fixed in newer versions:
1. Always check the current version with `craft version`
2. Run `craft upgrade` to get the latest features, bug fixes, and improvements
3. The CLI will notify you when updates are available during normal operation
4. New versions may include important bug fixes, performance improvements, or new commands

**When to upgrade:**
- Before starting new tasks or workflows
- After encountering unexpected errors
- When documentation mentions features not available in your version
- Periodically to stay up to date with latest improvements

## Configuration

### Config File Location

Configuration is stored in `~/.craft-cli/config.json`

You can edit this file directly or use `craft config` commands to manage it.

### Config File Structure

```json
{
  "default_format": "json",
  "active_profile": "work",
  "profiles": {
    "work": {
      "type": "rest",
      "url": "https://connect.craft.do/links/WORK_LINK/api/v1",
      "api_key": "pdk_your_api_key_here"
    },
    "work-mcp": {
      "type": "mcp",
      "mcp_url": "https://mcp.craft.do/links/WORK_LINK/mcp",
      "access_mode": "public",
      "document_scope": "connection-defined"
    },
    "personal": {
      "type": "rest",
      "url": "https://connect.craft.do/links/PERSONAL_LINK/api/v1"
    }
  }
}
```

**Field Descriptions:**
- `default_format`: Default output format (`json`, `table`, or `markdown`)
- `active_profile`: Name of the currently active profile
- `profiles`: Map of named profiles, each containing:
  - `type`: `rest` or `mcp`; omitted legacy profiles are treated as `rest`
  - `url`: Craft REST API URL from your workspace link
  - `mcp_url`: Craft MCP URL; only read when `type` is `mcp`
  - `api_key`: (Optional) API key for authentication

Use `craft profiles add-rest` / `craft profiles add-mcp` or `craft config add-rest` / `craft config add-mcp` instead of hand-editing this file. If you do hand-edit, remember that MCP must be its own profile with `"type": "mcp"`.

### Understanding Permissions

Both **public links** and **API keys** can have different permission levels. These permissions are configured in Craft (not in this CLI):

**Permission Levels:**
- **Read-only**: Can list, get, and search documents
- **Write-only**: Can create, update, and delete documents
- **Read-write**: Full access to all operations

**How to Set Permissions:**
1. In Craft, go to your workspace settings
2. Find the share link or API key settings
3. Configure the permission level (read-only, write-only, or read-write)

**Testing Your Permissions:**
```bash
# Show current profile info and test permissions
craft info --test-permissions

# Try operations with dry-run to check permissions
craft create --title "Test" --dry-run
craft delete <doc-id> --dry-run
```

### Troubleshooting Permission Errors

If you get `PERMISSION_DENIED` errors:

1. **Check your link/key permissions in Craft**
   - Public links: Check share settings in Craft
   - API keys: Verify key permissions in workspace settings

2. **Understand the operation requirements**
   - List/Get/Search require read permission
   - Create/Update require write permission
   - Delete requires write permission

3. **Common scenarios**
   - Read-only key trying to create → Need write permission
   - Write-only key trying to list → Need read permission
   - Expired or invalid API key → Regenerate in Craft

4. **Test your setup**
   ```bash
   craft info --test-permissions
   ```

### Security Notes

- API keys are stored in **plain text** in the config file
- Ensure appropriate file permissions: `chmod 600 ~/.craft-cli/config.json`
- Never commit your config file to version control
- Regenerate API keys if accidentally exposed
- Use different profiles for different security levels

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (invalid input, missing arguments, permission denied) |
| 2 | API error (server issues, network problems, authentication failure) |
| 3 | Configuration error |

**Error Categories** (with `--json-errors`):
- `AUTH_ERROR` - Invalid or missing API key
- `PERMISSION_DENIED` - API key lacks required permissions (read-only vs read-write)
- `NOT_FOUND` - Resource not found
- `IMAGE_ASSET_UNAVAILABLE` - MCP write could not resolve a markdown image URL; validate it with `craft images view URL`, or upload local bytes with `craft upload FILE --page PAGE_ID`
- `PAYLOAD_TOO_LARGE` - Request too large (use `--chunk-bytes` to tune)
- `RATE_LIMIT` - Too many requests
- `API_ERROR` - Server-side error
- `CONFIG_ERROR` - Configuration issue

## Examples

### Workflow: Daily Notes

```bash
# Add today's accomplishments
craft local today "Completed feature X"

# Create a meeting note
craft create --title "Meeting Notes $(date +%Y-%m-%d)" --stdin << EOF
# Team Standup

## Discussed
- Project timeline
- Resource allocation

## Action Items
- [ ] Follow up with design team
EOF
```

### Workflow: Export and Backup

```bash
# Export all documents to files
for id in $(craft list --id-only -q); do
  title=$(craft get $id --output-only title -q)
  craft get $id --raw > "backup/${title}.md"
done
```

### LLM Integration

```bash
# Get document content for LLM processing
content=$(craft get <doc-id> --raw -q)

# List documents as structured data
craft list -q | jq '.[] | {id, title, updated}'

# Create document from LLM output
llm_response | craft create --title "Generated Content" --stdin
```

## Development

### Project Structure

```
craft-cli/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
│   ├── root.go             # Root command and global flags
│   ├── config.go           # Profile management (add, use, list, remove)
│   ├── profiles.go         # Typed REST/MCP profiles
│   ├── mcp.go              # Craft MCP client commands
│   ├── documents.go        # Document helpers such as resolve-link
│   ├── blocks.go           # Block CRUD, styling, exploration, revert
│   ├── collections.go      # Collections, items, schema, MCP-backed views
│   ├── folders.go          # Folder CRUD and MCP icon exploration
│   ├── tasks.go            # Task list/add/update/delete
│   ├── upload.go           # File upload and raw upload payloads
│   ├── whiteboards.go      # Whiteboard commands
│   ├── audit.go            # Agent-DX audit scorecard
│   ├── schema.go           # Machine-readable command manifest
│   ├── validate.go         # Local proof-of-behavior checks
│   ├── setup.go            # Interactive setup wizard
│   ├── list.go             # List documents
│   ├── get.go              # Get document details
│   ├── create.go           # Create documents
│   ├── update.go           # Update documents
│   ├── delete.go           # Delete documents
│   ├── search.go           # Search documents
│   ├── local.go            # macOS Craft app integration
│   ├── upgrade.go          # Self-update functionality
│   ├── version.go          # Version information
│   ├── completion.go       # Shell completions
│   ├── output.go           # Output formatting (JSON, table, markdown)
│   └── info.go             # API info command
├── internal/
│   ├── api/
│   │   ├── client.go       # Craft REST API client
│   │   └── client_test.go
│   ├── mcp/
│   │   ├── client.go       # Streamable HTTP MCP client
│   │   └── client_test.go
│   ├── config/
│   │   ├── config.go       # Configuration management
│   │   └── config_test.go
│   └── models/
│       └── document.go     # Document data structures
├── docs/
│   ├── capabilities/       # REST vs MCP capability maps
│   ├── contracts/          # Captured REST/MCP contract snapshots
│   ├── command-reference.json
│   └── llm/                # Agent-facing docs and output parity notes
├── install.sh              # One-line installer script
├── .goreleaser.yml         # Release configuration
└── README.md
```

### Prerequisites

- Go 1.21 or later
- goreleaser (for releases)

### Building

```bash
# Build for current platform
go build -o craft .

# Build all platforms
goreleaser build --snapshot --clean

# Create a release
goreleaser release --clean
```

### Testing

```bash
go test ./... -v
go test ./... -cover

# Agent readiness
craft audit agent-dx --format json

# Static analysis and release packaging
go vet ./...
goreleaser build --snapshot --clean

# Public live REST/MCP checks
CRAFT_LIVE_TESTS=1 \
CRAFT_LIVE_REST_URL=https://connect.craft.do/links/HHRuPxZZTJ6/api/v1 \
CRAFT_LIVE_MCP_URL=https://mcp.craft.do/links/wlYPoWSB9T/mcp \
CRAFT_LIVE_WRITEONLY_URL=https://connect.craft.do/links/Dfl9gELEXWY/api/v1 \
go test ./... -run Live

# Approved mutation fixture, key supplied only through environment
CRAFT_LIVE_TESTS=1 CRAFT_LIVE_MUTATION_TESTS=1 \
CRAFT_LIVE_REST_URL=https://connect.craft.do/links/5VruASgpXo0/api/v1 \
CRAFT_LIVE_REST_KEY=<redacted> \
go test ./internal/api -run LiveRESTMutation
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Support

- GitHub Issues: https://github.com/nerveband/craft-cli/issues
- Craft API Docs: https://connect.craft.do/api-docs
