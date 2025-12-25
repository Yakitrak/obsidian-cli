package obsidian_test

import (
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestValidateTargetName(t *testing.T) {
	assert.Error(t, obsidian.ValidateTargetName(""))
	assert.Error(t, obsidian.ValidateTargetName(" "))
	assert.Error(t, obsidian.ValidateTargetName("has space"))
	assert.Error(t, obsidian.ValidateTargetName("add"))
	assert.NoError(t, obsidian.ValidateTargetName("inbox"))
}

func TestTargetUnmarshalScalarTreatsAsFile(t *testing.T) {
	var cfg obsidian.TargetsConfig
	err := yaml.Unmarshal([]byte("inbox: Inbox.md\n"), &cfg)
	assert.NoError(t, err)
	tg := cfg["inbox"]
	assert.Equal(t, "file", tg.Type)
	assert.Equal(t, "Inbox.md", tg.File)
}

func TestResolveTargetNotePath(t *testing.T) {
	vault := t.TempDir()
	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)

	t.Run("file target", func(t *testing.T) {
		rel, abs, err := obsidian.ResolveTargetNotePath(vault, obsidian.Target{Type: "file", File: "Inbox"}, now)
		assert.NoError(t, err)
		assert.Equal(t, "Inbox", rel)
		assert.Contains(t, abs, "Inbox.md")
	})

	t.Run("folder target expands pattern", func(t *testing.T) {
		rel, abs, err := obsidian.ResolveTargetNotePath(vault, obsidian.Target{Type: "folder", Folder: "Daily", Pattern: "YYYY-MM-DD_HH"}, now)
		assert.NoError(t, err)
		assert.Equal(t, "Daily/2024-01-15_14", rel)
		assert.Contains(t, abs, "Daily")
		assert.Contains(t, abs, "2024-01-15_14.md")
	})

	t.Run("reject escape", func(t *testing.T) {
		_, _, err := obsidian.ResolveTargetNotePath(vault, obsidian.Target{Type: "file", File: "../escape.md"}, now)
		assert.Error(t, err)
	})
}

func TestPlanTargetAppend(t *testing.T) {
	vault := t.TempDir()
	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)

	plan, err := obsidian.PlanTargetAppend(vault, "inbox", obsidian.Target{Type: "file", File: "Inbox.md"}, now)
	assert.NoError(t, err)
	assert.Equal(t, "inbox", plan.TargetName)
	assert.True(t, plan.WillCreateFile)
	assert.False(t, plan.WillCreateDirs)
}
