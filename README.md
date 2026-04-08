# bunny-cli

Command-line interface for the [bunny.net](https://bunny.net) API. Manage DNS zones, pull zones, storage, edge scripts, Shield security, Stream video, and more from your terminal.

## Requirements

- Go 1.26+

## Install

```bash
go install github.com/built-fast/bunny-cli@latest
```

Or build from source:

```bash
make build    # produces ./bin/bunny
```

## Configuration

Run the interactive setup:

```bash
bunny configure
```

This creates `~/.config/bunny/config.toml` with your API key.

You can also configure via environment variable:

```bash
export BUNNY_API_KEY=your-api-key
```

Or pass flags directly:

```bash
bunny dns list --api-key <key>
```

Precedence: flags > environment variables > config file.

## Usage

```bash
# List resources
bunny dns list --limit 50
bunny pullzones list

# Get a resource
bunny dns get <zone_id>
bunny pullzones get <pullzone_id>

# Create
bunny dns create --domain example.com

# Update
bunny pullzones update <pullzone_id> --name my-zone

# Destructive operations require confirmation (or --yes)
bunny dns delete <zone_id>
bunny dns delete <zone_id> --yes

# File operations
bunny storage ls my-zone/path/
bunny storage cp local-file.txt my-zone/remote/path/
bunny storage rm my-zone/remote/path/file.txt
```

### Resources

| Resource | Commands |
|---|---|
| `dns` | list, get, create, update, delete, export, import |
| `dns records` | list, get, create, update, delete |
| `dns dnssec` | enable, disable, get |
| `pullzones` | list, get, create, update, delete, purge |
| `pullzones edge-rules` | list, create, update, delete |
| `pullzones hostnames` | list, add, remove, set-certificate, load-certificate, remove-certificate |
| `storagezones` | list, get, create, update, delete, reset-password |
| `storage` | ls, cp, rm |
| `scripts` | list, get, create, update, delete, publish, rotate-key |
| `scripts code` | get, update |
| `scripts releases` | list |
| `scripts secrets` | list, create, update, delete |
| `shield zones` | list, get, create, update, delete |
| `shield waf` | get, update, custom-rules, triggered-rules |
| `shield rate-limits` | list, get, create, update, delete |
| `shield access-lists` | list, get, create, update, delete |
| `shield bot-detection` | get, update |
| `shield upload-scanning` | get, update |
| `shield event-logs` | list |
| `shield metrics` | get |
| `stream libraries` | list, get, create, update, delete |
| `stream videos` | list, get, create, update, delete, upload, fetch, reencode, repackage |
| `stream collections` | list, get, create, update, delete |
| `stream captions` | list, add, delete |
| `stream statistics` | get |
| `statistics` | *(global CDN statistics)* |
| `regions` | *(list CDN regions)* |
| `countries` | *(list countries)* |
| `account` | get, update |
| `billing` | get, summary, apply-promo-code |

### Output

```bash
# Table (default), JSON, or pretty JSON
bunny dns list --output table
bunny dns list --output json
bunny dns list --output json-pretty

# Select specific fields
bunny dns list --field id,domain

# Built-in jq filtering (no external jq needed)
bunny pullzones list --jq '.[] | select(.Enabled == true) | .Name'
```

### Shell completion

```bash
bunny completion bash
bunny completion zsh
bunny completion fish
bunny completion powershell
```

## Development

Install dependencies:

```bash
brew bundle
```

Run the full check suite (formatting, linting, tests, surface snapshot):

```bash
make check
```

Individual targets:

```bash
make build          # Build binary
make test           # Unit tests
make test-e2e       # E2E tests (BATS + Prism mock server)
make lint           # golangci-lint
make fmt            # Format code
make surface        # Regenerate CLI surface snapshot
```

## License

MIT
