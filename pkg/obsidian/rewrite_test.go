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
