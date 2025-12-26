# Obsidian CLI

---

## CLI Help (Generated)

```text
$ obsidian-cli --help
obsidian-cli - CLI to open, search, move, create, delete and update notes

Usage:
  obsidian-cli [command]

Available Commands:
  alias          Generate a shell alias snippet or install a symlink shortcut
  completion     Generate the autocompletion script for the specified shell
  create         Creates note in vault
  daily          Creates or opens daily note in vault
  delete         Delete note in vault
  help           Help about any command
  move           Move or rename note in vault and update corresponding links
  open           Opens note in vault by note name
  print          Print contents of note
  print-default  Prints default vault name and path
  search         Fuzzy searches and opens note in vault
  search-content Search note content for search term
  set-default    Sets default vault

Flags:
  -h, --help      help for obsidian-cli
  -v, --version   version for obsidian-cli

Use "obsidian-cli [command] --help" for more information about a command.
```

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

For detailed help (including examples) for a specific command:

```bash
obsidian-cli <command> --help
```

### Command Shortcut (Alias)

If you want a shorter command name (for example `obsi`), you can either:

- Create a shell alias (session-scoped unless you add it to your shell profile):

  ```bash
  # zsh/bash
  eval "$(obsidian-cli alias obsi --shell zsh)"
  ```

- Or install a persistent symlink shortcut (recommended):

  ```bash
  obsidian-cli alias obsi --symlink --dir "$HOME/.local/bin"
  ```

### Editor Flag

The `search`, `search-content`, `create`, and `move` commands support the `--editor` (or `-e`) flag, which opens notes in your default text editor instead of the Obsidian application. This is useful for quick edits or when working in a terminal-only environment.

The editor is determined by the `EDITOR` environment variable. If not set, it defaults to `vim`.

**Supported editors:**

- Terminal editors: vim, nano, emacs, etc.
- GUI editors with wait flag: VSCode (`code`), Sublime Text (`subl`), Atom, TextMate
  - The CLI automatically adds the `--wait` flag for supported GUI editors to ensure they block until you close the file

**Example:**

```bash
# Set your preferred editor (add to ~/.zshrc or ~/.bashrc to make permanent)
export EDITOR="code"  # or "vim", "nano", "subl", etc.

# Use with supported commands
obsidian-cli search --editor
obsidian-cli search-content "term" --editor
obsidian-cli create "note.md" --open --editor
obsidian-cli move "old.md" "new.md" --open --editor
```

### Set Default Vault

Defines default vault for future usage. If not set, pass `--vault` flag for other commands. You don't provide the path to vault here, just the name.

```bash
obsidian-cli set-default "{vault-name}"
```

Note: `open` and other commands in `obsidian-cli` use this vault's base directory as the working directory, not the current working directory of your terminal.

`set-default` stores the default vault name in `preferences.json` under your OS user config directory (`os.UserConfigDir()`), at:

- `obsidian-cli/preferences.json`

This preferences file also supports optional per-vault settings under `vault_settings` (for example, daily note configuration). For example:

```json
{
  "default_vault_name": "My Vault",
  "vault_settings": {
    "My Vault": {
      "daily_note": {
        "folder": "Daily",
        "filename_pattern": "{YYYY-MM-DD}",
        "template_path": "Templates/Daily.md",
        "create_if_missing": true
      }
    }
  }
}
```

<details>
<summary><code>set-default</code> command reference (help, flags, aliases)</summary>

```text
$ obsidian-cli set-default --help
Sets the default vault for all commands.

The vault name must match exactly as it appears in Obsidian.
Once set, you won't need to specify --vault for each command.

Usage:
  obsidian-cli set-default <vault> [flags]

Aliases:
  set-default, sd

Examples:
  # Set default vault
  obsidian-cli set-default "My Vault"

  # Verify it worked
  obsidian-cli print-default

Flags:
  -h, --help   help for set-default
```

</details>

### Print Default Vault

Prints default vault and path. Please set this with `set-default` command if not set.

```bash
# print the default vault name and path
obsidian-cli print-default

# print only the vault path
obsidian-cli print-default --path-only
```

<details>
<summary><code>print-default</code> command reference (help, flags, aliases)</summary>

```text
$ obsidian-cli print-default --help
Shows the currently configured default vault.

Use --path-only to output just the path, useful for scripting.

Usage:
  obsidian-cli print-default [flags]

Aliases:
  print-default, pd

Examples:
  # Show default vault info
  obsidian-cli print-default

  # Get just the path (for scripts)
  obsidian-cli print-default --path-only

Flags:
  -h, --help        help for print-default
      --path-only   print only the vault path
```

</details>

You can add this to your shell configuration file (like `~/.zshrc`) to quickly navigate to the default vault:

```bash
obs_cd() {
    local result=$(obsidian-cli print-default --path-only)
    [ -n "$result" ] && cd -- "$result"
}
```

Then you can use `obs_cd` to navigate to the default vault directory within your terminal.

### Config Files

`obsidian-cli` reads and writes configuration under your OS user config directory (`os.UserConfigDir()`):

- `obsidian-cli/preferences.json` (default vault name + optional per-vault `vault_settings`)

It also reads Obsidian’s vault list from:

- `obsidian/obsidian.json` (Obsidian config, used for vault discovery)

Note: when writing `preferences.json`, the CLI attempts to create the config directory with mode `0750` and the file with mode `0600` (confirmed from `os.MkdirAll(…, 0750)` / `os.WriteFile(…, 0600)` in code).

### Open Note

Open given note name in Obsidian. Note can also be an absolute path from top level of vault.

```bash
# Opens note in obsidian vault
obsidian-cli open "{note-name}"

# Opens note in specified obsidian vault
obsidian-cli open "{note-name}" --vault "{vault-name}"

```

### Daily Note

Open the daily note in Obsidian (via Obsidian URI).

Note: creation/templates are controlled by Obsidian’s daily note settings/plugins.

```bash
# Creates / opens daily note in obsidian vault
obsidian-cli daily

# Creates / opens daily note in specified obsidian vault
obsidian-cli daily --vault "{vault-name}"

```

### Search Note

Starts a fuzzy search displaying notes in the terminal from the vault. You can hit enter on a note to open that in Obsidian.

```bash
# Searches in default obsidian vault
obsidian-cli search

# Searches in specified obsidian vault
obsidian-cli search --vault "{vault-name}"

# Searches and opens selected note in your default editor
obsidian-cli search --editor

```

### Search Note Content

Searches for notes containing search term in the content of notes. It will display a list of matching notes with the line number and a snippet of the matching line. You can hit enter on a note to open that in Obsidian.

```bash
# Searches for content in default obsidian vault
obsidian-cli search-content "search term"

# Searches for content in specified obsidian vault
obsidian-cli search-content "search term" --vault "{vault-name}"

# Searches and opens selected note in your default editor
obsidian-cli search-content "search term" --editor

```

### Print Note

Prints the contents of given note name or path in Obsidian.

```bash
# Prints note in default vault
obsidian-cli print "{note-name}"

# Prints note by path in default vault
obsidian-cli print "{note-path}"

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

# Creates note and opens it in your default editor
obsidian-cli create "{note-name}" --content "abcde" --open --editor

```

### Move / Rename Note

Moves a given note (path from top level of vault) to a new path. If given the same path but a different name, it's treated as a rename.

When moving/renaming, `obsidian-cli` updates links inside your vault to match the new location, including:

- Wikilinks: `[[note]]`, `[[folder/note]]`, `[[folder/note|alias]]`, `[[folder/note#heading]]`
- Markdown links: `[text](folder/note.md)`, `[text](./folder/note.md)`, and the same forms without the `.md` extension

```bash
# Renames a note in default obsidian
obsidian-cli move "{current-note-path}" "{new-note-path}"

# Renames a note and given obsidian
obsidian-cli move "{current-note-path}" "{new-note-path}" --vault "{vault-name}"

# Renames a note in default obsidian and opens it
obsidian-cli move "{current-note-path}" "{new-note-path}" --open

# Renames a note and opens it in your default editor
obsidian-cli move "{current-note-path}" "{new-note-path}" --open --editor
```

### Delete Note

Deletes a given note (path from top level of vault).

If other notes link to the note, `delete` prints the incoming links and prompts for confirmation. The default is **No** (press Enter to cancel).

Use `--force` (`-f`) to skip confirmation (recommended for scripts). Alias: `delete, del`. Heads up: `daily` uses alias `d`, so `delete` uses `del` to avoid ambiguity.

```bash
# Delete a note in the default vault
obsidian-cli delete "{note-path}"

# Force delete without prompt
obsidian-cli delete "{note-path}" --force

# Delete a note in a specific vault
obsidian-cli delete "{note-path}" --vault "{vault-name}"
```

<details>
<summary><code>delete</code> command reference (help, flags, aliases)</summary>

```text
$ obsidian-cli delete --help
Delete a note from the vault.

If other notes link to the note, you'll be prompted to confirm.
Use --force to skip confirmation (recommended for scripts).

Usage:
  obsidian-cli delete <note> [flags]

Aliases:
  delete, del

Examples:
  # Delete a note (prompts if linked)
  obsidian-cli delete "old-note"

  # Force delete without prompt
  obsidian-cli delete "temp" --force

  # Delete from specific vault
  obsidian-cli delete "note" --vault "Archive"

Flags:
  -f, --force          skip confirmation if the note has incoming links
  -h, --help           help for delete
  -v, --vault string   vault name
```

</details>

## Contribution

Fork the project, add your feature or fix and submit a pull request. You can also open an [issue](https://github.com/yakitrak/obsidian-cli/issues/new/choose) to report a bug or request a feature.

## Acknowledgements

- Link-update support for path-based wikilinks and markdown links builds on upstream PR #58: https://github.com/Yakitrak/obsidian-cli/pull/58

## License

Available under [MIT License](./LICENSE)
