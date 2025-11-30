package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
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

func TestMoveNotesTool(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	// Seed notes
	if err := os.WriteFile(filepath.Join(vaultPath, "Note.md"), []byte("# Note"), 0o644); err != nil {
		t.Fatalf("failed to write note: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultPath, "Ref.md"), []byte("See [[Note]]"), 0o644); err != nil {
		t.Fatalf("failed to write ref: %v", err)
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

	tool := MoveNotesTool(cfg)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "move_notes",
			Arguments: map[string]interface{}{
				"moves": []map[string]string{
					{"source": "Note", "target": "Folder/Note"},
				},
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

	var parsed MoveNotesResponse
	if assert.NoError(t, json.Unmarshal([]byte(text.Text), &parsed)) {
		if assert.Len(t, parsed.Moves, 1) {
			assert.Equal(t, "Folder/Note.md", parsed.Moves[0].Target)
			assert.GreaterOrEqual(t, parsed.Moves[0].LinkUpdates, 1)
		}
	}

	updated, readErr := os.ReadFile(filepath.Join(vaultPath, "Ref.md"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(updated), "[[Folder/Note]]")
}

func TestListPropertiesTool(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatalf("failed to create vault dir: %v", err)
	}

	content := `---
office: AOGR
count: 2
---`
	if err := os.WriteFile(filepath.Join(vaultPath, "note.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write note.md: %v", err)
	}

	obsidianConfig := filepath.Join(tempDir, "obsidian.json")
	configJSON := `{"vaults":{"vault":{"path":"` + vaultPath + `"}}}`
	if err := os.WriteFile(obsidianConfig, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write obsidian config: %v", err)
	}

	origConfig := obsidian.ObsidianConfigFile
	obsidian.ObsidianConfigFile = func() (string, error) { return obsidianConfig, nil }
	defer func() { obsidian.ObsidianConfigFile = origConfig }()

	cfg := Config{
		Vault:     &obsidian.Vault{Name: "vault"},
		VaultPath: vaultPath,
		Debug:     false,
	}

	tool := ListPropertiesTool(cfg)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_properties",
			Arguments: map[string]interface{}{
				"enumThreshold": float64(5),
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

	var parsed PropertyListResponse
	err = json.Unmarshal([]byte(text.Text), &parsed)
	assert.NoError(t, err)
	assert.Len(t, parsed.Properties, 2)

	var officeFound bool
	for _, p := range parsed.Properties {
		if p.Name == "office" {
			officeFound = true
			assert.Equal(t, 1, p.NoteCount)
			assert.Equal(t, []string{"AOGR"}, p.EnumValues)
		}
	}
	assert.True(t, officeFound, "expected office property in response")
}
