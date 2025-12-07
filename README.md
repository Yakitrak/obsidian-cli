# obsidian-cli

[![CI](https://github.com/dcolthorp/obsidian-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dcolthorp/obsidian-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/dcolthorp/obsidian-cli)](https://github.com/dcolthorp/obsidian-cli/releases)
[![Go](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/github/license/dcolthorp/obsidian-cli)](LICENSE)

A fast, feature-rich command-line interface for [Obsidian](https://obsidian.md) vaults. Search, query, and manage your notes without leaving the terminal—or expose your vault to AI assistants via MCP.

## Features

- **Query & Filter** — Find notes by tag, filename pattern, folder, or frontmatter property with boolean logic
- **Graph Analysis** — Explore link structures, communities, orphans, and HITS-based importance (hub/authority scores)
- **Bulk Operations** — Add, rename, or delete tags and properties across matching notes
- **Move & Rename** — Relocate files with automatic backlink updates (git-aware)
- **MCP Server** — Expose vault operations to Claude, Cursor, VS Code, and other AI assistants
- **LLM-Ready Output** — `prompt` command formats notes for clipboard or pipe to AI tools

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap dcolthorp/homebrew-tap
brew install obsidian-cli
```

### Go Install

```bash
go install github.com/dcolthorp/obsidian-cli@latest
```

### Manual

Download the latest binary for your platform from [Releases](https://github.com/dcolthorp/obsidian-cli/releases) and add it to your `PATH`.

---

## Quick Start

```bash
# Set your default vault (required once)
obsidian-cli vault set-default "MyVault"

# List notes by tag
obsidian-cli list tag:project

# Search notes interactively
obsidian-cli search

# Open or create today's daily note
obsidian-cli daily

# Print note contents
obsidian-cli note print "Ideas/Roadmap.md"
```

---

## Commands

### Vault Management

| Command                    | Description                              |
| -------------------------- | ---------------------------------------- |
| `vault set-default <name>` | Set the default vault for all commands   |
| `vault print-default`      | Show current default vault               |
| `vault list`               | List all known vaults                    |
| `vault add <name> <path>`  | Register a vault not tracked by Obsidian |
| `vault remove <name>`      | Remove a vault from CLI preferences      |
| `vault install-ignore`     | Write default `.obsidianignore` to vault |

### Notes & Files

| Command              | Description                                                |
| -------------------- | ---------------------------------------------------------- |
| `note open <note>`   | Open a note in Obsidian                                    |
| `note print <note>`  | Print note contents to stdout                              |
| `note create <note>` | Create a new note (`--content`, `--overwrite`, `--append`) |
| `note delete <note>` | Delete a note                                              |
| `note info <note>`   | Show frontmatter and tags for a note                       |
| `daily`              | Open/create today's daily note                             |
| `search`             | Fuzzy-search notes in terminal                             |

### File Operations

| Command                                  | Description                                               |
| ---------------------------------------- | --------------------------------------------------------- |
| `file move <src> <dst>`                  | Move or rename a file (updates backlinks by default)      |
| `file move --to-folder <dir> <files...>` | Bulk move files to a folder                               |
| `file rename <src> <dst>`                | Rename with backlink rewrite and git history preservation |

**Flags:** `--overwrite`, `--update-backlinks` (default true), `--no-backlinks`, `--open`

### Query & List

```bash
# Folder listing
obsidian-cli list Notes/

# Filename pattern (glob)
obsidian-cli list find:*meeting*

# Tag filter (hierarchical)
obsidian-cli list tag:project

# Property filter (frontmatter or inline Key:: Value)
obsidian-cli list Status:active

# Boolean logic
obsidian-cli list "tag:project AND NOT tag:archived"
obsidian-cli list "( find:*2024* OR tag:yearly ) AND Office:Remote"

# Follow wikilinks
obsidian-cli list tag:project -d 2           # 2 levels deep
obsidian-cli list find:roadmap --backlinks   # include backlinks
```

**Flags:** `-d/--depth`, `--skip-anchors`, `--skip-embeds`, `--backlinks`, `-a/--absolute`

### Prompt (LLM Output)

Format notes for AI consumption. Copies to clipboard when run in a terminal.

```bash
obsidian-cli prompt tag:project
obsidian-cli prompt find:meeting -d 1 --backlinks
obsidian-cli prompt Notes/ --suppress-tags private,draft
```

Files tagged `no-prompt` are excluded by default. Use `--no-suppress` to include them.

### Tags

```bash
# List all tags with counts
obsidian-cli tags list
obsidian-cli tags list --match tag:project --json

# Add tags to matching notes
obsidian-cli tags add urgent review --inputs tag:project

# Delete tags
obsidian-cli tags delete old-tag --dry-run

# Rename tags (hierarchical)
obsidian-cli tags rename status/wip --to status/in-progress
```

### Properties

```bash
# Inspect properties across the vault
obsidian-cli properties list
obsidian-cli properties list --only status --value-counts --json

# Set property on matching notes
obsidian-cli properties set status --value '"active"' --inputs tag:project

# Delete properties
obsidian-cli properties delete draft --inputs tag:archive

# Rename/merge properties
obsidian-cli properties rename Status --to status --merge
```

**Flags:** `--source` (all|frontmatter|inline), `--match`, `--only`, `--verbose`, `--dry-run`

### Graph Analysis

```bash
# Per-note link degrees
obsidian-cli graph degrees

# Mutual-link clusters (SCCs)
obsidian-cli graph clusters

# Communities (label propagation)
obsidian-cli graph communities
obsidian-cli graph community c1234abcd --tags --neighbors

# Orphaned notes
obsidian-cli graph orphans

# Exclude patterns from all graph commands
obsidian-cli graph ignore tag:periodic find:*Archive*
```

**Flags:** `--limit`, `--all`, `--exclude`, `--include`, `--min-degree`, `--mutual-only`

---

## MCP Server

Run `obsidian-cli` as a [Model Context Protocol](https://modelcontextprotocol.io) server to expose your vault to AI assistants.

```bash
obsidian-cli mcp --vault "MyVault"
obsidian-cli mcp --vault "MyVault" --debug
```

### Configuration

**Claude Desktop** — Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%/Claude/claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obsidian-cli",
      "args": ["mcp", "--vault", "MyVault"]
    }
  }
}
```

**Cursor / VS Code** — Configure via your MCP extension settings with the same command and args.

### Available Tools

| Tool               | Description                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `files`            | List/fetch files with optional content, frontmatter, and backlinks |
| `list_tags`        | List tags with individual and aggregate counts                     |
| `list_properties`  | Inspect property usage across the vault                            |
| `community_list`   | List graph communities with anchors, sizes, and top notes          |
| `community_detail` | Full detail for a specific community                               |
| `vault_context`    | Compact overview of vault communities and key notes                |
| `note_context`     | Graph + community context for one or more notes                    |
| `daily_note`       | Get today's daily note path and content                            |
| `daily_note_path`  | Get daily note path for a given date                               |
| `rename_note`      | Rename a note with backlink updates                                |
| `move_notes`       | Move one or more notes                                             |

**Write Tools** (requires `--read-write` flag):

| Tool                | Description                                           |
| ------------------- | ----------------------------------------------------- |
| `mutate_tags`       | Add, delete, or rename tags (supports `dryRun`)       |
| `mutate_properties` | Set, delete, or rename properties (supports `dryRun`) |

### Input Patterns

All query tools accept the same pattern syntax:

- `find:*.md` — filename glob
- `tag:project` — hierarchical tag match
- `Status:active` — property filter
- `folder/path` — literal path
- Boolean: `["tag:meeting", "AND", "NOT", "tag:archived"]`

### Example Queries

```json
// Find project notes with their backlinks
{"inputs": ["tag:project"], "includeBacklinks": true, "includeContent": false}

// Get notes by author (wikilink property)
{"inputs": ["Author:[[Jane Smith]]"]}

// Preview tag rename
{"op": "rename", "fromTags": ["old"], "toTag": "new", "dryRun": true}
```

---

## YAML Formatting Note

Tag and property mutations re-serialize frontmatter using Go's `yaml.v3`. This may reorder keys, normalize indentation to 2 spaces, and strip comments. Note content is unaffected.

---

## Development

```bash
# Build
go build ./...
make build-all              # cross-compile to bin/

# Test
go test ./...
make test_coverage          # generates coverage.out

# Format
go fmt ./...

# Run locally
go run . --help
go run . list tag:project --vault "TestVault"
```

---

## Acknowledgments

This project is a fork of [Yakitrak/obsidian-cli](https://github.com/Yakitrak/obsidian-cli) by Kartikay Jainwal. Thank you to the original author and contributors.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Submit a pull request

See [open issues](https://github.com/dcolthorp/obsidian-cli/issues) for ideas or to report bugs.

---

## License

[MIT](LICENSE)
