package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Short:   "Fuzzy searches and opens note in vault",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		fuzzyFinder := obsidian.FuzzyFinder{}
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			log.Fatalf("failed to retrieve 'editor' flag: %v", err)
		}
		err = actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, useEditor)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(searchCmd)
}
