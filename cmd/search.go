package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Short:   "Fuzzy searches and opens note in vault",
	Long: `Opens an interactive fuzzy finder to search and open notes.

Type to filter notes by filename. Press Enter to open the selected
note in Obsidian, or use --editor to open in your $EDITOR.`,
	Example: `  # Interactive search
  obsidian-cli search

  # Search and open in editor
  obsidian-cli search --editor

  # Search in specific vault
  obsidian-cli search --vault "Work"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		fuzzyFinder := obsidian.FuzzyFinder{}
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			return fmt.Errorf("failed to retrieve 'editor' flag: %w", err)
		}
		return actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, useEditor)
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(searchCmd)
}
