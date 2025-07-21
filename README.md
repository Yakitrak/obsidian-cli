# Obsidian CLI

---

## ![obsidian-cli Usage](./docs/usage.png)

## Description

Obsidian is a powerful and extensible knowledge base application
that works on top of your local folder of plain text notes. This CLI tool (written in Go) will let you interact with the application using the terminal. You are currently able to open, search, move, create, update and delete notes.

---

## Install

### Windows

You will need to have [Scoop](https://scoop.sh/) installed. On powershell run:

```
scoop bucket add scoop-yakitrak https://github.com/yakitrak/scoop-yakitrak.git
```

```
scoop install obsidian-cli
```

### Mac and Linux

You will need to have [Homebrew](https://brew.sh/) installed.

```Bash
brew tap yakitrak/yakitrak
```

```Bash
brew install yakitrak/yakitrak/obsidian-cli
```

For full installation instructions, see [Mac and Linux manual](https://yakitrak.github.io/obsidian-cli-docs/docs/install/mac-and-linux).

## Usage

### Help

```bash
# See All command instructions
obsidian-cli --help
```

### Set Default Vault

Defines default vault for future usage. If not set, pass `--vault` flag for other commands. You don't provide the path to vault here, just the name.

```bash
obsidian-cli set-default "{vault-name}"
```

Note: `open` and other commands in `obsidian-cli` use this vault's base directory as the working directory, not the current working directory of your terminal.

### Print Default Vault

Prints default vault and path. Please set this with `set-default` command if not set.

```bash
obsidian-cli print-default
```

### Open Note

Open given note name in Obsidian. Note can also be an absolute path from top level of vault.

```bash
# Opens note in obsidian vault
obsidian-cli open "{note-name}"

# Opens note in specified obsidian vault
obsidian-cli open "{note-name}" --vault "{vault-name}"

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
obsidian-cli print "{note-name}"

# Prints note in specified obsidian
obsidian-cli print "{note-name}" --vault "{vault-name}"

```

### Create / Update Note

Creates note (can also be a path with name) in vault. By default, if the note exists, it will create another note but passing `--overwrite` or `--append` can be used to edit the named note.

```bash
# Creates empty note in default obsidian and opens it
obsidian-cli create "{note-name}"

# Creates empty note in given obsidian and opens it
obsidian-cli create "{note-name}"  --vault "{vault-name}"

# Creates note in default obsidian with content
obsidian-cli create "{note-name}" --content "abcde"

# Creates note in default obsidian with content - overwrite existing note
obsidian-cli create "{note-name}" --content "abcde" --overwrite

# Creates note in default obsidian with content - append existing note
obsidian-cli create "{note-name}" --content "abcde" --append

# Creates note and opens it
obsidian-cli create "{note-name}" --content "abcde" --open

```

### Move / Rename Note

Moves a given note(path from top level of vault) with new name given (top level of vault). If given same path but different name then its treated as a rename. All links inside vault are updated to match new name.

```bash
# Renames a note in default obsidian
obsidian-cli move "{current-note-path}" "{new-note-path}"

# Renames a note and given obsidian
obsidian-cli move "{current-note-path}" "{new-note-path}" --vault "{vault-name}"

# Renames a note in default obsidian and opens it
obsidian-cli move "{current-note-path}" "{new-note-path}" --open
```

### Delete Note

Deletes a given note (path from top level of vault).

```bash
# Deletes a note in default obsidian
obsidian-cli delete "{note-path}"

# Deletes a note in given obsidian
obsidian-cli delete "{note-path}" --vault "{vault-name}"
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

# Find notes containing "project" and follow links, skipping anchored links
obsidian-cli list find:project -f --skip-anchors

# Find notes containing "notes" and follow links, skipping embedded links
obsidian-cli list find:notes -f --skip-embeds
```

### File Information

Show detailed information about a file including its frontmatter and all tags.

```bash
# Show file info in default vault
obsidian-cli info "Notes/Project.md"

# Show file info in specified vault
obsidian-cli info "Notes/Project.md" --vault "{vault-name}"
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
```

### Manage Tags

List, delete, or rename tags across your vault. By default, shows all tags with individual and aggregate counts.

**Note:** Delete and rename operations only affect tags with non-zero "Indiv" counts (exact tag matches). Tags with zero individual count are parent tags in hierarchies (e.g., `context` when only `context/work` exists) and won't be modified.

```bash
# List all tags with counts
obsidian-cli tags

# List tags as JSON
obsidian-cli tags --json

# List tags as markdown table
obsidian-cli tags --markdown

# Delete specific tags from all notes (multiple syntaxes supported)
obsidian-cli tags --delete work urgent              # Space-separated
obsidian-cli tags --delete work,urgent              # Comma-separated
obsidian-cli tags --delete work --delete urgent     # Repeated flags

# Rename tags across all notes (multiple syntaxes supported)
obsidian-cli tags --rename old --to new
obsidian-cli tags --rename foo bar --to baz         # Multiple source tags
obsidian-cli tags --rename foo,bar --to baz         # Comma-separated

# Preview changes without making them
obsidian-cli tags --delete work urgent --dry-run

# Control parallelism with custom worker count
obsidian-cli tags --delete work urgent --workers 4
```

## YAML formatting changes when editing tags

When tag-editing operations (delete/rename) touch the YAML front-matter, the block is re-emitted using Go’s `yaml.v3` encoder. As a result:

- Keys may appear in a different order than before.
- Nested items (like the `tags:` list) are indented with two spaces.
- Comments or blank lines inside the front-matter are stripped.
- Line endings are normalised to LF (`\n`).

This is cosmetic—note content is unaffected—but if you rely on exact front-matter formatting be aware the tool will produce canonical YAML instead of preserving original whitespace/comments.

## Contribution

Fork the project, add your feature or fix and submit a pull request. You can also open an [issue](https://github.com/yakitrak/obsidian-cli/issues/new/choose) to report a bug or request a feature.

## License

Available under [MIT License](./LICENSE)
