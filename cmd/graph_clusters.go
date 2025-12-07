package cmd

import (
	"fmt"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	clustersSkipAnchors bool
	clustersSkipEmbeds  bool
)

var graphClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Show mutual-link clusters (strongly connected components)",
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
					SkipAnchors: clustersSkipAnchors,
					SkipEmbeds:  clustersSkipEmbeds,
				},
				MinDegree:         graphMinDegree,
				MutualOnly:        graphMutualOnly,
				RecencyCascade:    graphRecencyCascade,
				RecencyCascadeSet: true,
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

		fmt.Printf("Mutual-link clusters for vault %q (%s)\n", selectedVault, vaultPath)

		var clusters [][]string
		for _, component := range analysis.StrongComponents {
			if len(component) <= 1 {
				continue
			}
			clusters = append(clusters, component)
		}

		if len(clusters) == 0 {
			fmt.Println("  (none)")
			return nil
		}

		limit := graphLimit
		if graphShowAll || limit <= 0 || limit > len(clusters) {
			limit = len(clusters)
		}
		if !graphShowAll && limit < len(clusters) {
			fmt.Printf("Showing top %d of %d clusters:\n", limit, len(clusters))
		}

		for i := 0; i < limit; i++ {
			component := clusters[i]
			fmt.Printf("  size %d: %s\n", len(component), strings.Join(component, ", "))
		}
		if !graphShowAll && limit < len(clusters) {
			fmt.Printf("  ... (%d more)\n", len(clusters)-limit)
		}

		return nil
	},
}

func init() {
	graphClustersCmd.Flags().BoolVar(&clustersSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphClustersCmd.Flags().BoolVar(&clustersSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphClustersCmd)
}
