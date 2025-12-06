package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"random": map[string]any{"path": vaultPath},
		},
	})
	if err := os.WriteFile(configFile, configBody, 0o644); err != nil {
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
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"random": map[string]any{"path": vaultPath},
		},
	})
	if err := os.WriteFile(configFile, configBody, 0o644); err != nil {
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
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"vault": map[string]any{"path": vaultPath},
		},
	})
	if err := os.WriteFile(obsidianConfig, configBody, 0o644); err != nil {
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
				"valueLimit": float64(5),
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

func TestListPropertiesToolWithOnlyRaisesValueLimit(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatalf("failed to create vault dir: %v", err)
	}

	var contentBuilder strings.Builder
	contentBuilder.WriteString("---\nwho:\n")
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&contentBuilder, "  - Person %02d\n", i)
	}
	contentBuilder.WriteString("omit: skip\n---")

	if err := os.WriteFile(filepath.Join(vaultPath, "note.md"), []byte(contentBuilder.String()), 0o644); err != nil {
		t.Fatalf("failed to write note.md: %v", err)
	}

	obsidianConfig := filepath.Join(tempDir, "obsidian.json")
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"vault": map[string]any{"path": vaultPath},
		},
	})
	if err := os.WriteFile(obsidianConfig, configBody, 0o644); err != nil {
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
				"only": []interface{}{"who"},
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
	if !assert.Len(t, parsed.Properties, 1) {
		return
	}

	assert.Equal(t, "who", parsed.Properties[0].Name)
	assert.Equal(t, 30, parsed.Properties[0].DistinctValueCount)
	assert.False(t, parsed.Properties[0].TruncatedValueSet)
	assert.Len(t, parsed.Properties[0].EnumValues, 30, "valueLimit should be raised to enumerate all values")
}

func TestGraphStatsTool(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault")
	require.NoError(t, os.MkdirAll(vaultPath, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(vaultPath, "A.md"), []byte("Link to [[B]]"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(vaultPath, "B.md"), []byte(""), 0o644))

	configFile := filepath.Join(root, "obsidian.json")
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"random": map[string]any{"path": vaultPath},
		},
	})
	require.NoError(t, os.WriteFile(configFile, configBody, 0o644))
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

	tool := GraphStatsTool(cfg)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "graph_stats",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := tool(context.Background(), req)
	require.NoError(t, err)
	if !assert.Len(t, resp.Content, 1) {
		return
	}

	text, ok := resp.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", resp.Content[0])
	}
	assert.Contains(t, text.Text, "graph_stats removed")
}

func TestCommunityListAndDetailTools(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault")
	require.NoError(t, os.MkdirAll(vaultPath, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(vaultPath, "A.md"), []byte("Link to [[B]]"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(vaultPath, "B.md"), []byte("Link to [[A]]"), 0o644))

	configFile := filepath.Join(root, "obsidian.json")
	configBody, _ := json.Marshal(map[string]any{
		"vaults": map[string]any{
			"random": map[string]any{"path": vaultPath},
		},
	})
	require.NoError(t, os.WriteFile(configFile, configBody, 0o644))
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

	listTool := CommunityListTool(cfg)
	listReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "community_list",
			Arguments: map[string]interface{}{},
		},
	}
	listResp, err := listTool(context.Background(), listReq)
	require.NoError(t, err)
	require.Len(t, listResp.Content, 1)
	listText, ok := listResp.Content[0].(mcp.TextContent)
	require.True(t, ok)

	var listParsed CommunityListResponse
	require.NoError(t, json.Unmarshal([]byte(listText.Text), &listParsed))
	require.NotEmpty(t, listParsed.Communities)

	commID := listParsed.Communities[0].ID

	detailTool := CommunityDetailTool(cfg)
	detailReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "community_detail",
			Arguments: map[string]interface{}{
				"id": commID,
			},
		},
	}

	detailResp, err := detailTool(context.Background(), detailReq)
	require.NoError(t, err)
	require.Len(t, detailResp.Content, 1)
	detailText, ok := detailResp.Content[0].(mcp.TextContent)
	require.True(t, ok)

	var detailParsed CommunityDetailResponse
	require.NoError(t, json.Unmarshal([]byte(detailText.Text), &detailParsed))
	assert.Equal(t, commID, detailParsed.ID)
	require.Len(t, detailParsed.Members, 2)

	fileReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "community_detail",
			Arguments: map[string]interface{}{
				"file": filepath.Join(vaultPath, "A.md"),
			},
		},
	}

	fileResp, err := detailTool(context.Background(), fileReq)
	require.NoError(t, err)
	require.Len(t, fileResp.Content, 1)
	fileText, ok := fileResp.Content[0].(mcp.TextContent)
	require.True(t, ok)

	var fileParsed CommunityDetailResponse
	require.NoError(t, json.Unmarshal([]byte(fileText.Text), &fileParsed))
	assert.Equal(t, commID, fileParsed.ID)
	require.Len(t, fileParsed.Members, 2)
}
