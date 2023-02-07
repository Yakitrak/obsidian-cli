# Obsidian CLI Tool

![Obs Usage](./docs/obs-usage.png)

CLI made to search, open and creates files in Obsidian. Written in [Go](https://go.dev/), built on top of existing [Obsidian URIs](https://help.obsidian.md/Advanced+topics/Using+obsidian+URI)

## Description

[Obsidian](https://obsidian.md/) is a powerful and extensible knowledge base application
that works on top of your local folder of plain text files. This CLI tool will let you interact with the application using the terminal. You are currently able to open, search and create files.

## Install

### Homebrew

```Bash
brew tap yakitrak/yakitrak
```

```Bash
brew install yakitrak/yakitrak/obs
```

### Manual

1. Download the file from https://github.com/Yakitrak/obsidian-cli/releases
2. Set it to path:

- Mac: `sudo cp {name-of-file} /usr/local/bin/obs`
- Linux: `sudo cp {name-of-file} /usr/local/bin/obs`
- Windows: Set `{name-of-file}.exe` to path using [this](https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/)

## Usage

### Help

```bash
# See All command instructions
obs --help
```

### Set Default Vault

Defines default vault for future usage. If not set, pass `--vault` flag for other commands

```bash
obs set-default "{vault-name}"
```

### Open File

Open given file name in Obsidian. File can also be a path.

```bash
# Opens file in obsidian
obs open "{file}"

# Opens file in specified vault
obs open "{file}" --vault "{vault-name}"

```

### Search File

Opens obsidian search tab with given search text

```bash
# Searches in default vault
obs search "{search-text}"

# Searches in specified vault
obs search "{search-text}" --vault "{vault-name}"

```

### Create / Update File

Creates file (can also be a path with name) in vault. By default if the file exists, it will create another file but passing `--overwrite` or `--append` can be used to edit the named file.

```bash
# Creates empty file in default vault and opens it
obs create "{file}"

# Creates empty file in given vault and opens it
obs create "{file}"  --vault "{vault-name}"

# Creates file in default vault with content
obs create "{file}" --content "abcde"

# Creates file in default vault with content - overwrite existing file
obs create "{file}" --content "abcde" --overwrite

# Creates file in default vault with content - append existing file
obs create "{file}" --content "abcde" --append

```
