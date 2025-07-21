package mcp

import "github.com/Yakitrak/obsidian-cli/pkg/obsidian"

// Config holds configuration for MCP tools
type Config struct {
	Vault          *obsidian.Vault
	VaultPath      string
	Debug          bool
	SuppressedTags []string
	ReadWrite      bool
}
