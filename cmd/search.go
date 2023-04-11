package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"
	"log"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Short:   "Searches note in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchText := args[0]
		uri := pkg.SearchNotes(utils.UriConstructor, utils.GetDefaultVault(vaultName), searchText)
		err := utils.UriExecute(uri)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(searchCmd)
}
