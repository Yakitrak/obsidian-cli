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
				MinDegree:  graphMinDegree,
				MutualOnly: graphMutualOnly,
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
		pagerank  float64
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
			pagerank:  n.Pagerank,
			inbound:   n.Inbound,
			outbound:  n.Outbound,
			community: n.Community,
			tags:      n.Tags,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].pagerank == nodes[j].pagerank {
			return nodes[i].path < nodes[j].path
		}
		return nodes[i].pagerank > nodes[j].pagerank
	})

	max := limit
	if showAll {
		max = len(nodes)
	}
	if max > len(nodes) {
		max = len(nodes)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Top %d by pagerank:\n", max)
	for i := 0; i < max; i++ {
		n := nodes[i]
		tagStr := ""
		if len(n.tags) > 0 {
			tagStr = fmt.Sprintf(" tags:%s", strings.Join(n.tags, ","))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s pr=%.4f in=%d out=%d community=%s%s\n", i+1, n.path, n.pagerank, n.inbound, n.outbound, n.community, tagStr)
	}
	if !showAll && len(nodes) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(nodes)-max)
	}
	fmt.Fprintln(cmd.OutOrStdout())

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
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s in=%d out=%d pr=%.4f community=%s\n", i+1, n.path, n.inbound, n.outbound, n.pagerank, n.community)
	}
	if !showAll && len(topInbound) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topInbound)-max)
	}
	fmt.Fprintln(cmd.OutOrStdout())

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
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s out=%d in=%d pr=%.4f community=%s\n", i+1, n.path, n.outbound, n.inbound, n.pagerank, n.community)
	}
	if !showAll && len(topOutbound) > max {
		fmt.Fprintf(cmd.OutOrStdout(), "  ... (%d more)\n", len(topOutbound)-max)
	}
}
