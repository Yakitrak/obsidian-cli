package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var moveCmd = &cobra.Command{
	Use:     "move <from-note-path> <to-note-path>",
	Aliases: []string{"m"},
	Short:   "Move or rename note in vault and update corresponding links",
	Long: `Moves or renames a note and updates all links pointing to it.

This command safely renames notes by also updating any [[wikilinks]]
or [markdown](links) that reference the moved note.`,
	Example: `  # Rename a note
  obsidian-cli move "Old Name" "New Name"

  # Move to a different folder
  obsidian-cli move "Inbox/note" "Projects/note"

  # Move and open the result
  obsidian-cli move "temp" "Archive/temp" --open`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		currentName := args[0]
		newName := args[1]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			return fmt.Errorf("failed to parse --editor flag: %w", err)
		}
		params := actions.MoveParams{
			CurrentNoteName: currentName,
			NewNoteName:     newName,
			ShouldOpen:      shouldOpen,
			UseEditor:       useEditor,
		}
		return actions.MoveNote(&vault, &note, &uri, params)
	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	moveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	moveCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	rootCmd.AddCommand(moveCmd)
}
