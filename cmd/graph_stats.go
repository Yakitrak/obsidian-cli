package cmd

import (
	"fmt"
	"sort"
	"strings"

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

		analysis, err := actions.GraphAnalysis(&vault, &note, actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: graphSkipAnchors,
					SkipEmbeds:  graphSkipEmbeds,
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

		fmt.Fprintf(cmd.OutOrStdout(), "Graph for vault %q (%s)\n", selectedVault, vaultPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Nodes: %d  Edges: %d  Orphans: %d  Communities: %d\n\n", analysis.Stats.NodeCount, analysis.Stats.EdgeCount, len(analysis.Orphans), len(analysis.Communities))

		printTopNodes(cmd, analysis, graphLimit, graphShowAll)

		return nil
	},
}

func init() {
	graphStatsCmd.Flags().BoolVar(&graphSkipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	graphStatsCmd.Flags().BoolVar(&graphSkipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")

	graphCmd.AddCommand(graphStatsCmd)
}

func printTopNodes(cmd *cobra.Command, analysis *obsidian.GraphAnalysis, limit int, showAll bool) {
	if analysis == nil || len(analysis.Nodes) == 0 {
		return
	}

	type nodeRank struct {
		path      string
		title     string
		hub       float64
		authority float64
		inbound   int
		outbound  int
		community string
		tags      []string
	}

	var nodes []nodeRank
	for _, n := range analysis.Nodes {
		nodes = append(nodes, nodeRank{
			path:      n.Path,
			title:     n.Title,
			hub:       n.Hub,
			authority: n.Authority,
			inbound:   n.Inbound,
			outbound:  n.Outbound,
			community: n.Community,
			tags:      n.Tags,
		})
	}

	max := limit
	if showAll {
		max = len(nodes)
	}
	if max > len(nodes) {
		max = len(nodes)
	}

	// Top by Authority (cornerstone concepts)
	topAuthority := append([]nodeRank{}, nodes...)
	sort.Slice(topAuthority, func(i, j int) bool {
		if topAuthority[i].authority == topAuthority[j].authority {
			return topAuthority[i].path < topAuthority[j].path
		}
		return topAuthority[i].authority > topAuthority[j].authority
	})
	fmt.Fprintf(cmd.OutOrStdout(), "Top %d by authority (cornerstone concepts):\n", max)
	for i := 0; i < max; i++ {
		n := topAuthority[i]
		tagStr := ""
		if len(n.tags) > 0 {
			tagStr = fmt.Sprintf(" tags:%s", strings.Join(n.tags, ","))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s auth=%.4f hub=%.4f in=%d out=%d community=%s%s\n", i+1, n.path, n.authority, n.hub, n.inbound, n.outbound, n.community, tagStr)
	}
	if !showAll && len(topAuthority) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topAuthority)-max)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Top by Hub (index/MOC notes)
	topHub := append([]nodeRank{}, nodes...)
	sort.Slice(topHub, func(i, j int) bool {
		if topHub[i].hub == topHub[j].hub {
			return topHub[i].path < topHub[j].path
		}
		return topHub[i].hub > topHub[j].hub
	})
	fmt.Fprintf(cmd.OutOrStdout(), "Top %d by hub (index/MOC notes):\n", max)
	for i := 0; i < max; i++ {
		n := topHub[i]
		tagStr := ""
		if len(n.tags) > 0 {
			tagStr = fmt.Sprintf(" tags:%s", strings.Join(n.tags, ","))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s hub=%.4f auth=%.4f in=%d out=%d community=%s%s\n", i+1, n.path, n.hub, n.authority, n.inbound, n.outbound, n.community, tagStr)
	}
	if !showAll && len(topHub) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topHub)-max)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Top by inbound links
	topInbound := append([]nodeRank{}, nodes...)
	sort.Slice(topInbound, func(i, j int) bool {
		if topInbound[i].inbound == topInbound[j].inbound {
			return topInbound[i].path < topInbound[j].path
		}
		return topInbound[i].inbound > topInbound[j].inbound
	})
	fmt.Fprintf(cmd.OutOrStdout(), "Top %d by inbound links:\n", max)
	for i := 0; i < max; i++ {
		n := topInbound[i]
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s in=%d out=%d auth=%.4f hub=%.4f community=%s\n", i+1, n.path, n.inbound, n.outbound, n.authority, n.hub, n.community)
	}
	if !showAll && len(topInbound) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topInbound)-max)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Top by outbound links
	topOutbound := append([]nodeRank{}, nodes...)
	sort.Slice(topOutbound, func(i, j int) bool {
		if topOutbound[i].outbound == topOutbound[j].outbound {
			return topOutbound[i].path < topOutbound[j].path
		}
		return topOutbound[i].outbound > topOutbound[j].outbound
	})
	fmt.Fprintf(cmd.OutOrStdout(), "Top %d by outbound links:\n", max)
	for i := 0; i < max; i++ {
		n := topOutbound[i]
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s out=%d in=%d auth=%.4f hub=%.4f community=%s\n", i+1, n.path, n.outbound, n.inbound, n.authority, n.hub, n.community)
	}
	if !showAll && len(topOutbound) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topOutbound)-max)
	}
}
