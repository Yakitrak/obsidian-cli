package obsidian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteLinksInContent_Wikilinks(t *testing.T) {
	content := "See [[Old Note]] and [[Old Note|Alias]] and [[Old Note#Heading]] and [[Old Note#^block|Alias]]."
	rewritten, count := RewriteLinksInContent(content, "Old Note.md", "Folder/New Note.md")

	assert.Equal(t, 4, count)
	assert.Contains(t, rewritten, "[[Folder/New Note]]")
	assert.Contains(t, rewritten, "[[Folder/New Note|Alias]]")
	assert.Contains(t, rewritten, "[[Folder/New Note#Heading]]")
	assert.Contains(t, rewritten, "[[Folder/New Note#^block|Alias]]")
}

func TestRewriteLinksInContent_EmbedAndNoExt(t *testing.T) {
	content := "![[Old Note]] and [[Other Note]]"
	rewritten, count := RewriteLinksInContent(content, "Old Note", "New Note.md")

	assert.Equal(t, 1, count)
	assert.Contains(t, rewritten, "![[New Note]]")
	assert.Contains(t, rewritten, "[[Other Note]]")
}

func TestRewriteLinksInContent_MarkdownLinks(t *testing.T) {
	content := "See [text](Old Note.md) and [section](Old Note.md#heading) and [ext](https://example.com)."
	rewritten, count := RewriteLinksInContent(content, "Old Note.md", "Folder/New Note.md")

	assert.Equal(t, 2, count)
	assert.Contains(t, rewritten, "[text](Folder/New Note.md)")
	assert.Contains(t, rewritten, "[section](Folder/New Note.md#heading)")
	assert.Contains(t, rewritten, "[ext](https://example.com)")
}

func TestRewriteLinksInContent_NoMatch(t *testing.T) {
	content := "Nothing to change [[Another Note]] and [link](another.md)."
	rewritten, count := RewriteLinksInContent(content, "Old Note.md", "New Note.md")

	assert.Equal(t, 0, count)
	assert.Equal(t, content, rewritten)
}

func TestRewriteLinksInContent_WikilinkWithoutFolder(t *testing.T) {
	// Bug fix: wikilinks often omit the folder path, but oldPath includes it
	content := "See [[Old Note]] and [[Old Note|Alias]]."
	rewritten, count := RewriteLinksInContent(content, "Notes/Old Note.md", "Notes/New Note.md")

	assert.Equal(t, 2, count)
	// Links should be rewritten to basename only (preserving original style)
	assert.Contains(t, rewritten, "[[New Note]]")
	assert.Contains(t, rewritten, "[[New Note|Alias]]")
}

func TestRewriteLinksInContent_WikilinkWithFolder(t *testing.T) {
	// When the wikilink includes the folder, preserve it
	content := "See [[Notes/Old Note]] and [[Notes/Old Note|Alias]]."
	rewritten, count := RewriteLinksInContent(content, "Notes/Old Note.md", "Archive/New Note.md")

	assert.Equal(t, 2, count)
	// Links should be rewritten with full path
	assert.Contains(t, rewritten, "[[Archive/New Note]]")
	assert.Contains(t, rewritten, "[[Archive/New Note|Alias]]")
}

func TestRewriteLinksInContent_MarkdownLinkWithoutFolder(t *testing.T) {
	// Bug fix: markdown links often omit the folder path
	content := "See [text](Old Note.md) and [section](Old Note.md#heading)."
	rewritten, count := RewriteLinksInContent(content, "Notes/Old Note.md", "Notes/New Note.md")

	assert.Equal(t, 2, count)
	// Links should be rewritten to basename only (preserving original style)
	assert.Contains(t, rewritten, "[text](New Note.md)")
	assert.Contains(t, rewritten, "[section](New Note.md#heading)")
}

func TestRewriteLinksInContentWithOptions_DuplicateBasename(t *testing.T) {
	// When basenameUnique is false, only match full paths (not basename-only links)
	// This prevents incorrectly rewriting links that might point to a different file with the same name
	content := "See [[Old Note]] and [[Notes/Old Note]] and [[Archive/Old Note]]."

	// With basenameUnique=false, only the explicit path match should be rewritten
	rewritten, count := RewriteLinksInContentWithOptions(content, "Notes/Old Note.md", "Notes/New Note.md", false)

	assert.Equal(t, 1, count)
	assert.Contains(t, rewritten, "[[Old Note]]")             // NOT rewritten (ambiguous)
	assert.Contains(t, rewritten, "[[Notes/New Note]]")       // Rewritten (explicit match)
	assert.Contains(t, rewritten, "[[Archive/Old Note]]")     // NOT rewritten (different folder)
}

func TestRewriteLinksInContentWithOptions_UniqueBasename(t *testing.T) {
	// When basenameUnique is true, basename-only links should be rewritten
	content := "See [[Old Note]] and [[Notes/Old Note]] and [[Archive/Old Note]]."

	// With basenameUnique=true, both basename and explicit path matches should be rewritten
	rewritten, count := RewriteLinksInContentWithOptions(content, "Notes/Old Note.md", "Notes/New Note.md", true)

	assert.Equal(t, 2, count)
	assert.Contains(t, rewritten, "[[New Note]]")             // Rewritten (basename match, unique)
	assert.Contains(t, rewritten, "[[Notes/New Note]]")       // Rewritten (explicit match)
	assert.Contains(t, rewritten, "[[Archive/Old Note]]")     // NOT rewritten (different folder)
}

func TestRewriteLinksInContent_CaseInsensitive(t *testing.T) {
	// On macOS/Windows, Obsidian links are case-insensitive
	// This test verifies the behavior matches the current OS
	content := "See [[old note]] and [[NOTES/OLD NOTE]] and [[Old Note]]."
	rewritten, count := RewriteLinksInContent(content, "Notes/Old Note.md", "Notes/New Note.md")

	if IsCaseInsensitiveFS() {
		// macOS/Windows: all three should match
		assert.Equal(t, 3, count)
		assert.Contains(t, rewritten, "[[New Note]]")         // Was [[old note]] - basename match
		assert.Contains(t, rewritten, "[[Notes/New Note]]")   // Was [[NOTES/OLD NOTE]] - full path match
	} else {
		// Linux: only exact case matches
		assert.Equal(t, 1, count)
		assert.Contains(t, rewritten, "[[old note]]")         // NOT rewritten (case mismatch)
		assert.Contains(t, rewritten, "[[NOTES/OLD NOTE]]")   // NOT rewritten (case mismatch)
		assert.Contains(t, rewritten, "[[Notes/New Note]]")   // Rewritten (exact match)
	}
}

func TestNormalizeForComparison(t *testing.T) {
	if IsCaseInsensitiveFS() {
		// macOS/Windows: should lowercase
		assert.Equal(t, "old note", NormalizeForComparison("Old Note"))
		assert.Equal(t, "old note", NormalizeForComparison("OLD NOTE"))
	} else {
		// Linux: should preserve case
		assert.Equal(t, "Old Note", NormalizeForComparison("Old Note"))
		assert.Equal(t, "OLD NOTE", NormalizeForComparison("OLD NOTE"))
	}
}
