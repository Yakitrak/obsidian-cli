package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
)

var semanticCmd = &cobra.Command{
	Use:   "semantic",
	Short: "Semantic indexing and search over your vault",
}

var semanticIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build or refresh the semantic embeddings index",
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		embCfg, err := obsidian.LoadEmbeddingsConfig(vaultPath)
		if err != nil {
			return err
		}
		applySemanticOverrides(cmd, &embCfg)
		if embCfg.IndexPath == "" {
			embCfg.IndexPath = embeddings.DefaultIndexPath(vaultPath)
		}

		apiKey := embeddings.ResolveAPIKey(semanticAPIKey)
		provider, err := embeddings.NewProvider(embCfg.ProviderCfg(apiKey))
		if err != nil {
			return err
		}

		store, err := sqlite.Open(embCfg.IndexPath, provider.Dimensions())
		if err != nil {
			return err
		}
		defer store.Close()

		indexer := embeddings.NewIndexer(store, provider, vaultPath)
		if embCfg.BatchSize > 0 {
			indexer.BatchSize = embCfg.BatchSize
		}
		if embCfg.MaxConcurrency > 0 {
			indexer.MaxConcurrent = embCfg.MaxConcurrency
		}

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

		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		embCfg, err := obsidian.LoadEmbeddingsConfig(vaultPath)
		if err != nil {
			return err
		}
		if !embCfg.Enabled {
			return fmt.Errorf("semantic index not enabled for this vault; run `obsidian-cli semantic index` first")
		}
		applySemanticOverrides(cmd, &embCfg)
		if embCfg.IndexPath == "" {
			embCfg.IndexPath = embeddings.DefaultIndexPath(vaultPath)
		}
		if _, err := os.Stat(embCfg.IndexPath); err != nil {
			return fmt.Errorf("semantic index missing at %s (run semantic index)", embCfg.IndexPath)
		}

		apiKey := embeddings.ResolveAPIKey(semanticAPIKey)
		provider, err := embeddings.NewProvider(embCfg.ProviderCfg(apiKey))
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
		cands, err := store.SearchNotesByVector(cmd.Context(), vecs[0], limit)
		if err != nil {
			return err
		}

		if len(cands) == 0 {
			fmt.Println("No semantic matches found.")
			return nil
		}

		for _, c := range cands {
			fmt.Printf("%s\t%.4f\n", c.ID, c.Score)
		}
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

	semanticCmd.AddCommand(semanticIndexCmd)
	semanticCmd.AddCommand(semanticSearchCmd)
	semanticCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name (uses default if unset)")
	rootCmd.AddCommand(semanticCmd)
}
