package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/embeddings"
	"github.com/atomicobject/obsidian-cli/pkg/embeddings/sqlite"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	semanticProvider       string
	semanticModel          string
	semanticEndpoint       string
	semanticIndexPath      string
	semanticDimensions     int
	semanticBatchSize      int
	semanticMaxConcurrency int
	semanticAPIKey         string
	semanticLimit          int
	semanticChunks         bool
)

var semanticCmd = &cobra.Command{
	Use:   "semantic",
	Short: "Semantic indexing and search over your vault",
}

var semanticIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build or refresh the semantic embeddings index",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}

		provider, providerCfg, err := prepareProvider(embCfg)
		if err != nil {
			return err
		}

		store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
		if err != nil {
			return err
		}
		defer store.Close()

		indexer := configureIndexer(embCfg, provider, providerCfg, vaultPath, store)

		if err := indexer.SyncVault(cmd.Context()); err != nil {
			return err
		}

		embCfg.Enabled = true
		if err := obsidian.SaveEmbeddingsConfig(vaultPath, embCfg); err != nil {
			return err
		}
		if err := ensureIndexGitignored(vaultPath, embCfg.IndexPath); err != nil {
			return fmt.Errorf("wrote index but could not update gitignore: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Semantic index updated at %s using %s (%s)\n", embCfg.IndexPath, embCfg.Provider, embCfg.Model)
		return nil
	},
}

var semanticSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Semantic search indexed notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			return errors.New("query cannot be empty")
		}

		_, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}

		if !embCfg.Enabled {
			return fmt.Errorf("semantic index not enabled for this vault; run `obsidian-cli semantic index` first")
		}
		if _, err := os.Stat(embCfg.IndexPath); err != nil {
			return fmt.Errorf("semantic index missing at %s (run semantic index)", embCfg.IndexPath)
		}

		provider, _, err := prepareProvider(embCfg)
		if err != nil {
			return err
		}

		store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
		if err != nil {
			return err
		}
		defer store.Close()

		vecs, err := provider.EmbedTexts(cmd.Context(), []string{query})
		if err != nil {
			return err
		}
		if len(vecs) == 0 {
			return errors.New("provider returned no embeddings")
		}

		limit := semanticLimit
		if limit <= 0 {
			limit = 10
		}
		if semanticChunks {
			chunks, skipped, err := store.SearchChunksByVector(cmd.Context(), vecs[0], limit)
			if err != nil {
				return err
			}
			if len(chunks) == 0 {
				fmt.Println("No semantic matches found.")
				if skipped > 0 {
					fmt.Printf("(Skipped %d chunks due to dimension mismatch; consider rebuilding index)\n", skipped)
				}
				return nil
			}
			for _, c := range chunks {
				label := c.Heading
				if label == "" {
					label = c.Breadcrumb
				}
				fmt.Printf("%s\t%.4f\t[%d] %s\t%s\n", c.NoteID, c.Score, c.ChunkIndex, label, c.Breadcrumb)
			}
			if skipped > 0 {
				fmt.Printf("(Skipped %d chunks due to dimension mismatch; consider rebuilding index)\n", skipped)
			}
			return nil
		}

		cands, skipped, err := store.SearchNotesByVector(cmd.Context(), vecs[0], limit)
		if err != nil {
			return err
		}

		if len(cands) == 0 {
			fmt.Println("No semantic matches found.")
			if skipped > 0 {
				fmt.Printf("(Skipped %d chunks due to dimension mismatch; consider rebuilding index)\n", skipped)
			}
			return nil
		}

		for _, c := range cands {
			display := c.Title
			if c.Heading != "" {
				display = fmt.Sprintf("%s #%s", c.Title, c.Heading)
			}
			fmt.Printf("%s\t%.4f\t%s\n", c.ID, c.Score, display)
		}
		if skipped > 0 {
			fmt.Printf("(Skipped %d chunks due to dimension mismatch; consider rebuilding index)\n", skipped)
		}
		return nil
	},
}

var semanticStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show semantic index status and metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("Vault: %s\n", vaultPath)
		fmt.Printf("Enabled: %v\n", embCfg.Enabled)
		fmt.Printf("Index: %s\n", embCfg.IndexPath)
		fmt.Printf("Provider: %s (%s)\n", embCfg.Provider, embCfg.Model)
		if embCfg.Dimensions > 0 {
			fmt.Printf("Configured dimensions: %d\n", embCfg.Dimensions)
		}
		if embCfg.BatchSize > 0 || embCfg.MaxConcurrency > 0 {
			fmt.Printf("Batch size: %d  Max concurrent: %d\n", embCfg.BatchSize, embCfg.MaxConcurrency)
		}

		if !embCfg.Enabled {
			return nil
		}

		provider, _, perr := prepareProvider(embCfg)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "Warning: provider unavailable (%v); status may be incomplete\n", perr)
		}

		store, err := sqlite.Open(embCfg.IndexPath, embCfg.Dimensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Index unavailable at %s: %v\n", embCfg.IndexPath, err)
			return nil
		}
		defer store.Close()

		meta, ok, err := store.Metadata(cmd.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read index metadata: %v\n", err)
		}
		if ok {
			fmt.Printf("Index dimensions: %d\n", meta.Dimensions)
			fmt.Printf("Index provider/model: %s (%s)\n", meta.Provider, meta.Model)
			if !meta.LastSync.IsZero() {
				fmt.Printf("Last sync: %s\n", meta.LastSync.Format(time.RFC3339))
			}
			if provider != nil && meta.Dimensions > 0 && provider.Dimensions() > 0 && meta.Dimensions != provider.Dimensions() {
				fmt.Fprintf(os.Stderr, "Warning: provider dimensions (%d) differ from index dimensions (%d); rebuild recommended\n", provider.Dimensions(), meta.Dimensions)
			}
		} else {
			fmt.Println("Index metadata: not initialized")
		}

		if notes, chunks, err := store.Stats(cmd.Context()); err == nil {
			fmt.Printf("Indexed notes: %d  chunks: %d\n", notes, chunks)
		}
		return nil
	},
}

var semanticEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable semantic indexing for this vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}
		embCfg.Enabled = true
		if embCfg.IndexPath == "" {
			embCfg.IndexPath = embeddings.DefaultIndexPath(vaultPath)
		}
		if err := obsidian.SaveEmbeddingsConfig(vaultPath, embCfg); err != nil {
			return err
		}
		if err := ensureIndexGitignored(vaultPath, embCfg.IndexPath); err != nil {
			return fmt.Errorf("enabled but could not update gitignore: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Semantic indexing enabled. Run `obsidian-cli semantic index` to build or refresh the index.")
		return nil
	},
}

var semanticDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable semantic indexing and suppress related tooling",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}
		embCfg.Enabled = false
		if err := obsidian.SaveEmbeddingsConfig(vaultPath, embCfg); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Semantic indexing disabled. MCP semantic features and background watchers will be turned off.")
		return nil
	},
}

var semanticRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Incrementally refresh the semantic index (soft rebuild)",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}
		if !embCfg.Enabled {
			return fmt.Errorf("semantic indexing is disabled; run `obsidian-cli semantic enable` or `semantic index` first")
		}

		provider, providerCfg, err := prepareProvider(embCfg)
		if err != nil {
			return err
		}

		store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
		if err != nil {
			return err
		}
		defer store.Close()

		indexer := configureIndexer(embCfg, provider, providerCfg, vaultPath, store)
		if err := indexer.SyncVault(cmd.Context()); err != nil {
			return err
		}
		embCfg.Enabled = true
		if err := obsidian.SaveEmbeddingsConfig(vaultPath, embCfg); err != nil {
			return err
		}
		if err := ensureIndexGitignored(vaultPath, embCfg.IndexPath); err != nil {
			return fmt.Errorf("refreshed but could not update gitignore: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Semantic index refreshed at %s\n", embCfg.IndexPath)
		return nil
	},
}

var semanticRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Full rebuild of the semantic index (drops and recreates)",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, embCfg, err := loadVaultAndConfig(cmd)
		if err != nil {
			return err
		}

		provider, providerCfg, err := prepareProvider(embCfg)
		if err != nil {
			return err
		}

		if err := os.Remove(embCfg.IndexPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove old index: %w", err)
		}

		store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
		if err != nil {
			return err
		}
		defer store.Close()

		indexer := configureIndexer(embCfg, provider, providerCfg, vaultPath, store)
		if err := indexer.SyncVault(cmd.Context()); err != nil {
			return err
		}
		embCfg.Enabled = true
		if err := obsidian.SaveEmbeddingsConfig(vaultPath, embCfg); err != nil {
			return err
		}
		if err := ensureIndexGitignored(vaultPath, embCfg.IndexPath); err != nil {
			return fmt.Errorf("rebuilt but could not update gitignore: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Semantic index rebuilt at %s\n", embCfg.IndexPath)
		return nil
	},
}

func applySemanticOverrides(cmd *cobra.Command, cfg *embeddings.Config) {
	if cmd.Flags().Changed("provider") {
		cfg.Provider = semanticProvider
	}
	if cmd.Flags().Changed("model") {
		cfg.Model = semanticModel
	}
	if cmd.Flags().Changed("endpoint") {
		cfg.Endpoint = semanticEndpoint
	}
	if cmd.Flags().Changed("dimensions") && semanticDimensions > 0 {
		cfg.Dimensions = semanticDimensions
	}
	if cmd.Flags().Changed("index-path") {
		cfg.IndexPath = semanticIndexPath
	}
	if cmd.Flags().Changed("batch-size") && semanticBatchSize > 0 {
		cfg.BatchSize = semanticBatchSize
	}
	if cmd.Flags().Changed("max-concurrent") && semanticMaxConcurrency > 0 {
		cfg.MaxConcurrency = semanticMaxConcurrency
	}
}

func vaultPathOrDefault() (string, error) {
	vault := obsidian.Vault{Name: vaultName}
	return vault.Path()
}

func loadVaultAndConfig(cmd *cobra.Command) (string, embeddings.Config, error) {
	vaultPath, err := vaultPathOrDefault()
	if err != nil {
		return "", embeddings.Config{}, err
	}

	embCfg, err := obsidian.LoadEmbeddingsConfig(vaultPath)
	if err != nil {
		return "", embeddings.Config{}, err
	}
	applySemanticOverrides(cmd, &embCfg)
	if embCfg.IndexPath == "" {
		embCfg.IndexPath = embeddings.DefaultIndexPath(vaultPath)
	}
	return vaultPath, embCfg, nil
}

func prepareProvider(cfg embeddings.Config) (embeddings.Provider, embeddings.ProviderConfig, error) {
	apiKey := embeddings.ResolveAPIKey(semanticAPIKey)
	providerCfg := cfg.ProviderCfg(apiKey)
	provider, err := embeddings.NewProvider(providerCfg)
	if err != nil {
		return nil, embeddings.ProviderConfig{}, err
	}
	return provider, providerCfg, nil
}

func configureIndexer(embCfg embeddings.Config, provider embeddings.Provider, providerCfg embeddings.ProviderConfig, vaultPath string, store *sqlite.Store) *embeddings.Indexer {
	indexer := embeddings.NewIndexer(store, provider, providerCfg, vaultPath)
	if embCfg.BatchSize > 0 {
		indexer.BatchSize = embCfg.BatchSize
	}
	if embCfg.MaxConcurrency > 0 {
		indexer.MaxConcurrent = embCfg.MaxConcurrency
	}
	return indexer
}

func ensureIndexGitignored(vaultPath, indexPath string) error {
	rel, err := filepath.Rel(vaultPath, indexPath)
	if err != nil {
		rel = indexPath
	}
	entry := filepath.ToSlash(rel)
	gitignore := filepath.Join(vaultPath, ".obsidian-cli", ".gitignore")

	var lines []string
	if content, err := os.ReadFile(gitignore); err == nil {
		lines = strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	} else {
		lines = []string{}
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	lines = append(lines, entry)
	// Ensure trailing newline for cleanliness.
	if err := os.MkdirAll(filepath.Dir(gitignore), 0o755); err != nil {
		return err
	}
	return os.WriteFile(gitignore, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func init() {
	semanticCmd.PersistentFlags().StringVar(&semanticProvider, "provider", "", "Embedding provider (openai, ollama)")
	semanticCmd.PersistentFlags().StringVar(&semanticModel, "model", "", "Embedding model name")
	semanticCmd.PersistentFlags().StringVar(&semanticEndpoint, "endpoint", "", "Custom endpoint for provider")
	semanticCmd.PersistentFlags().IntVar(&semanticDimensions, "dimensions", 0, "Embedding dimensions (override provider default)")
	semanticCmd.PersistentFlags().StringVar(&semanticIndexPath, "index-path", "", "Custom path for the semantic index")
	semanticCmd.PersistentFlags().StringVar(&semanticAPIKey, "api-key", "", "API key for cloud providers (defaults to OBSIDIAN_CLI_OPENAI_API_KEY/OPENAI_API_KEY)")

	semanticIndexCmd.Flags().IntVar(&semanticBatchSize, "batch-size", 0, "Batch size for embedding requests")
	semanticIndexCmd.Flags().IntVar(&semanticMaxConcurrency, "max-concurrent", 0, "Maximum concurrent embedding batches")

	semanticSearchCmd.Flags().IntVarP(&semanticLimit, "limit", "n", 10, "Number of matches to return")
	semanticSearchCmd.Flags().BoolVar(&semanticChunks, "chunks", false, "Return chunk-level matches with headings/breadcrumbs")

	semanticCmd.AddCommand(semanticIndexCmd)
	semanticCmd.AddCommand(semanticSearchCmd)
	semanticCmd.AddCommand(semanticStatusCmd)
	semanticCmd.AddCommand(semanticEnableCmd)
	semanticCmd.AddCommand(semanticDisableCmd)
	semanticCmd.AddCommand(semanticRefreshCmd)
	semanticCmd.AddCommand(semanticRebuildCmd)
	semanticCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name (uses default if unset)")
	rootCmd.AddCommand(semanticCmd)
}
