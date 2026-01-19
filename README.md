# Craft CLI

A cross-platform command-line interface for interacting with Craft Documents via REST API. Built for speed, token efficiency, and seamless integration with LLMs and automation agents.

## Features

- Fast, single-binary Go application
- Cross-platform support (macOS, Linux, Windows)
- REST API integration with Craft Documents
- Multiple output formats (JSON, Table, Markdown)
- Persistent configuration management
- Comprehensive error handling
- Built for LLM/agent integration

## Installation

### Download Pre-compiled Binaries

Download the latest release for your platform from the [releases page](https://github.com/ashrafali/craft-cli/releases).

#### macOS (ARM64)
```bash
curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Darwin_arm64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Darwin_x86_64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

#### Linux (x64)
```bash
curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Linux_x86_64.tar.gz | tar xz
sudo mv craft /usr/local/bin/
```

#### Windows (x64)
Download `craft-cli_Windows_x86_64.zip` from releases and add to your PATH.

### Build from Source

```bash
git clone https://github.com/ashrafali/craft-cli.git
cd craft-cli
go build -o craft .
```

## Configuration

### Set API URL

First, configure your Craft API URL:

```bash
craft config set-api https://connect.craft.do/links/YOUR_LINK/api/v1
```

Configuration is stored in `~/.craft-cli/config.json`.

### View Current Configuration

```bash
craft config get-api
```

### Reset Configuration

```bash
craft config reset
```

### Override API URL Per Command

Use the `--api-url` flag to override the configured URL for any command:

```bash
craft list --api-url https://connect.craft.do/links/ANOTHER_LINK/api/v1
```

## Usage

### List Documents

```bash
# List all documents (JSON format - default)
craft list

# List as human-readable table
craft list --format table

# List as markdown
craft list --format markdown
```

### Get a Document

```bash
# Get document by ID
craft get <document-id>

# Save document content to file
craft get <document-id> --output document.md

# Get with custom format
craft get <document-id> --format markdown
```

### Search Documents

```bash
# Search for documents
craft search "query terms"

# Search with table output
craft search "query" --format table
```

### Create a Document

```bash
# Create with title only
craft create --title "My New Document"

# Create from file
craft create --title "My Document" --file content.md

# Create with inline markdown
craft create --title "Quick Note" --markdown "# Hello\nThis is content"

# Create as child of another document
craft create --title "Child Doc" --parent <parent-id>
```

### Update a Document

```bash
# Update title
craft update <document-id> --title "New Title"

# Update from file
craft update <document-id> --file updated-content.md

# Update with inline markdown
craft update <document-id> --markdown "# Updated\nNew content"

# Update both title and content
craft update <document-id> --title "New Title" --file content.md
```

### Delete a Document

```bash
craft delete <document-id>
```

### Show API Information

```bash
# Show API info and recent documents
craft info

# List all available documents
craft docs
```

### Version

```bash
craft version
```

## Output Formats

The CLI supports three output formats:

- **json** (default): Machine-readable JSON output, ideal for LLMs and scripts
- **table**: Human-readable table format
- **markdown**: Markdown-formatted output

Set a default format in your config or use `--format` flag per command.

## Error Handling

The CLI provides clear error messages with appropriate exit codes:

- **Exit Code 0**: Success
- **Exit Code 1**: User error (invalid input, missing arguments)
- **Exit Code 2**: API error (server-side issues)
- **Exit Code 3**: Configuration error

Common error messages:

- `authentication failed. Check API URL` - Invalid or unauthorized API URL
- `resource not found` - Document ID doesn't exist
- `rate limit exceeded. Retry later` - Too many requests
- `no API URL configured. Run 'craft config set-api <url>' first` - Configuration missing

## Testing

The project includes comprehensive tests with >80% coverage:

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Development

### Prerequisites

- Go 1.21 or later
- goreleaser (for building releases)

### Project Structure

```
craft-cli/
├── cmd/                 # CLI commands (cobra)
│   ├── root.go         # Root command
│   ├── config.go       # Configuration commands
│   ├── list.go         # List documents
│   ├── get.go          # Get single document
│   ├── search.go       # Search documents
│   ├── create.go       # Create document
│   ├── update.go       # Update document
│   ├── delete.go       # Delete document
│   ├── info.go         # Info & docs commands
│   ├── version.go      # Version command
│   └── output.go       # Output formatting
├── internal/
│   ├── api/            # API client
│   │   ├── client.go
│   │   └── client_test.go
│   ├── config/         # Config management
│   │   ├── config.go
│   │   └── config_test.go
│   └── models/         # Data structures
│       └── document.go
├── .goreleaser.yml     # Cross-platform builds
├── README.md
├── go.mod
└── main.go
```

### Building

```bash
# Build for current platform
go build -o craft .

# Build with goreleaser (all platforms)
goreleaser build --snapshot --clean

# Create a release
goreleaser release --clean
```

## API Integration

This CLI wraps the Craft Documents REST API. For API documentation, visit your Craft space's API docs:

```
https://connect.craft.do/link/YOUR_LINK/docs/v1
```

## Examples

### Workflow: Create and Update

```bash
# Create a new document
craft create --title "Project Notes" --file notes.md

# Get the document ID from output, then update
craft update <doc-id> --title "Updated Project Notes"

# Verify the update
craft get <doc-id> --format markdown
```

### Workflow: Search and Export

```bash
# Search for documents
craft search "meeting notes" --format table

# Export a specific document
craft get <doc-id> --output exported-notes.md
```

### LLM Integration

The JSON output format (default) is optimized for LLM consumption:

```bash
# Get all documents as JSON
craft list | jq '.[] | {id, title}'

# Search and pipe to another tool
craft search "query" | python process.py
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Support

For issues and questions:
- GitHub Issues: https://github.com/ashrafali/craft-cli/issues
- Craft API Docs: Check your space's API documentation

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [goreleaser](https://goreleaser.com/) - Release automation
