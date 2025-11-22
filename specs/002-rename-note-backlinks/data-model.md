# Data Model: Rename Note with Backlink Updates

## Entities

### Vault
- **Attributes**: rootPath, isGitRepo, ignoredPaths, defaultFileExtension
- **Relationships**: contains Notes; provides context for Backlink scanning
- **Validation**: rootPath must exist and be readable; ignoredPaths honored during scans/rewrites.

### Note
- **Attributes**: path, title, content, aliases (from links), anchors (headers, block IDs)
- **Relationships**: belongs to Vault; targeted by Backlinks.
- **Validation**: path must be unique within vault; content must be readable before rename; target path must not overwrite existing note unless explicitly allowed.

### Backlink
- **Attributes**: sourcePath, linkText, linkTarget, linkType (wikilink, markdown link, embed), fragment (header/block ref), aliasText
- **Relationships**: points from one Note to another.
- **Validation**: updates must preserve aliasText and fragment while adjusting linkTarget to new path; links in ignoredPaths are skipped.

### Rename Request
- **Attributes**: sourcePath, targetPath, invocationContext (CLI or MCP), overwriteAllowed (default false)
- **Relationships**: operates within a Vault and affects Notes and Backlinks.
- **Validation**: sourcePath must exist; targetPath must be a valid note path; overwriteAllowed required to replace existing target; when isGitRepo true, working tree must permit git rename.

### Operation Result
- **Attributes**: renamedPath, linkUpdateCount, skippedFiles, errors (if any), gitHistoryPreserved (bool), summaryMessage
- **Relationships**: produced from processing a Rename Request.
- **Validation**: summaryMessage must reflect success/failure and enumerate skips/errors; errors list is empty on success.
