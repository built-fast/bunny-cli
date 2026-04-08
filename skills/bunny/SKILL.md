---
name: bunny
description: Bunny CLI — manage bunny.net CDN, storage, DNS, video streaming, edge scripts, and security from the command line. Use this skill when the user needs to interact with the bunny.net API.
triggers:
  - bunny
  - CDN
  - pull zone
  - edge storage
  - bunny dns
  - video streaming
  - edge scripting
  - bunny shield
---

# Bunny CLI Skill

The `bunny` CLI manages bunny.net resources from the terminal. This document
teaches agents how to invoke any command correctly.

## Authentication

Precedence (highest to lowest):

1. `--api-key` flag on any command
2. `BUNNY_API_KEY` environment variable
3. Config file (created via `bunny configure`)

```bash
# Interactive setup (writes ~/.config/bunny/config.toml)
bunny configure

# One-off override
bunny pullzones list --api-key <key>

# Environment variable
export BUNNY_API_KEY=<key>
bunny pullzones list
```

### Edge Storage Authentication

Edge Storage file operations (`bunny storage`) use per-zone passwords, not the
main API key. The CLI auto-detects the password and hostname from the storage
zone when possible. You can also provide them explicitly:

```bash
bunny storage ls my-zone/
bunny storage --password <zone-password> --hostname <zone-hostname> ls my-zone/
```

## Output Modes

| Flag | Effect |
|---|---|
| `--output table` | Human-readable table (default for TTY) |
| `--output json` | Compact JSON |
| `--output json-pretty` | Indented JSON |
| `--jq <expr>` | Apply jq expression to JSON output (built-in, no external jq) |
| `--field <fields>` | Comma-separated field names to display |

Agent invariant: always use `--output json` when parsing output programmatically.
The `--jq` flag implies JSON output. `--jq` and `--output table` are mutually
exclusive.

### Field Selection

```bash
bunny pullzones list --field id,name,origin-url
bunny pullzones get 123 --field id,name --output json
```

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error (invalid flags, validation failure, network error, server error, auth error, not found) |

Error messages are written to stderr. Agents should check exit code and parse
stderr for diagnostics.

## Pagination

List commands support these flags:

| Flag | Type | Description |
|---|---|---|
| `--limit` | int | Maximum number of items to return (default varies) |
| `--all` | bool | Fetch all pages automatically |

When `--output json` is used, list commands return the complete JSON array.
Use `--all` to fetch every item across all pages.

```bash
bunny pullzones list --all --output json
bunny pullzones list --limit 5 --output json
```

## Destructive Operations

Commands that delete or modify data irreversibly prompt for confirmation.
Pass `--yes` to skip the prompt (required for non-interactive/agent use):

```bash
bunny pullzones delete 123 --yes
bunny storagezones delete 456 --yes
bunny storage rm my-zone/path/file.txt --yes
```

## File-Based Input

Create and update commands accept JSON or YAML input via `--from-file` / `-F`:

```bash
bunny pullzones create --from-file pullzone.json
bunny pullzones create -F pullzone.yaml
echo '{"Name":"test"}' | bunny pullzones create --from-file -
```

CLI flags override values from the file. Use `--no-input` with `--from-file` to
disable interactive prompts.

## Watch Mode

Get commands support `--watch <interval>` to poll for changes:

```bash
bunny pullzones get 123 --watch 5s
bunny dns get 456 --watch 10s
```

## Command Aliases

| Full Name | Alias |
|---|---|
| `pullzones` | `pz` |
| `storagezones` | `sz` |
| `dns` | `dnszone` |
| `scripts` | `compute` |
| `shield zones` | `shield zone` |
| `shield access-lists` | `shield access-list` |
| `shield rate-limits` | `shield rate-limit`, `shield ratelimits` |
| `shield bot-detection` | `shield bot` |
| `shield upload-scanning` | `shield scanning` |
| `scripts variables` | `scripts vars` |
| `stream libraries` | `stream lib` |
| `statistics` | `stats` |
| `account api-keys` | `account apikeys` |
| `shield waf custom-rules` | `shield waf custom-rule` |

## Skill Management

```bash
# Print this SKILL.md to stdout
bunny skill

# Install SKILL.md to ~/.agents/skills/bunny/ with Claude Code symlink
bunny skill install

# Print the skill installation path
bunny skill path
```

The skill auto-refreshes when the CLI version changes.

## Shell Completions

```bash
# Bash — add to ~/.bashrc
source <(bunny completion bash)

# Zsh — add to ~/.zshrc (before compinit)
source <(bunny completion zsh)

# Fish
bunny completion fish | source

# PowerShell
bunny completion powershell | Out-String | Invoke-Expression
```

---

## Command Reference

### Configure

```bash
bunny configure
```

Interactive setup for API key and preferences. Writes to
`~/.config/bunny/config.toml`.

### Version

```bash
bunny version
```

Print CLI version, commit hash, and build date.

---

### Pull Zones

```bash
bunny pullzones list [--limit int] [--all] [--search string]
bunny pullzones get <id> [--watch string]
bunny pullzones create [--name string] [--origin-url string] [--type int] [--from-file string] [--no-input]
bunny pullzones update <id> [--origin-url string] [--origin-host-header string] [--add-host-header] [--verify-origin-ssl] [--from-file string]
bunny pullzones delete <id> --yes
```

#### Pull Zone Hostnames

```bash
bunny pullzones hostnames list <pull_zone_id>
bunny pullzones hostnames add <pull_zone_id> --hostname string
bunny pullzones hostnames remove <pull_zone_id> --hostname string [--yes]
```

#### Pull Zone Purge

```bash
bunny pullzones purge <pull_zone_id> [--tag string] [--yes]
```

#### Pull Zone Edge Rules

```bash
bunny pullzones edge-rules list <pull_zone_id>
bunny pullzones edge-rules add <pull_zone_id> [--action-type int] [--action-parameter1 string] [--action-parameter2 string] [--trigger-matching-type int] [--description string] [--enabled] [--from-file string]
bunny pullzones edge-rules enable <pull_zone_id> <edge_rule_id>
bunny pullzones edge-rules disable <pull_zone_id> <edge_rule_id>
bunny pullzones edge-rules delete <pull_zone_id> <edge_rule_id> --yes
```

---

### Storage Zones

```bash
bunny storagezones list [--limit int] [--all] [--search string] [--include-deleted]
bunny storagezones get <id> [--watch string]
bunny storagezones create [--name string] [--region string] [--zone-tier int] [--replication-regions stringSlice] [--from-file string] [--no-input]
bunny storagezones update <id> [--origin-url string] [--custom-404-file-path string] [--rewrite-404-to-200] [--replication-zones stringSlice] [--from-file string]
bunny storagezones delete <id> [--delete-linked-pull-zones] --yes
bunny storagezones reset-password <id> [--read-only] --yes
```

---

### Edge Storage (File Operations)

Storage file operations use per-zone passwords. The `--password` and
`--hostname` flags are persistent on the `storage` parent command.

```bash
bunny storage ls <zone>[/<path>]
bunny storage cp <src> <dst> [--checksum]
bunny storage rm <zone>/<path> --yes
```

Upload: `bunny storage cp local-file.txt my-zone/remote/path.txt`
Download: `bunny storage cp my-zone/remote/path.txt local-file.txt`

---

### DNS Zones

```bash
bunny dns list [--limit int] [--all] [--search string]
bunny dns get <id> [--watch string]
bunny dns create [--domain string] [--from-file string] [--no-input]
bunny dns update <id> [--soa-email string] [--nameserver1 string] [--nameserver2 string] [--custom-nameservers-enabled] [--logging-enabled] [--logging-ip-anonymization-enabled] [--log-anonymization-type int] [--certificate-key-type int] [--from-file string]
bunny dns delete <id> --yes
bunny dns import <zone_id> --file string
bunny dns export <zone_id> [--output-file string]
```

#### DNSSEC

```bash
bunny dns dnssec enable <zone_id>
bunny dns dnssec disable <zone_id> --yes
```

#### DNS Records

```bash
bunny dns records list <zone_id>
bunny dns records add <zone_id> [--name string] [--type string] [--value string] [--ttl int] [--priority int] [--weight int] [--port int] [--comment string] [--disabled] [--from-file string] [--no-input]
bunny dns records update <zone_id> <record_id> [--name string] [--type string] [--value string] [--ttl int] [--priority int] [--weight int] [--port int] [--comment string] [--disabled] [--from-file string]
bunny dns records delete <zone_id> <record_id> --yes
```

---

### Stream (Video)

#### Stream Libraries

```bash
bunny stream libraries list [--limit int] [--all] [--search string]
bunny stream libraries get <id> [--watch string]
bunny stream libraries create [--name string] [--replication-regions stringSlice] [--from-file string] [--no-input]
bunny stream libraries update <id> [--name string] [--enabled-resolutions string] [--allow-direct-play] [--enable-drm] [--enable-mp4-fallback] [--keep-original-files] [--webhook-url string] [--replication-regions stringSlice] [--from-file string]
bunny stream libraries delete <id> --yes
bunny stream libraries reset-api-key <id> --yes
bunny stream libraries languages
```

#### Stream Videos

```bash
bunny stream videos list <library_id> [--limit int] [--all] [--search string] [--collection string] [--order-by string]
bunny stream videos get <library_id> <video_id> [--watch string]
bunny stream videos create <library_id> [--title string] [--collection-id string] [--thumbnail-time int] [--from-file string] [--no-input]
bunny stream videos update <library_id> <video_id> [--title string] [--collection-id string] [--from-file string]
bunny stream videos delete <library_id> <video_id> --yes
bunny stream videos upload <library_id> <file> [--title string] [--collection-id string]
bunny stream videos fetch <library_id> --url string [--title string] [--collection-id string]
bunny stream videos reencode <library_id> <video_id>
bunny stream videos transcribe <library_id> <video_id> [--source-language string] [--languages stringSlice] [--generate-title] [--generate-description] [--generate-chapters] [--generate-moments]
```

#### Stream Collections

```bash
bunny stream collections list <library_id> [--limit int] [--all] [--search string] [--order-by string]
bunny stream collections get <library_id> <collection_id> [--watch string]
bunny stream collections create <library_id> [--name string] [--no-input]
bunny stream collections update <library_id> <collection_id> [--name string]
bunny stream collections delete <library_id> <collection_id> --yes
```

#### Stream Captions

```bash
bunny stream captions add <library_id> <video_id> --srclang string --label string --file string
bunny stream captions delete <library_id> <video_id> --srclang string --yes
```

#### Stream Statistics & Heatmap

```bash
bunny stream statistics <library_id> [--date-from string] [--date-to string] [--hourly] [--video-guid string]
bunny stream heatmap <library_id> <video_id>
```

---

### Edge Scripts

```bash
bunny scripts list [--limit int] [--all] [--search string] [--type string]
bunny scripts get <id> [--watch string]
bunny scripts create [--name string] [--type string] [--code string] [--from-file string] [--no-input]
bunny scripts update <id> [--name string] [--type string] [--from-file string]
bunny scripts delete <id> [--delete-linked-pullzones] --yes
bunny scripts rotate-key <id> --yes
bunny scripts statistics <id> [--date-from string] [--date-to string] [--hourly] [--load-latest]
```

#### Script Code

```bash
bunny scripts code get <id> [--output-file string]
bunny scripts code set <id> --file string
```

#### Script Releases

```bash
bunny scripts publish <script_id> [uuid] [--note string]
bunny scripts releases list <script_id> [--limit int] [--all]
bunny scripts releases active <script_id>
```

#### Script Variables

```bash
bunny scripts variables list <script_id>
bunny scripts variables get <script_id> <variable_id>
bunny scripts variables add <script_id> [--name string] [--default-value string] [--required] [--from-file string] [--no-input]
bunny scripts variables update <script_id> <variable_id> [--default-value string] [--required] [--from-file string]
bunny scripts variables delete <script_id> <variable_id> --yes
```

#### Script Secrets

```bash
bunny scripts secrets list <script_id>
bunny scripts secrets add <script_id> [--name string] [--secret string] [--from-file string] [--no-input]
bunny scripts secrets update <script_id> <secret_id> [--secret string] [--from-file string]
bunny scripts secrets delete <script_id> <secret_id> --yes
```

---

### Shield (Security)

#### Shield Zones

```bash
bunny shield zones list [--limit int] [--all]
bunny shield zones get <shield_zone_id> [--watch string]
bunny shield zones get-by-pullzone <pull_zone_id>
bunny shield zones create [--pull-zone-id int64] [--from-file string] [--no-input]
bunny shield zones update <shield_zone_id> [--waf-enabled] [--waf-execution-mode int] [--waf-profile-id int] [--waf-realtime-threat-intel] [--waf-request-header-logging] [--ddos-execution-mode int] [--ddos-sensitivity int] [--ddos-challenge-window int] [--learning-mode] [--whitelabel-response-pages] [--from-file string]
```

#### Shield WAF Rules

```bash
bunny shield waf rules list <shield_zone_id>
```

#### Shield WAF Custom Rules

```bash
bunny shield waf custom-rules list <shield_zone_id> [--limit int] [--all]
bunny shield waf custom-rules get <id> [--watch string]
bunny shield waf custom-rules create [--rule-name string] [--rule-description string] [--shield-zone-id int64] [--from-file string] [--no-input]
bunny shield waf custom-rules update <id> [--rule-name string] [--rule-description string] [--from-file string]
bunny shield waf custom-rules delete <id> --yes
```

#### Shield WAF Engine & Profiles

```bash
bunny shield waf engine
bunny shield waf profiles
```

#### Shield WAF Triggered Rules

```bash
bunny shield waf triggered list <shield_zone_id>
bunny shield waf triggered update <shield_zone_id> [--rule-id string] [--action int]
```

#### Shield Rate Limits

```bash
bunny shield rate-limits list <shield_zone_id> [--limit int] [--all]
bunny shield rate-limits get <id> [--watch string]
bunny shield rate-limits create [--rule-name string] [--rule-description string] [--shield-zone-id int64] [--from-file string] [--no-input]
bunny shield rate-limits update <id> [--rule-name string] [--rule-description string] [--from-file string]
bunny shield rate-limits delete <id> --yes
```

#### Shield Access Lists

```bash
bunny shield access-lists list <shield_zone_id>
bunny shield access-lists get <shield_zone_id> <id> [--watch string]
bunny shield access-lists create <shield_zone_id> [--name string] [--description string] [--type int] [--content string] [--from-file string] [--no-input]
bunny shield access-lists update <shield_zone_id> <id> [--name string] [--content string] [--from-file string]
bunny shield access-lists delete <shield_zone_id> <id> --yes
bunny shield access-lists config update <shield_zone_id> <config_id> [--enabled] [--action int]
```

#### Shield Bot Detection

```bash
bunny shield bot-detection get <shield_zone_id>
bunny shield bot-detection update <shield_zone_id> [--execution-mode int] [--fingerprint-aggression int] [--fingerprint-sensitivity int] [--ip-sensitivity int] [--request-integrity int] [--from-file string]
```

#### Shield Upload Scanning

```bash
bunny shield upload-scanning get <shield_zone_id>
bunny shield upload-scanning update <shield_zone_id> [--enabled] [--antivirus-scanning-mode int] [--csam-scanning-mode int] [--from-file string]
```

#### Shield Event Logs

```bash
bunny shield event-logs <shield_zone_id> <date> [--continuation-token string]
```

#### Shield Metrics

```bash
bunny shield metrics overview <shield_zone_id>
bunny shield metrics detailed <shield_zone_id> [--start-date string] [--end-date string] [--resolution int]
bunny shield metrics rate-limits <shield_zone_id>
bunny shield metrics waf-rule <shield_zone_id> <rule_id>
bunny shield metrics bot-detection <shield_zone_id>
bunny shield metrics upload-scanning <shield_zone_id>
```

---

### Account

#### API Keys

```bash
bunny account api-keys list [--limit int] [--all]
```

#### Audit Log

```bash
bunny account audit-log <date> [--limit int] [--all] [--order string] [--product stringSlice] [--resource-type stringSlice] [--resource-id stringSlice] [--actor-id stringSlice]
```

---

### Billing

```bash
bunny billing summary
bunny billing details
bunny billing records
bunny billing invoice <billing-record-id> [--output-file string]
```

---

### Statistics

```bash
bunny statistics [--date-from string] [--date-to string] [--pull-zone int64] [--server-zone-id int64] [--hourly] [--load-errors]
```

---

### Regions

```bash
bunny regions
```

Lists all available CDN regions.

---

### Countries

```bash
bunny countries
```

Lists all countries with their CDN pricing tiers.

---

## Common Agent Workflows

### Create a Pull Zone

```bash
bunny pullzones create --name my-site --origin-url https://origin.example.com --output json
```

### Upload a File to Edge Storage

```bash
bunny storage cp ./build/index.html my-zone/www/index.html
```

### Purge CDN Cache

```bash
bunny pullzones purge 123456 --yes
```

### Create a DNS Zone and Add Records

```bash
bunny dns create --domain example.com --output json
bunny dns records add <zone_id> --name www --type CNAME --value cdn.example.com --ttl 300
```

### Deploy an Edge Script

```bash
bunny scripts create --name my-worker --type 1 --output json
bunny scripts code set <script_id> --file worker.js
bunny scripts publish <script_id> --note "Initial deploy"
```

### Set Up Shield Protection

```bash
bunny shield zones create --pull-zone-id 123456 --output json
bunny shield zones update <shield_zone_id> --waf-enabled --waf-execution-mode 1
```

### Download a Billing Invoice

```bash
bunny billing records --output json
bunny billing invoice <billing-record-id> --output-file invoice.pdf
```
