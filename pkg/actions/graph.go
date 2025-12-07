package actions

import (
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/cache"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// GraphStats returns link-graph degree counts and SCCs for the vault.
func GraphStats(vault obsidian.VaultManager, note obsidian.NoteManager, options obsidian.WikilinkOptions) (*obsidian.GraphStats, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	return obsidian.ComputeGraphStats(vaultPath, note, options)
}

// Orphans returns notes with zero inbound and outbound links.
func Orphans(vault obsidian.VaultManager, note obsidian.NoteManager, options obsidian.WikilinkOptions) ([]string, error) {
	stats, err := GraphStats(vault, note, options)
	if err != nil {
		return nil, err
	}
	return stats.Orphans(), nil
}

// GraphAnalysisParams control graph analysis inputs.
type GraphAnalysisParams struct {
	Options         obsidian.GraphAnalysisOptions
	ExcludePatterns []string
	IncludePatterns []string
	UseConfig       bool
	AnalysisCache   *cache.AnalysisCache
}

// GraphAnalysis returns a richer graph representation (hub/authority scores, communities, degrees, neighbors).
func GraphAnalysis(vault obsidian.VaultManager, note obsidian.NoteManager, params GraphAnalysisParams) (*obsidian.GraphAnalysis, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	excludePatterns := params.ExcludePatterns
	if params.UseConfig {
		if cfg, err := obsidian.LoadVaultGraphConfig(vaultPath); err == nil {
			excludePatterns = append(cfg.GraphIgnore, excludePatterns...)
		}
	}
	excludePatterns = expandPatterns(excludePatterns)
	includePatterns := expandPatterns(params.IncludePatterns)

	excludedSet := make(map[string]struct{})
	if len(excludePatterns) > 0 {
		parsed, expr, err := ParseInputsWithExpression(excludePatterns)
		if err != nil {
			return nil, err
		}
		matches, err := ListFiles(vault, note, ListParams{
			Inputs:      parsed,
			Expression:  expr,
			MaxDepth:    0,
			SkipAnchors: false,
			SkipEmbeds:  false,
		})
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			excludedSet[obsidian.NormalizePath(m)] = struct{}{}
		}
	}

	includedSet := make(map[string]struct{})
	if len(includePatterns) > 0 {
		parsed, expr, err := ParseInputsWithExpression(includePatterns)
		if err != nil {
			return nil, err
		}
		matches, err := ListFiles(vault, note, ListParams{
			Inputs:      parsed,
			Expression:  expr,
			MaxDepth:    0,
			SkipAnchors: false,
			SkipEmbeds:  false,
		})
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			includedSet[obsidian.NormalizePath(m)] = struct{}{}
		}
	}

	options := params.Options
	if !options.RecencyCascadeSet {
		options.RecencyCascade = true
		options.RecencyCascadeSet = true
	}
	options.ExcludedPaths = excludedSet
	options.IncludedPaths = includedSet

	if params.AnalysisCache != nil {
		return params.AnalysisCache.GraphAnalysis(vaultPath, note, options)
	}
	return obsidian.ComputeGraphAnalysis(vaultPath, note, options)
}

func expandPatterns(patterns []string) []string {
	var out []string
	for _, p := range patterns {
		fields := strings.Fields(p)
		if len(fields) == 0 {
			continue
		}
		out = append(out, fields...)
	}
	return out
}

// ExpandPatterns splits whitespace-separated pattern strings into individual patterns.
// For example, ["tag:foo tag:bar", "find:*"] becomes ["tag:foo", "tag:bar", "find:*"].
func ExpandPatterns(patterns []string) []string {
	return expandPatterns(patterns)
}
