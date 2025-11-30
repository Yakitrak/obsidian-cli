package cmd

import (
	"fmt"
	"sort"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	graphSkipAnchors bool
	graphSkipEmbeds  bool
)

var graphStatsCmd = &cobra.Command{
	Use:     "degrees",
	Aliases: []string{"stats"},
	Short:   "Show wikilink in/out degree counts for all notes",
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
			SkipAnchors: graphSkipAnchors,
			SkipEmbeds:  graphSkipEmbeds,
		})
		if err != nil {
			return err
		}

		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		fmt.Printf("Link degrees for vault %q (%s)\n", selectedVault, vaultPath)
		fmt.Printf("Notes scanned: %d\n\n", len(stats.Nodes))

		fmt.Println("Degree counts (outbound -> inbound):")
		paths := make([]string, 0, len(stats.Nodes))
		for path := range stats.Nodes {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		for _, path := range paths {
			node := stats.Nodes[path]
			fmt.Printf("  %s: out=%d in=%d\n", path, node.Outbound, node.Inbound)
		}

		return nil
	},
}

func init() {
	graphStatsCmd.Flags().BoolVar(&graphSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphStatsCmd.Flags().BoolVar(&graphSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphStatsCmd)
}
