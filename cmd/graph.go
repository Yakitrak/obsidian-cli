package cmd

import (
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

func init() {
	graphCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	graphCmd.PersistentFlags().IntVar(&graphLimit, "limit", 25, "max items to show in summaries (authority/hub, communities, clusters)")
	graphCmd.PersistentFlags().BoolVar(&graphShowAll, "all", false, "show full listings instead of summaries")
	graphCmd.PersistentFlags().BoolVar(&graphNoColor, "no-color", false, "disable colored graph output")
	graphCmd.PersistentFlags().StringSliceVar(&graphExcludePatterns, "exclude", nil, "exclude notes matching these patterns (same syntax as list/prompt)")
	graphCmd.PersistentFlags().StringSliceVar(&graphIncludePatterns, "include", nil, "include only notes matching these patterns (same syntax as list/prompt)")
	graphCmd.PersistentFlags().IntVar(&graphMinDegree, "min-degree", 2, "drop notes whose in+out degree is below this number before analysis (0 = no filter)")
	graphCmd.PersistentFlags().BoolVar(&graphMutualOnly, "mutual-only", false, "only consider mutual (bidirectional) links when building the graph")
	rootCmd.AddCommand(graphCmd)
}
