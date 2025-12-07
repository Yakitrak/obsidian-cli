package cmd

import (
	"fmt"
	"sort"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	orphansSkipAnchors bool
	orphansSkipEmbeds  bool
)

var orphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "List notes with no inbound or outbound wikilinks",
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedVault := vaultName
		if selectedVault == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				return err
			}
			selectedVault = defaultName
		}

		vault := obsidian.Vault{Name: selectedVault}
		note := obsidian.Note{}

		analysis, err := actions.GraphAnalysis(&vault, &note, actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: orphansSkipAnchors,
					SkipEmbeds:  orphansSkipEmbeds,
				},
				MinDegree:      graphMinDegree,
				MutualOnly:     graphMutualOnly,
				RecencyCascade: graphRecencyCascade,
			},
			ExcludePatterns: graphExcludePatterns,
			IncludePatterns: graphIncludePatterns,
		})
		if err != nil {
			return err
		}

		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		fmt.Printf("Orphans (no inbound or outbound wikilinks) in %q (%s):\n", selectedVault, vaultPath)
		if len(analysis.Orphans) == 0 {
			fmt.Println("  (none)")
			return nil
		}

		sorted := make([]string, len(analysis.Orphans))
		copy(sorted, analysis.Orphans)
		sort.Strings(sorted)

		for _, path := range sorted {
			fmt.Printf("  %s\n", path)
		}
		return nil
	},
}

func init() {
	orphansCmd.Flags().BoolVar(&orphansSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	orphansCmd.Flags().BoolVar(&orphansSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(orphansCmd)
}
