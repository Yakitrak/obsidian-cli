package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils/uri"
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
		searchUri, err := pkg.SearchNotes(vaultName, searchText)
		if err != nil {
			log.Fatal(err)
		}
		err = uri.Execute(searchUri)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(searchCmd)
}
