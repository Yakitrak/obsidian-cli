package obsidian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetFrontmatterPropertyCreatesBlock(t *testing.T) {
	content := "# Title\nbody"
	updated, changed, err := SetFrontmatterProperty(content, "status", "done", false)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, updated, "status: done")
	assert.Contains(t, updated, "Title")
}

func TestSetFrontmatterPropertyRespectsOverwrite(t *testing.T) {
	content := "---\nstatus: open\n---\nbody"
	updated, changed, err := SetFrontmatterProperty(content, "status", "done", false)
	assert.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, content, updated)

	updated, changed, err = SetFrontmatterProperty(content, "status", "done", true)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, updated, "status: done")
}

func TestDeleteFrontmatterProperties(t *testing.T) {
	content := "---\nstatus: open\nowner: me\n---\nbody"
	updated, changed, err := DeleteFrontmatterProperties(content, []string{"status"})
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.NotContains(t, updated, "status:")
	assert.Contains(t, updated, "owner:")
}

func TestRenameFrontmatterPropertiesMerge(t *testing.T) {
	content := "---\nstatus: open\nStatus: pending\nlabels:\n  - a\n---\nbody"
	updated, changed, err := RenameFrontmatterProperties(content, []string{"status"}, "labels", true)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, updated, "labels:")
	assert.Contains(t, updated, "- a")
	assert.Contains(t, updated, "- open")
	assert.NotContains(t, updated, "status:")
}

func TestRenameFrontmatterPropertiesNoMergeKeepsDestination(t *testing.T) {
	content := "---\nstatus: open\nlabels: a\n---\nbody"
	updated, changed, err := RenameFrontmatterProperties(content, []string{"status"}, "labels", false)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Contains(t, updated, "labels: a")
	assert.NotContains(t, updated, "open")
}
