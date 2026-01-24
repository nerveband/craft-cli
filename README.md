# Craft CLI

A powerful command-line interface for interacting with Craft Documents. Built for speed, automation, and seamless integration with LLMs and scripting workflows.

![Craft CLI Demo](demo.gif)

## Features

- **Multi-Profile Support** - Store multiple Craft API connections and switch between them
- **Multiple Output Formats** - JSON (default), Table, and Markdown outputs
- **LLM/Script Friendly** - Quiet mode, JSON errors, field extraction, stdin support
- **Local Craft Integration** - Open documents, create new docs, search directly in Craft app (macOS)
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
craft update <document-id> --title "Updated Title"

# Delete a document
craft delete <document-id>

# Preview delete without executing
craft delete <document-id> --dry-run
```

### Multi-Profile Management

Store and switch between multiple Craft API connections:

```bash
# Add profiles
craft config add work https://connect.craft.do/links/WORK_LINK/api/v1
craft config add personal https://connect.craft.do/links/PERSONAL_LINK/api/v1

# List all profiles (* = active)
craft config list

# Switch active profile
craft config use personal

# Remove a profile
craft config remove old-profile

# Reset all configuration
craft config reset

# Override profile for single command
craft list --api-url https://connect.craft.do/links/OTHER_LINK/api/v1
```

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

# Raw content output
craft get <doc-id> --raw

# No table headers
craft list --format table --no-headers

# Dry-run mode
craft create --title "Test" --dry-run

# Read content from stdin
cat document.md | craft create --title "Imported" --stdin
echo "New content" | craft update <doc-id> --stdin
```

### Output Formats

```bash
# JSON (default) - best for scripts and LLMs
craft list --format json

# Table - human readable
craft list --format table

# Markdown - documentation friendly
craft get <doc-id> --format markdown
```

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

## Configuration

Configuration is stored in `~/.craft-cli/config.json`:

```json
{
  "default_format": "json",
  "active_profile": "work",
  "profiles": {
    "work": {
      "url": "https://connect.craft.do/links/WORK_LINK/api/v1"
    },
    "personal": {
      "url": "https://connect.craft.do/links/PERSONAL_LINK/api/v1"
    }
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (invalid input, missing arguments) |
| 2 | API error (server issues, network problems) |
| 3 | Configuration error |

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
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Support

- GitHub Issues: https://github.com/nerveband/craft-cli/issues
- Craft API Docs: https://support.craft.do/hc/en-us/articles/23702897811612
