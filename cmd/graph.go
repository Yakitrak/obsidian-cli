package cmd

import (
	"fmt"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Link graph utilities (wikilinks)",
}

var graphExcludePatterns []string
var graphIncludePatterns []string
var graphLimit int
var graphShowAll bool
var graphNoColor bool
var graphMinDegree int
var graphMutualOnly bool
var graphRecencyCascade bool
var graphTimings bool

func init() {
	graphCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	graphCmd.PersistentFlags().IntVar(&graphLimit, "limit", 100, "max items to show in summaries (authority/hub, communities, clusters)")
	graphCmd.PersistentFlags().BoolVar(&graphShowAll, "all", false, "show full listings instead of summaries")
	graphCmd.PersistentFlags().BoolVar(&graphNoColor, "no-color", false, "disable colored graph output")
	graphCmd.PersistentFlags().StringSliceVar(&graphExcludePatterns, "exclude", nil, "exclude notes matching these patterns (same syntax as list/prompt)")
	graphCmd.PersistentFlags().StringSliceVar(&graphIncludePatterns, "include", nil, "include only notes matching these patterns (same syntax as list/prompt)")
	graphCmd.PersistentFlags().IntVar(&graphMinDegree, "min-degree", 2, "drop notes whose in+out degree is below this number before analysis (0 = no filter)")
	graphCmd.PersistentFlags().BoolVar(&graphMutualOnly, "mutual-only", false, "only consider mutual (bidirectional) links when building the graph")
	graphCmd.PersistentFlags().BoolVar(&graphRecencyCascade, "recency-cascade", true, "cascade inferred recency beyond 1 hop (disable for legacy single-hop)")
	graphCmd.PersistentFlags().BoolVar(&graphTimings, "timings", false, "print graph analysis timings (load/build/hits/label/recency/total)")
	rootCmd.AddCommand(graphCmd)
}

func printTimings(cmd *cobra.Command, t interface{}) {
	fmt.Fprintf(cmd.OutOrStdout(), "Timings:\n")
	switch v := t.(type) {
	case obsidian.GraphTimingsMillis:
		fmt.Fprintf(cmd.OutOrStdout(), "  load:    %d ms\n", v.LoadEntries)
		fmt.Fprintf(cmd.OutOrStdout(), "  build:   %d ms\n", v.BuildGraph)
		fmt.Fprintf(cmd.OutOrStdout(), "  hits:    %d ms\n", v.HITS)
		fmt.Fprintf(cmd.OutOrStdout(), "  label:   %d ms\n", v.LabelProp)
		fmt.Fprintf(cmd.OutOrStdout(), "  recency: %d ms\n", v.Recency)
		fmt.Fprintf(cmd.OutOrStdout(), "  total:   %d ms\n", v.Total)
	case obsidian.GraphTimings:
		fmt.Fprintf(cmd.OutOrStdout(), "  load:    %s\n", v.LoadEntries)
		fmt.Fprintf(cmd.OutOrStdout(), "  build:   %s\n", v.BuildGraph)
		fmt.Fprintf(cmd.OutOrStdout(), "  hits:    %s\n", v.HITS)
		fmt.Fprintf(cmd.OutOrStdout(), "  label:   %s\n", v.LabelProp)
		fmt.Fprintf(cmd.OutOrStdout(), "  recency: %s\n", v.Recency)
		fmt.Fprintf(cmd.OutOrStdout(), "  total:   %s\n", v.Total)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "  (unknown timings format)\n")
	}
}
