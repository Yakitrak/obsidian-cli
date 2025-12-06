# Obsidian CLI

---

## ![obsidian-cli Usage](./docs/usage.png)

## Description

Obsidian is a powerful and extensible knowledge base application
that works on top of your local folder of plain text notes. This CLI tool (written in Go) will let you interact with the application using the terminal. You are currently able to open, search, move, create, update and delete notes.

---

## Install

### Homebrew (macOS/Linux)

```bash
brew tap dcolthorp/homebrew-tap
brew install obsidian-cli
```

### Using Go

If you have Go installed, you can install directly:

```bash
go install github.com/atomicobject/obsidian-cli@latest
```

### Manual Installation

 Download the latest release for your platform from the [releases page](https://github.com/atomicobject/obsidian-cli/releases) and add the binary to your PATH.

### Releasing (maintainers)

1. Add tokens to `.env` (loaded by direnv): `GITHUB_TOKEN` (main repo release) and `BREW_GITHUB_TOKEN` (tap write). You can also keep the aliases `OBSIDIAN_CLI_RELEASE_GITHUB_TOKEN` and `OBSIDIAN_CLI_BREW_GITHUB_TOKEN`/`DCOLTHORP_BREW_GITHUB_TOKEN` if you prefer. Then run `direnv allow`.
2. Verify tokens have `Contents: Read/Write` on `atomicobject/obsidian-cli` and `dcolthorp/homebrew-tap` (org SSO authorized if required).
3. Dry-run (no publish): `make release-dry`
4. Publish: `make release`

## Usage

### Help

```bash
# See All command instructions
obsidian-cli --help
```

### Set Default Vault

Defines default vault for future usage. If not set, pass `--vault` flag for other commands. You don't provide the path to vault here, just the name.

```bash
obsidian-cli vault set-default "{vault-name}"
```

Note: `open` and other commands in `obsidian-cli` use this vault's base directory as the working directory, not the current working directory of your terminal.

### Add Vault Path (outside Obsidian)

Store a vault name and path directly in `obsidian-cli` preferences. This is useful when the vault is not registered in Obsidian's `obsidian.json`. Use `--force` to overwrite an existing mapping.

```bash
obsidian-cli vault add "{vault-name}" "/path/to/vault" [--force]
```

### Remove Vault Path

Remove a vault mapping from `obsidian-cli` preferences.

```bash
obsidian-cli vault remove "{vault-name}"
```

### List Vault Paths

Show vault mappings stored in `obsidian-cli` preferences and Obsidian's own `obsidian.json` (includes the default marker when set).

```bash
obsidian-cli vault list
```

### Print Default Vault

Prints default vault and path. Please set this with `set-default` command if not set.

```bash
obsidian-cli vault print-default
```

### Open Note

Open given note name in Obsidian. Note can also be an absolute path from top level of vault.

```bash
# Opens note in obsidian vault
obsidian-cli note open "{note-name}"

# Opens note in specified obsidian vault
obsidian-cli note open "{note-name}" --vault "{vault-name}"

```

### Daily Note

Open daily note in Obsidian. It will create one (using template) if one does not exist.

```bash
# Creates / opens daily note in obsidian vault
obsidian-cli daily

# Creates / opens daily note in specified obsidian vault
obsidian-cli daily --vault "{vault-name}"

```

### Search Note

Starts a fuzzy search displaying notes in the terminal from the vault. You can hit enter on a note to open that in Obsidian

```bash
# Searches in default obsidian vault
obsidian-cli search

# Searches in specified obsidian vault
obsidian-cli search --vault "{vault-name}"

```

### Print Note

Prints the contents of given note name in Obsidian.

```bash
# Prints note in default vault
obsidian-cli note print "{note-name}"

# Prints note in specified obsidian
obsidian-cli note print "{note-name}" --vault "{vault-name}"

```

### Create / Update Note

Creates note (can also be a path with name) in vault via `note create`. By default, if the note exists, it will create another note but passing `--overwrite` or `--append` can be used to edit the named note.

```bash
# Creates empty note in default obsidian and opens it
obsidian-cli note create "{note-name}"

# Creates empty note in given obsidian and opens it
obsidian-cli note create "{note-name}"  --vault "{vault-name}"

# Creates note in default obsidian with content
obsidian-cli note create "{note-name}" --content "abcde"

# Creates note in default obsidian with content - overwrite existing note
obsidian-cli note create "{note-name}" --content "abcde" --overwrite

# Creates note in default obsidian with content - append existing note
obsidian-cli note create "{note-name}" --content "abcde" --append

# Creates note and opens it
obsidian-cli note create "{note-name}" --content "abcde" --open

```

### Move / Rename Note

Moves or renames notes within the vault. By default, Obsidian links match note names, so backlinks are **not** rewritten unless you opt in with `--update-backlinks`. Supports bulk moves to a folder to avoid multiple vault scans.

```bash
# Rename or move a single note
obsidian-cli note move "{current-note-path}" "{new-note-path}" [--vault "{vault-name}"] [--overwrite] [--update-backlinks] [--open]

# Bulk move notes into a folder (preserves filenames)
obsidian-cli note move --to-folder "Archive/2024" "NoteA.md" "NoteB.md" [--vault "{vault-name}"] [--overwrite] [--update-backlinks]
```

Flags:

- `--to-folder`: move one or more sources into the given folder (filenames are preserved)
- `--overwrite`: allow replacing an existing target
- `--update-backlinks`: rewrite backlinks to the new path(s); defaults to false for moves
- `--open`: open the first moved note in Obsidian after moving

### Rename Note (backlink-safe, git-aware)

Renames a note and rewrites backlinks (aliases, headers, block refs, embeds) to the new path. Tries `git mv` when the vault is a git repo (to keep history); if git cannot complete the move (e.g., due to conflicts), falls back to a filesystem rename.

```bash
obsidian-cli note rename "{source-note}" "{target-note}" [--vault "{vault-name}"] [--overwrite] [--no-backlinks]
```

Flags:

- `--overwrite`: allow replacing an existing target
- `--no-backlinks`: skip backlink rewrites (defaults to rewriting)

### Delete Note

Deletes a given note (path from top level of vault).

```bash
# Deletes a note in default obsidian
obsidian-cli note delete "{note-path}"

# Deletes a note in given obsidian
obsidian-cli note delete "{note-path}" --vault "{vault-name}"
```

### List Files

List files in your vault with various filtering options. Files can be filtered by folder, filename patterns, or tags.

```bash
# List files in the Notes folder
obsidian-cli list Notes

# Find files with "project" in the filename
obsidian-cli list find:project

# Find notes tagged with "career-pathing"
obsidian-cli list tag:career-pathing

# Find notes tagged with "career-pathing" and follow wikilinks 2 levels deep
obsidian-cli list tag:"career-pathing" -d 2

# Find notes (by filename) and follow links 1 level deep, skipping anchored links
obsidian-cli list find:project -d 1 --skip-anchors

# Find notes (by filename) and follow links 1 level deep, skipping embedded links
obsidian-cli list find:notes -d 1 --skip-embeds

# Include first-degree backlinks for matches (aliases/heading/block/embed supported)
obsidian-cli list tag:research --backlinks

# Filter by any property (frontmatter or inline Key:: Value)
obsidian-cli list Office:AOGR

# Combine filters with boolean logic (inputs default to OR)
obsidian-cli list tag:project AND tag:person
obsidian-cli list tag:project AND NOT Office:AOGR
obsidian-cli list "( tag:project AND NOT tag:archived ) OR find:*proposal*"
obsidian-cli list tag:project OR tag:research
obsidian-cli list "( find:*notes* OR find:*journal* ) AND NOT tag:private"
```

Patterns support `find:` (filename glob, matches file names only—not content), `tag:` (frontmatter + inline hashtags), `key:value` (frontmatter + inline `Key:: Value`, including Dataview), boolean `AND/OR/NOT`, and parentheses. Terms without operators are ORed.

### Properties (frontmatter + inline)

Inspect how properties are used across the vault. Scans YAML frontmatter and Dataview-style inline fields (`Key:: Value`) by default, infers shapes/types, and enumerates small value sets.

```bash
# List all properties with inferred type/shape and enums when small
obsidian-cli properties

# Focus on matching files (same patterns as list: find:, tag:, paths)
obsidian-cli properties --match "tag:career-pathing" --match "find:*Strategy*"

# Exclude tags from the report
obsidian-cli properties --exclude-tags

# Only report on specific properties (auto-raises value limit to show more values)
obsidian-cli properties --only who --only status

# Frontmatter-only scan
obsidian-cli properties --source frontmatter

# Include per-value counts in value lists
obsidian-cli properties --value-counts

# Show enums even for mixed types and raise value limit to 50
obsidian-cli properties --verbose

# JSON output
obsidian-cli properties --json
```

Key flags:

- `--match/-m`: restrict analysis to files matched by find/tag/path patterns (supports AND/OR/NOT with parentheses)
- `--exclude-tags`: omit the `tags` property
- `--source`: choose `all` (default), `frontmatter`, or `inline` (Dataview `Key:: Value`)
- `--only`: limit results to the named property/properties; when set, the default `value-limit` is raised to `max-values-1`
- `--value-limit`: max distinct values to emit inline (default 5, or `max-values-1` when `--only` is present)
- `--max-values`: cap distinct values tracked per property (except tags); automatically raised to `value-limit+1` if set lower
- `--verbose`: allow enums for mixed types and bump value limit to 50
- `--value-counts`: include per-value note counts in output

### Manage Properties (frontmatter)

Add, rename/merge, or delete frontmatter properties across matching files.

```bash
# Set/overwrite a property on matching files (value accepts YAML)
obsidian-cli properties set status --value '"in-progress"' --inputs tag:project find:*Q4*
obsidian-cli properties set effort --value 3 --inputs Notes/Project.md --overwrite

# Delete properties (optionally scoped to inputs)
obsidian-cli properties delete status owner --inputs tag:archive

# Rename and merge values when destination already exists (default --merge=true)
obsidian-cli properties rename status --to state
obsidian-cli properties rename status State --to state --merge=false  # keep existing dest value
```

### Link Graph (wikilinks)

Summarize connectedness; all commands are under `graph`:

```bash
# Per-note in/out degrees
obsidian-cli graph degrees --vault "{vault-name}"

# Mutual-link clusters (SCCs)
obsidian-cli graph clusters --vault "{vault-name}"

# Communities (looser clusters via label propagation)
obsidian-cli graph communities --vault "{vault-name}"

# Notes with no inbound or outbound links (self-links ignored)
obsidian-cli graph orphans --vault "{vault-name}"

# Skip anchors or embeds when parsing links
obsidian-cli graph clusters --skip-anchors --skip-embeds

# Exclude notes by pattern (same DSL as list/prompt) and persist defaults
obsidian-cli graph communities --exclude "tag:periodic" --exclude "find:*Archive*"
obsidian-cli graph ignore tag:periodic find:*Archive*

# Show more detail
obsidian-cli graph degrees --limit 10        # top 10 lists (pagerank/in/out)
obsidian-cli graph communities --all         # list every community/member
obsidian-cli graph community c1234abcd --tags --neighbors --all  # inspect a specific community (by id)
obsidian-cli graph community "Projects/Note.md" --tags           # inspect the community containing a specific file
```

Graph ignores are stored per-vault in `.obsidian-cli/config.json` (set via `graph ignore`). Patterns use the same AND/OR/NOT DSL as `list`/`prompt`.

### Tags (listing)

List tags with per-note and hierarchical counts. You can scope listing to a subset of files using the same match patterns as `list`.

```bash
obsidian-cli tags list
obsidian-cli tags list --match "tag:project" --match "find:*2025*"
obsidian-cli tags list --json
```

`--match` accepts the same patterns and boolean logic as `list` and only applies to `tags`/`tags list`.

### File Information

Show detailed information about a file including its frontmatter and all tags.

```bash
# Show file info in default vault
obsidian-cli note info "Notes/Project.md"

# Show file info in specified vault
obsidian-cli note info "Notes/Project.md" --vault "{vault-name}"
```

### Prompt (LLM Format)

List files with contents formatted for LLM consumption. Similar to the list command but outputs file contents in a format optimized for LLMs. When run in a terminal, the output is automatically copied to clipboard.

**Tag Suppression**: By default, files tagged with `no-prompt` are excluded from output. This behavior can be controlled with `--suppress-tags` and `--no-suppress` flags.

```bash
# Format files in the Notes folder for LLM consumption
obsidian-cli prompt Notes

# Find and format files with "joe" in the filename
obsidian-cli prompt find:joe

# Find and format notes tagged with "career-pathing"
obsidian-cli prompt tag:career-pathing

# Find and format notes with tag, following links 2 levels deep
obsidian-cli prompt tag:"career-pathing" -d 2

# Find and format notes with "project", following links but skipping anchors
obsidian-cli prompt find:project -f --skip-anchors

# Tag suppression examples
obsidian-cli prompt tag:work --suppress-tags private,draft    # Exclude files with private or draft tags
obsidian-cli prompt Notes --no-suppress                       # Don't exclude any tags (including no-prompt)
obsidian-cli prompt find:project --suppress-tags no-prompt,private  # Custom suppression list

# Include first-degree backlinks for matches (aliases/heading/block/embed supported)
obsidian-cli prompt find:project --backlinks
```

`prompt` accepts the same patterns and boolean logic as `list` (find:/tag:/paths and `key:value` property filters).

### Manage Tags

List, add, delete, or rename tags across your vault. By default, `tags` lists all tags with individual and aggregate counts.

**Note:** Delete and rename operations only affect tags with non-zero "Indiv" counts (exact tag matches). Tags with zero individual count are parent tags in hierarchies (e.g., `context` when only `context/work` exists) and won't be modified.

```bash
# List all tags with counts
obsidian-cli tags list

# List tags as JSON
obsidian-cli tags list --json

# List tags as markdown table
obsidian-cli tags list --markdown

# Add tags to matching notes (combine multiple inputs for OR logic)
obsidian-cli tags add work urgent --inputs tag:project find:meeting

# Delete specific tags from all notes
obsidian-cli tags delete work urgent

# Rename tags across all notes
obsidian-cli tags rename old --to new

# Preview changes without making them
obsidian-cli tags delete work urgent --dry-run

# Control parallelism with custom worker count
obsidian-cli tags rename foo bar --to baz --workers 4
```

### Default ignores

obsidian-cli skips common tooling clutter even if you haven't created a `.obsidianignore` (e.g., `.git/`, `.cursor/`, `.codex/`, `.cache/`, `node_modules/`, `dist/`, `build/`, `.venv/`). To write the default ignore file into your vault for reuse, run:

```bash
obsidian-cli vault install-ignore --vault "MyVault"   # add --force to overwrite an existing file
```

## YAML formatting changes when editing tags or properties

When tag or property editing operations touch the YAML front-matter, the block is re-emitted using Go’s `yaml.v3` encoder. As a result:

- Keys may appear in a different order than before.
- Nested items (like the `tags:` list) are indented with two spaces.
- Comments or blank lines inside the front-matter are stripped.
- Line endings are normalised to LF (`\n`).

This is cosmetic—note content is unaffected—but if you rely on exact front-matter formatting be aware the tool will produce canonical YAML instead of preserving original whitespace/comments.

## MCP Server

Obsidian CLI can run as a Model Context Protocol (MCP) server, exposing its functionality as tools that can be used by MCP clients like Claude Desktop, Cursor, or VS Code.

### What is MCP?

The Model Context Protocol (MCP) is a protocol that allows AI assistants to interact with external tools and data sources in a standardized way. When obsidian-cli runs as an MCP server, AI assistants can use its commands as tools to interact with your Obsidian vault.

### Starting the MCP Server

```bash
# Start MCP server for default vault
obsidian-cli mcp

# Start MCP server for specific vault
obsidian-cli mcp --vault "MyVault"

# Start with debug output
obsidian-cli mcp --debug

# Suppress specific tags from results
obsidian-cli mcp --suppress-tags "no-prompt,private"

# Allow all tags (disable default suppression)
obsidian-cli mcp --no-suppress
```

### Available MCP Tools

The MCP server exposes the following tools:

Core (read-only):

- **`files`**: List files matching criteria and optionally include content/frontmatter. Returns JSON with `{vault,count,files:[{path,absolutePath?,tags,frontmatter?,content?}]}`. Use `match` patterns like `find:`, `tag:`, `key:value`, or paths; supports AND/OR/NOT with parentheses. `includeFrontmatter` true returns raw frontmatter/properties.
- **`list_tags`**: JSON list of tags with individual/aggregate counts. Supports `match` filter (same patterns/boolean logic as `files`).
- **`list_properties`**: Inspect properties across the vault (frontmatter + inline `Key:: Value`). Returns inferred shape/type, note counts, enums, and per-value counts. Supports `match` (same patterns/boolean logic as `files`), `excludeTags`, `only`, `source`, `valueLimit` (default 25, or `maxValues-1` when `only` is set), `maxValues`, `verbose`, and `valueCounts` (default true).
- **`move_notes`**: Move or rename one or more notes. Accepts `moves:[{source,target}]` (or single `source`/`target`). Options: `overwrite`, `updateBacklinks` (default false), `open`.
- **`daily_note`**: JSON describing the daily note `{path,date,exists,content}` (defaults to today).
- **`daily_note_path`**: JSON with `{path,date,exists}`.

Destructive (require `--read-write`):

- **`add_tags`**, **`delete_tags`**, **`rename_tag`** (all respect `dryRun`).
  - Example (delete_tags): `{"tags":["old","deprecated"],"dryRun":true}` to preview; set `dryRun:false` to apply. Destructive tools register only when the server is started with `--read-write`.

#### files inputs and options

- Inputs are **OR'd** and support:
  - `find:<pattern>`: filename glob (supports `*` and `?`), e.g. `find:2024-*.md`
  - `tag:<tag>`: hierarchical tag match (e.g., `tag:project` matches `project/work`)
  - Literal file or folder paths relative to the vault root, e.g., `Notes/Project.md` or `Daily Notes/`
- Useful flags:
  - `includeContent` (default `true`), `includeFrontmatter` to control payload size
  - `maxDepth` to traverse wikilinks; `skipAnchors` / `skipEmbeds` to filter link types
  - `suppressTags` to add more suppressed tags, `noSuppress` to disable defaults (configured via server flags)
  - `absolutePaths` to include absolute paths in each entry

Example call (arguments array):

```json
{
  "inputs": ["tag:project", "find:meeting*"],
  "maxDepth": 1,
  "includeContent": true,
  "includeFrontmatter": false
}
```

Example response (files):

```json
{
  "vault": "MyVault",
  "count": 2,
  "files": [
    {
      "path": "Notes/Project.md",
      "tags": ["project"],
      "content": "# Project\n..."
    },
    {
      "path": "Notes/meeting-notes.md",
      "tags": ["project", "meeting"],
      "content": "# Meeting\n..."
    }
  ]
}
```

`list_tags` response example:

```json
{ "tags": [{ "name": "project", "individualCount": 12, "aggregateCount": 20 }] }
```

`delete_tags` dry-run response example:

```json
{
  "dryRun": true,
  "notesTouched": 3,
  "tagChanges": { "old": 2, "deprecated": 1 },
  "filesChanged": ["Notes/a.md", "Notes/b.md", "Notes/c.md"]
}
```

### Claude Desktop Configuration

To use obsidian-cli as an MCP server with Claude Desktop, add this to your Claude configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "obsidian-cli": {
      "command": "/path/to/obsidian-cli",
      "args": ["mcp", "--vault", "YourVaultName"]
    }
  }
}
```

Replace `/path/to/obsidian-cli` with the actual path to your obsidian-cli executable and `YourVaultName` with your vault name.

### Other MCP Clients

The MCP server works with any MCP-compatible client. The server communicates over stdin/stdout using the standard MCP protocol.

### Example Usage in AI Assistants

Once configured, you can ask AI assistants to:

- "List all notes tagged with 'project'"
- "Show me the contents of my daily note"
- "Search for notes containing 'meeting notes'"
- "What tags are available in my vault?"
- "Show me information about the file 'Ideas/New Project.md'"

The AI assistant will use the appropriate MCP tools to interact with your Obsidian vault and provide the requested information.

## Acknowledgments

This project is a fork of [Yakitrak/obsidian-cli](https://github.com/Yakitrak/obsidian-cli) by Kartikay Jainwal. We thank the original author and contributors for their foundational work.

## Contribution

Fork the project, add your feature or fix and submit a pull request. You can also open an [issue](https://github.com/atomicobject/obsidian-cli/issues/new/choose) to report a bug or request a feature.

## License

Available under [MIT License](./LICENSE)
