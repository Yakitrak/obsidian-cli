package mcp

import (
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/server"
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
