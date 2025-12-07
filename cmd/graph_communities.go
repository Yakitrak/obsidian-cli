package cmd

import (
	"fmt"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	communitiesSkipAnchors bool
	communitiesSkipEmbeds  bool
)

var graphCommunitiesCmd = &cobra.Command{
	Use:   "communities",
	Short: "Show loosely connected communities (label propagation)",
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
					SkipAnchors: communitiesSkipAnchors,
					SkipEmbeds:  communitiesSkipEmbeds,
				},
				IncludeTags: true,
				MinDegree:   graphMinDegree,
				MutualOnly:  graphMutualOnly,
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

		communities := analysis.Communities
		fmt.Fprintf(cmd.OutOrStdout(), "Communities for vault %q (%s)\n", selectedVault, vaultPath)
		if len(communities) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  (none)")
			return nil
		}

		limit := graphLimit
		if graphShowAll || limit <= 0 || limit > len(communities) {
			limit = len(communities)
		}
		if !graphShowAll && limit < len(communities) {
			fmt.Fprintf(cmd.OutOrStdout(), "Showing top %d of %d communities:\n", limit, len(communities))
		}

		for i := 0; i < limit; i++ {
			c := communities[i]
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
			}
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStdout(), "  community %s (size %d)\n", colorCommunity(c.ID), len(c.Nodes))
			if c.Anchor != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "    anchor: %s\n", c.Anchor)
			}
			if c.Density > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "    density: %.3f\n", c.Density)
			}
			if c.Recency != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "    recency: %.1f days ago (%d in last %dd)\n", c.Recency.LatestAgeDays, c.Recency.RecentCount, c.Recency.WindowDays)
			}
			if len(c.TopTags) > 0 {
				var tags []string
				tagLimit := graphLimit
				if graphShowAll || tagLimit <= 0 || tagLimit > len(c.TopTags) {
					tagLimit = len(c.TopTags)
				}
				for j := 0; j < tagLimit; j++ {
					tt := c.TopTags[j]
					tags = append(tags, fmt.Sprintf("%s(%d)", tt.Tag, tt.Count))
				}
				if !graphShowAll && tagLimit < len(c.TopTags) {
					tags = append(tags, "...")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "    tags: %s\n", strings.Join(tags, ", "))
			}

			noteLimit := graphLimit
			if graphShowAll || noteLimit <= 0 || noteLimit > len(c.TopAuthority) {
				noteLimit = len(c.TopAuthority)
			}
			if noteLimit > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "    top notes (by authority):\n")
				for j := 0; j < noteLimit; j++ {
					p := c.TopAuthority[j]
					n := analysis.Nodes[p.Path]
					tagStr := ""
					if len(n.Tags) > 0 {
						tagStr = fmt.Sprintf(" tags:%s", strings.Join(n.Tags, ","))
					}
					fmt.Fprintf(cmd.OutOrStdout(), "      %d) %s auth=%.4f hub=%.4f in=%d out=%d%s\n", j+1, p.Path, p.Authority, p.Hub, n.Inbound, n.Outbound, tagStr)
				}
				if !graphShowAll && noteLimit < len(c.TopAuthority) {
					fmt.Fprintf(cmd.OutOrStdout(), "      ... (%d more)\n", len(c.TopAuthority)-noteLimit)
				}
			}
			if len(c.Bridges) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "    bridges: %s\n", strings.Join(c.Bridges, ", "))
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}

func colorCommunity(id string) string {
	if graphNoColor {
		return id
	}
	return "\033[36m" + id + "\033[0m"
}

func init() {
	graphCommunitiesCmd.Flags().BoolVar(&communitiesSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphCommunitiesCmd.Flags().BoolVar(&communitiesSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphCommunitiesCmd)
}
