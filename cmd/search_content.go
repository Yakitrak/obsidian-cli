package cmd

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var searchContentCmd = &cobra.Command{
	Use:     "search-content [search term]",
	Short:   "Search node content for search term",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"sc"},
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		fuzzyFinder := obsidian.FuzzyFinder{}

		searchTerm := args[0]
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			log.Fatalf("Failed to parse 'editor' flag: %v", err)
		}
		err = actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, searchTerm, useEditor)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	searchContentCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchContentCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(searchContentCmd)
}
