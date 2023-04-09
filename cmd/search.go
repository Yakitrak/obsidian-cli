package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Short:   "Searches note in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := pkg.SearchNotes(args[0], utils.GetDefaultVault(vaultName))
		utils.UriExecute(uri)
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(searchCmd)
}
