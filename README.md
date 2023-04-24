# Obsidian CLI

![Obs Usage](./docs/obs-usage.png)

## Description

Obsidian is a powerful and extensible knowledge base application 
that works on top of your local folder of plain text notes. This CLI tool (written in Go) will let you interact with the application using the terminal. You are currently able to open, search, move, create, update and delete notes.

---

[Installation and Usage Guide](https://yakitrak.github.io/obs/)

---

## Install

```Bash
brew tap yakitrak/yakitrak
```

```Bash
brew install yakitrak/yakitrak/obs
```

For full installation instructions, see [manual](https://yakitrak.github.io/obs/docs/install/mac-and-linux).

## Usage

### Help

```bash
# See All command instructions
obs --help
```

### Set Default Vault

Defines default vault for future usage. If not set, pass `--vault` flag for other commands. You don't provide the path to vault here, just the name.

```bash
obs set-default "{vault-name}"
```

Note: `open` and other commands in `obs` use this vault's base directory as the working directory, not the current working directory of your terminal.

### Print Default Vault

Prints default vault and path. Please set this with `set-default` command if not set.

```bash
obs print-default
```

### Open Note

Open given note name in Obsidian. Note can also be an absolute path from top level of vault.

```bash
# Opens note in obsidian
obs open "{note-name}"

# Opens note in specified obsidian
obs open "{note-name}" --vault "{vault-name}"

```

### Search Note

Opens obsidian search tab with given search text

```bash
# Searches in default obsidian
obs search "{search-text}"

# Searches in specified obsidian
obs search "{search-text}" --vault "{vault-name}"

```

### Create / Update Note

Creates note (can also be a path with name) in vault. By default if the note exists, it will create another note but passing `--overwrite` or `--append` can be used to edit the named note.

```bash
# Creates empty note in default obsidian and opens it
obs create "{note-name}"

# Creates empty note in given obsidian and opens it
obs create "{note-name}"  --vault "{vault-name}"

# Creates note in default obsidian with content
obs create "{note-name}" --content "abcde"

# Creates note in default obsidian with content - overwrite existing note
obs create "{note-name}" --content "abcde" --overwrite

# Creates note in default obsidian with content - append existing note
obs create "{note-name}" --content "abcde" --append

```

### Move / Rename Note

Moves a given note(path from top level of vault) with new name given (top level of vault). If given same path but different name then its treated as a rename. All links inside vault are updated to match new name.

```bash
# Renames a note in default obsidian
obs move "{current-note-path}" "{new-note-path}"

# Renames a note and given obsidian
obs move "{current-note-path}" "{new-note-path}" --vault "{vault-name}"

# Renames a note in default obsidian and opens it
obs move "{current-note-path}" "{new-note-path}" --open
```

### Delete Note

Deletes a given note (path from top level of vault).

```bash
# Renames a note in default obsidian
obs delete "{note-path}" 

# Renames a note in given obsidian
obs delete "{note-path}" --vault "{vault-name}"
```



