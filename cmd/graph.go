package cmd

import (
	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Link graph utilities (wikilinks)",
}

func init() {
	graphCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(graphCmd)
}
