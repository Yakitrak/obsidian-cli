package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRenameNoteTool(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	// Seed notes
	if err := os.WriteFile(filepath.Join(vaultPath, "Old.md"), []byte("# Old"), 0o644); err != nil {
		t.Fatalf("failed to write old note: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultPath, "Ref.md"), []byte("See [[Old|Alias]]"), 0o644); err != nil {
		t.Fatalf("failed to write ref note: %v", err)
	}

	// Configure obsidian vault resolution
	configFile := filepath.Join(root, "obsidian.json")
	configJSON := `{"vaults":{"random":{"path":"` + vaultPath + `"}}}`
	if err := os.WriteFile(configFile, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write obsidian config: %v", err)
	}
	origConfig := obsidian.ObsidianConfigFile
	obsidian.ObsidianConfigFile = func() (string, error) { return configFile, nil }
	defer func() { obsidian.ObsidianConfigFile = origConfig }()

	cfg := Config{
		Vault:          &obsidian.Vault{Name: "vault"},
		VaultPath:      vaultPath,
		Debug:          false,
		SuppressedTags: []string{},
		ReadWrite:      true,
	}

	tool := RenameNoteTool(cfg)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "rename_note",
			Arguments: map[string]interface{}{
				"source":          "Old",
				"target":          "New",
				"overwrite":       false,
				"updateBacklinks": true,
			},
		},
	}

	resp, err := tool(context.Background(), req)
	assert.NoError(t, err)
	if !assert.Len(t, resp.Content, 1) {
		return
	}

	text, ok := resp.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", resp.Content[0])
	}

	var parsed RenameNoteResponse
	if assert.NoError(t, json.Unmarshal([]byte(text.Text), &parsed)) {
		assert.Equal(t, "New.md", parsed.Path)
		assert.True(t, parsed.LinkUpdates >= 1)
	}

	newRef, readErr := os.ReadFile(filepath.Join(vaultPath, "Ref.md"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(newRef), "[[New|Alias]]")
}
