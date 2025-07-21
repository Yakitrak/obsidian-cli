package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Yakitrak/obsidian-cli/pkg/mcp"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var (
	mcpPageSize         int
	mcpInstructionsFile string
	mcpReadWrite        bool
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run an MCP server exposing obsidian-cli tools",
	Long: `Run a Model Context Protocol (MCP) server that exposes obsidian-cli functionality as tools.
The server communicates over stdin/stdout and can be used with MCP clients like Claude Desktop, Cursor, or VS Code.

The server exposes read-only tools by default:
- list_files: List files matching inputs (find:, tag:, literal)
- file_info: Get information about a specific file
- print_note: Print the contents of a note
- daily_note_path: Get the path to today's daily note
- search_text: Search for text within notes
- list_tags: List all tags in the vault
- open_in_os: Open a file in the default OS application

Example MCP client configuration (e.g., for Claude Desktop):
{
  "mcpServers": {
    "obsidian-cli": {
      "command": "/path/to/obsidian-cli",
      "args": ["mcp", "--vault", "MyVault"],
      "env": {}
    }
  }
}`,
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug output if debug flag is set
		if debug {
			log.SetOutput(os.Stderr)
		}

		// If no vault name is provided, get the default vault name
		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				log.Fatal(err)
			}
			vaultName = defaultName
		}

		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		// Build server instructions
		var instructions string
		if mcpInstructionsFile != "" {
			b, err := os.ReadFile(mcpInstructionsFile)
			if err != nil {
				log.Fatalf("Failed to read instructions file: %v", err)
			}
			instructions = string(b)
		} else {
			instructions = defaultInstructionsString()
		}

		// Create MCP server
		s := server.NewMCPServer(
			"obsidian-cli",
			"v0.15.0",
			server.WithToolCapabilities(false),
			server.WithInstructions(instructions),
		)

		// Configure MCP tools
		config := mcp.Config{
			Vault:          &vault,
			VaultPath:      vaultPath,
			Debug:          debug,
			SuppressedTags: []string{"no-prompt"}, // Default suppression
			ReadWrite:      mcpReadWrite,
		}

		// Add any additional suppressed tags from global flags
		if len(suppressTags) > 0 {
			config.SuppressedTags = append(config.SuppressedTags, suppressTags...)
		}

		// Override with --no-suppress if set
		if noSuppress {
			config.SuppressedTags = []string{}
		}

		// Register all MCP tools
		err = mcp.RegisterAll(s, config)
		if err != nil {
			log.Fatalf("Failed to register MCP tools: %v", err)
		}

		// Load resources from .cursor/rules/*.mdc files (first mechanism)
		rulesDir := filepath.Join(vaultPath, ".cursor", "rules")
		resources, err := mcp.LoadMDCResources(rulesDir)
		if err == nil {
			for _, res := range resources {
				s.AddResource(res.Resource, res.Handler)
			}
		} else if debug {
			log.Printf("No MDC resources loaded: %v", err)
		}

		if debug {
			log.Printf("Starting MCP server for vault '%s' at %s", vaultName, vaultPath)
		}

		// Run the MCP server over stdio
		err = server.ServeStdio(s)
		if err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	},
}

func init() {
	mcpCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	mcpCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output")
	mcpCmd.Flags().StringSliceVar(&suppressTags, "suppress-tags", nil, "additional tags to suppress/exclude from output (comma-separated)")
	mcpCmd.Flags().BoolVar(&noSuppress, "no-suppress", false, "disable all tag suppression, including default no-prompt tag")
	mcpCmd.Flags().IntVar(&mcpPageSize, "page-size", 1000, "maximum number of items to return in a single page")
	mcpCmd.Flags().StringVar(&mcpInstructionsFile, "instructions-file", "", "path to file containing custom MCP server instructions")
	mcpCmd.Flags().BoolVar(&mcpReadWrite, "read-write", false, "enable destructive operations (hidden for future use)")
	mcpCmd.Flags().MarkHidden("read-write") // Hide this for now

	rootCmd.AddCommand(mcpCmd)
}

// defaultInstructionsString returns a concise help message for the MCP client.
func defaultInstructionsString() string {
	return `This MCP server exposes Obsidian-CLI tools for interacting with your vault.

Main tools:
• prompt_files – return full note contents formatted for LLMs. Inputs accept file paths, tag:foo, find:bar, and link-following flags.
• daily_note – contents of today’s (or specified) daily note.
• file_info – frontmatter, tags, word counts for a file.
• print_note – raw contents of a note file.
• search_text – find notes containing literal text.
• list_tags – list all unique tags.
• open_in_os – open a note in your default OS application.

Best practices:
1. Prefer prompt_files when you need the actual text of notes; it supports recursion via links.
2. The server hides notes tagged no-prompt (and any tags in suppressTags) unless you disable suppression.
3. Use file_info before loading large notes to decide if they’re relevant.

Additional resources are available under the URI prefix obsidian-cli/help/* (see list_resources).`
}
