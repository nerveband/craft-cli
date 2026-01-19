# Craft CLI - Installation Guide

## Quick Install

### macOS (ARM64)
```bash
cd /usr/local/bin
curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Darwin_arm64.tar.gz | tar xz
chmod +x craft
```

### macOS (Intel)
```bash
cd /usr/local/bin
curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Darwin_x86_64.tar.gz | tar xz
chmod +x craft
```

### Linux (x64)
```bash
cd /usr/local/bin
sudo curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Linux_x86_64.tar.gz | tar xz
sudo chmod +x craft
```

### Linux (ARM64)
```bash
cd /usr/local/bin
sudo curl -L https://github.com/ashrafali/craft-cli/releases/latest/download/craft-cli_Linux_arm64.tar.gz | tar xz
sudo chmod +x craft
```

### Windows (x64)
1. Download `craft-cli_Windows_x86_64.zip` from [releases](https://github.com/ashrafali/craft-cli/releases/latest)
2. Extract `craft.exe` to a directory in your PATH
3. Verify installation: `craft version`

## Build from Source

### Prerequisites
- Go 1.21 or later
- Git

### Steps
```bash
# Clone the repository
git clone https://github.com/ashrafali/craft-cli.git
cd craft-cli

# Build
go build -o craft .

# Install (optional)
sudo mv craft /usr/local/bin/

# Verify
craft version
```

## Initial Setup

After installation, configure your Craft API URL:

```bash
craft config set-api https://connect.craft.do/links/YOUR_LINK/api/v1
```

To get your API URL:
1. Open Craft
2. Go to Space Settings
3. Navigate to API section
4. Copy the REST API URL

## Verify Installation

```bash
# Check version
craft version

# View configuration
craft config get-api

# List documents
craft list --format table

# Show API info
craft info
```

## Uninstall

### macOS/Linux
```bash
sudo rm /usr/local/bin/craft
rm -rf ~/.craft-cli
```

### Windows
1. Delete `craft.exe` from your PATH
2. Delete `%USERPROFILE%\.craft-cli` folder

## Troubleshooting

### "craft: command not found"
- Ensure `/usr/local/bin` is in your PATH
- Try running with full path: `/usr/local/bin/craft`

### "No API URL configured"
- Run: `craft config set-api <your-api-url>`
- Verify: `craft config get-api`

### Permission denied
- macOS/Linux: Run with `sudo` or ensure file is executable
- Windows: Run as Administrator or check file permissions

### API errors
- Verify API URL is correct
- Check network connectivity
- Ensure you have access to the Craft space

## Configuration File

Configuration is stored at `~/.craft-cli/config.json`:

```json
{
  "api_url": "https://connect.craft.do/links/xxx/api/v1",
  "default_format": "json"
}
```

You can edit this file directly or use `craft config` commands.

## Support

For issues and questions:
- GitHub Issues: https://github.com/ashrafali/craft-cli/issues
- Documentation: See README.md

## Next Steps

After installation:
1. Configure your API URL
2. Run `craft list` to test connectivity
3. Explore commands with `craft --help`
4. Read the full documentation in README.md
