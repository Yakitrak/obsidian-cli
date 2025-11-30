package actions

import "github.com/atomicobject/obsidian-cli/pkg/obsidian"

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
