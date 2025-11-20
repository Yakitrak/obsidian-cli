package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAll(t *testing.T) {
	// Create a test server
	s := server.NewMCPServer(
		"test-obsidian-cli",
		"v0.15.0",
		server.WithToolCapabilities(false),
	)

	// Create a test config
	vault := &obsidian.Vault{Name: "test-vault"}
	config := Config{
		Vault:          vault,
		VaultPath:      "/tmp/test-vault",
		Debug:          false,
		SuppressedTags: []string{"no-prompt"},
		ReadWrite:      false,
	}

	// Test that RegisterAll doesn't return an error
	err := RegisterAll(s, config)
	if err != nil {
		t.Errorf("RegisterAll returned error: %v", err)
	}

	// Note: We can't easily test the actual tool functionality without
	// setting up a full MCP client/server communication, but we can
	// verify that the registration completes successfully
}

func TestConfig(t *testing.T) {
	vault := &obsidian.Vault{Name: "test"}
	config := Config{
		Vault:          vault,
		VaultPath:      "/tmp",
		Debug:          true,
		SuppressedTags: []string{"test-tag"},
		ReadWrite:      true,
	}

	if config.Vault.Name != "test" {
		t.Errorf("Expected vault name 'test', got %s", config.Vault.Name)
	}

	if config.VaultPath != "/tmp" {
		t.Errorf("Expected vault path '/tmp', got %s", config.VaultPath)
	}

	if !config.Debug {
		t.Error("Expected debug to be true")
	}

	if len(config.SuppressedTags) != 1 || config.SuppressedTags[0] != "test-tag" {
		t.Errorf("Expected suppressed tags ['test-tag'], got %v", config.SuppressedTags)
	}

	if !config.ReadWrite {
		t.Error("Expected read-write to be true")
	}
}

func TestRegisterAllWithReadWrite(t *testing.T) {
	// Create a test server
	s := server.NewMCPServer(
		"test-obsidian-cli",
		"v0.15.0",
		server.WithToolCapabilities(false),
	)

	// Create a test config with ReadWrite enabled
	vault := &obsidian.Vault{Name: "test-vault"}
	config := Config{
		Vault:          vault,
		VaultPath:      "/tmp/test-vault",
		Debug:          false,
		SuppressedTags: []string{"no-prompt"},
		ReadWrite:      true, // Enable destructive operations
	}

	// Test that RegisterAll doesn't return an error
	err := RegisterAll(s, config)
	if err != nil {
		t.Errorf("RegisterAll returned error: %v", err)
	}

	// TODO: In the future, we could test that specific tools like add_tags, delete_tags,
	// and rename_tag are registered when ReadWrite is true, but that would require
	// reflection or exposing the registered tools from the MCP server
}

func TestFilesToolBacklinks(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatalf("failed to create vault dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(vaultPath, "target.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to write target.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultPath, "ref.md"), []byte("Link to [[target]]"), 0o644); err != nil {
		t.Fatalf("failed to write ref.md: %v", err)
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

	tool := FilesTool(cfg)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "files",
			Arguments: map[string]interface{}{
				"inputs":           []interface{}{"target.md"},
				"includeBacklinks": true,
			},
		},
	}

	resp, err := tool(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	if !assert.Len(t, resp.Content, 1) {
		return
	}

	text, ok := resp.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", resp.Content[0])
	}

	var parsed FilesResponse
	err = json.Unmarshal([]byte(text.Text), &parsed)
	assert.NoError(t, err)
	assert.Len(t, parsed.Files, 1)
	assert.Equal(t, "target.md", parsed.Files[0].Path)
	if assert.NotNil(t, parsed.Files[0].Backlinks) {
		assert.Len(t, parsed.Files[0].Backlinks, 1)
		assert.Equal(t, "ref.md", parsed.Files[0].Backlinks[0].Referrer)
	}
}
