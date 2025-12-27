package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var searchContentCmd = &cobra.Command{
	Use:   "search-content <term>",
	Short: "Search note content for search term",
	Long: `Searches the contents of all notes for a term.

Displays matching notes with line numbers and snippets. If multiple
matches are found, opens a fuzzy finder to select which note to open.`,
	Example: `  # Search for a term
  obsidian-cli search-content "TODO"

  # Search and open in editor
  obsidian-cli search-content "bug" --editor

  # Search in specific vault
  obsidian-cli search-content "project" --vault "Work"`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"sc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		fuzzyFinder := obsidian.FuzzyFinder{}

		searchTerm := args[0]
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			return fmt.Errorf("failed to parse 'editor' flag: %w", err)
		}
		return actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, searchTerm, useEditor)
	},
}

func init() {
	searchContentCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchContentCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(searchContentCmd)
}
