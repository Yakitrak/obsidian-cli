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

		stats, err := actions.GraphStats(&vault, &note, obsidian.WikilinkOptions{
			SkipAnchors: clustersSkipAnchors,
			SkipEmbeds:  clustersSkipEmbeds,
		})
		if err != nil {
			return err
		}

		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		fmt.Printf("Mutual-link clusters for vault %q (%s)\n", selectedVault, vaultPath)

		found := false
		for _, component := range stats.Components {
			if len(component) <= 1 {
				continue
			}
			found = true
			fmt.Printf("  size %d: %s\n", len(component), strings.Join(component, ", "))
		}
		if !found {
			fmt.Println("  (none)")
		}

		return nil
	},
}

func init() {
	graphClustersCmd.Flags().BoolVar(&clustersSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphClustersCmd.Flags().BoolVar(&clustersSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphClustersCmd)
}
