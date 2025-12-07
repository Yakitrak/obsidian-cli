package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/cache"
	"github.com/atomicobject/obsidian-cli/pkg/embeddings"
	"github.com/atomicobject/obsidian-cli/pkg/embeddings/sqlite"
	"github.com/atomicobject/obsidian-cli/pkg/mcp"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
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
- files: List files matching inputs (find:, tag:, literal) and optionally include content/frontmatter
- daily_note: JSON describing today's (or a given) daily note
- daily_note_path: Path info for a daily note
- list_tags: Tags with counts

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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
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

		// Create cache service but don't block on initial crawl.
		// The first tool call that needs vault data will trigger the crawl lazily.
		cacheService, err := cache.NewService(vaultPath, cache.Options{})
		if err != nil {
			log.Fatalf("Failed to initialize cache: %v", err)
		}
		defer cacheService.Close()

		analysisCache := cache.NewAnalysisCache(cacheService)

		// Configure MCP tools
		config := mcp.Config{
			Vault:          &vault,
			VaultPath:      vaultPath,
			Debug:          debug,
			SuppressedTags: []string{"no-prompt"}, // Default suppression
			ReadWrite:      mcpReadWrite,
			Cache:          cacheService,
			AnalysisCache:  analysisCache,
		}

		embCfg, err := obsidian.LoadEmbeddingsConfig(vaultPath)
		if err != nil {
			log.Printf("Failed to load embeddings config: %v", err)
		} else if embCfg.Enabled {
			apiKey := embeddings.ResolveAPIKey("")
			provider, err := embeddings.NewProvider(embCfg.ProviderCfg(apiKey))
			if err != nil {
				log.Printf("Embeddings provider unavailable; semantic features disabled: %v", err)
			} else {
				store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
				if err != nil {
					log.Printf("Embeddings index unavailable at %s: %v", embCfg.IndexPath, err)
				} else {
					defer store.Close()
					indexer := embeddings.NewIndexer(store, provider, embCfg.ProviderCfg(apiKey), vaultPath)
					if embCfg.BatchSize > 0 {
						indexer.BatchSize = embCfg.BatchSize
					}
					if embCfg.MaxConcurrency > 0 {
						indexer.MaxConcurrent = embCfg.MaxConcurrency
					}

					config.Embeddings = store
					config.EmbedProvider = provider
					config.EmbeddingsPath = embCfg.IndexPath
					config.EmbeddingsOn = true

					go func() {
						if err := indexer.SyncVault(ctx); err != nil {
							log.Printf("Embedding index refresh failed (semantic disabled): %v", err)
							return
						}
						watchEmbeddings(ctx, cacheService, indexer, vaultPath)
					}()
				}
			}
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

		// Add built-in agent guide resource.
		mcp.AddBuiltinResources(s)

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
• files – return matching files as JSON with optional content/frontmatter/backlinks. Inputs accept file paths, tag:foo, find:bar, and link-following flags.
• daily_note – JSON for today’s (or specified) daily note (path, exists, content).
• daily_note_path – JSON path/existence for a given date.
• list_tags – list tags with individual/aggregate counts.
• mutate_tags, mutate_properties – destructive tools (only available when server is started with --read-write). Each takes an op (add/delete/rename for tags; set/delete/rename for properties); delete/rename ops accept optional inputs to scope.

Best practices:
1. Prefer files when you need the actual text of notes; it supports recursion via links.
2. The server hides notes tagged no-prompt (and any tags in suppressTags) unless you disable suppression.
3. Use includeContent/includeFrontmatter/includeBacklinks flags to control payload size in files responses.

Additional resources are available under the URI prefix obsidian-cli/help/* (see list_resources).`
}

func watchEmbeddings(ctx context.Context, cacheService *cache.Service, indexer *embeddings.Indexer, vaultPath string) {
	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		lastFull := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				dirty := cacheService.DirtySnapshot()
				if err := cacheService.Refresh(ctx); err != nil {
					continue
				}
				now := time.Now()
				if len(dirty) == 0 && now.Sub(lastFull) < time.Hour {
					continue
				}

				// Periodic safety sweep to catch missed events or drifts.
				if now.Sub(lastFull) >= time.Hour {
					_ = indexer.SyncVault(ctx)
					lastFull = now
					continue
				}

				for rel, kind := range dirty {
					if filepath.Ext(rel) != ".md" {
						continue
					}
					switch kind {
					case cache.DirtyRemoved, cache.DirtyRenamed:
						_ = indexer.Index.DeleteNote(ctx, embeddings.NoteID(rel))
					default:
						if entry, ok := cacheService.Entry(rel); ok {
							info := embeddings.NoteFileInfo{
								ID:    embeddings.NoteID(entry.Path),
								Path:  filepath.Join(vaultPath, entry.Path),
								Title: filepath.Base(entry.Path),
								Size:  entry.Size,
								Mtime: entry.ModTime,
							}
							_ = indexer.UpdateNote(ctx, info, entry.Content)
						}
					}
				}
			}
		}
	}()
}
