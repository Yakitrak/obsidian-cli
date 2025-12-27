package obsidian_test

import (
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestExpandTemplateVariablesAt(t *testing.T) {
	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)
	in := []byte("{{title}}\n{{date}}\n{{time}}\n{{date:YYYY-[ToDo]-MM}}\n{{time:HH:mm:ss}}\n")
	out := string(obsidian.ExpandTemplateVariablesAt(in, "Note.md", now))
	assert.Contains(t, out, "Note\n")
	assert.Contains(t, out, "2024-01-15\n")
	assert.Contains(t, out, "14:30\n")
	assert.Contains(t, out, "2024-ToDo-01\n")
	assert.Contains(t, out, "14:30:52\n")
}
