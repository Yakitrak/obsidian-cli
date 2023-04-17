package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	temp2 "github.com/Yakitrak/obsidian-cli/temp/vaults"
	"github.com/Yakitrak/obsidian-cli/utils/temp"
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
		uri := pkg.SearchNotes(temp.UriConstructor, temp2.GetDefaultVault(vaultName), searchText)
		err := temp.UriExecute(uri)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(searchCmd)
}
