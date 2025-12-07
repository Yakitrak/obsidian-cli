package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var graphCommunityIncludeNeighbors bool
var graphCommunityIncludeTags bool

var graphCommunityCmd = &cobra.Command{
	Use:   "community <id>",
	Short: "Show details for a specific community",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		communityArg := args[0]

		selectedVault := vaultName
		if selectedVault == "" {
			v := &obsidian.Vault{}
			name, err := v.DefaultName()
			if err != nil {
				return err
			}
			selectedVault = name
		}

		vault := obsidian.Vault{Name: selectedVault}
		note := obsidian.Note{}

		analysis, err := actions.GraphAnalysis(&vault, &note, actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: graphSkipAnchors,
					SkipEmbeds:  graphSkipEmbeds,
				},
				IncludeTags:       graphCommunityIncludeTags,
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

		var target *obsidian.CommunitySummary
		for i := range analysis.Communities {
			if analysis.Communities[i].ID == communityArg {
				target = &analysis.Communities[i]
				break
			}
		}

		if target == nil {
			normalized := obsidian.NormalizePath(obsidian.AddMdSuffix(communityArg))
			if filepath.IsAbs(communityArg) {
				vaultPath, err := vault.Path()
				if err != nil {
					return err
				}
				if rel, err := filepath.Rel(vaultPath, communityArg); err == nil {
					normalized = obsidian.NormalizePath(obsidian.AddMdSuffix(rel))
				}
			}

			if _, ok := analysis.Nodes[normalized]; !ok {
				return fmt.Errorf("community %s not found and file %s not in graph (use vault-relative paths, ensure it exists, and check include/exclude filters)", communityArg, communityArg)
			}

			lookup := obsidian.CommunityMembershipLookup(analysis.Communities)
			target = lookup[normalized]
			if target == nil {
				return fmt.Errorf("file %s is not assigned to a community under current filters", communityArg)
			}
		}

		if err := printCommunityDetail(cmd, target, analysis, selectedVault); err != nil {
			return err
		}
		if graphTimings {
			printTimings(cmd, analysis.Timings.ToMillis())
		}
		return nil
	},
}

func printCommunityDetail(cmd *cobra.Command, target *obsidian.CommunitySummary, analysis *obsidian.GraphAnalysis, selectedVault string) error {
	edgeCount := obsidian.CommunityInternalEdges(target, analysis.Nodes)

	fmt.Fprintf(cmd.OutOrStdout(), "Community %s (size %d) in vault %q\n", colorCommunity(target.ID), len(target.Nodes), selectedVault)
	if target.Anchor != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  anchor: %s\n", target.Anchor)
	}
	if target.Density > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  density: %.3f\n", target.Density)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  edges (internal): %d\n", edgeCount)
	if len(target.TopTags) > 0 {
		var tags []string
		for _, tt := range target.TopTags {
			tags = append(tags, fmt.Sprintf("%s(%d)", tt.Tag, tt.Count))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  tags: %s\n", strings.Join(tags, ", "))
	}
	if len(target.Bridges) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  bridges: %s\n", strings.Join(target.Bridges, ", "))
	}

	limit := graphLimit
	if graphShowAll || limit <= 0 || limit > len(target.Nodes) {
		limit = len(target.Nodes)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nMembers (sorted by authority):\n")
	members := rankCommunityMembers(target.Nodes, analysis)
	for i := 0; i < limit; i++ {
		m := members[i]
		tagStr := ""
		if len(m.tags) > 0 && graphCommunityIncludeTags {
			tagStr = fmt.Sprintf(" tags:%s", strings.Join(m.tags, ","))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s auth=%.4f hub=%.4f in=%d out=%d%s\n", i+1, m.path, m.authority, m.hub, m.in, m.out, tagStr)
		if graphCommunityIncludeNeighbors {
			fmt.Fprintf(cmd.OutOrStdout(), "      neighbors: %s\n", strings.Join(m.neighbors, ", "))
		}
	}
	if !graphShowAll && len(members) > limit {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(members)-limit)
	}

	return nil
}

func rankCommunityMembers(members []string, analysis *obsidian.GraphAnalysis) []struct {
	path      string
	hub       float64
	authority float64
	in        int
	out       int
	tags      []string
	neighbors []string
} {
	var ranked []struct {
		path      string
		hub       float64
		authority float64
		in        int
		out       int
		tags      []string
		neighbors []string
	}
	for _, p := range members {
		n := analysis.Nodes[p]
		ranked = append(ranked, struct {
			path      string
			hub       float64
			authority float64
			in        int
			out       int
			tags      []string
			neighbors []string
		}{
			path:      p,
			hub:       n.Hub,
			authority: n.Authority,
			in:        n.Inbound,
			out:       n.Outbound,
			tags:      n.Tags,
			neighbors: n.Neighbors,
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].authority == ranked[j].authority {
			return ranked[i].path < ranked[j].path
		}
		return ranked[i].authority > ranked[j].authority
	})
	return ranked
}

func init() {
	graphCommunityCmd.Flags().BoolVar(&graphCommunityIncludeNeighbors, "neighbors", false, "show neighbor lists for members")
	graphCommunityCmd.Flags().BoolVar(&graphCommunityIncludeTags, "tags", false, "show tags for members")
	graphCommunityCmd.Flags().BoolVar(&graphSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphCommunityCmd.Flags().BoolVar(&graphSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphCommunityCmd)
}
