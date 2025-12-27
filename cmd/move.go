package cmd

import (
	"errors"
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var moveSelect bool
var moveCmd = &cobra.Command{
	Use:     "move [from-note-path] [to-note-path]",
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
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		currentName := ""
		newName := ""

		var vaultPath string
		needVaultPath := moveSelect || len(args) < 2
		if needVaultPath {
			if _, err := vault.DefaultName(); err != nil {
				return err
			}
			p, err := vault.Path()
			if err != nil {
				return err
			}
			vaultPath = p
		}

		switch len(args) {
		case 2:
			currentName = args[0]
			newName = args[1]
		case 1:
			if moveSelect {
				newName = args[0]
				selected, err := pickExistingNotePath(vaultPath)
				if err != nil {
					return err
				}
				currentName = selected
			} else {
				currentName = args[0]
				selected, err := promptNewNotePath(vaultPath)
				if err != nil {
					return err
				}
				newName = selected
			}
		case 0:
			if !moveSelect {
				return errors.New("from-note-path required (or use --ls)")
			}
			selected, err := pickExistingNotePath(vaultPath)
			if err != nil {
				return err
			}
			currentName = selected
			selected, err = promptNewNotePath(vaultPath)
			if err != nil {
				return err
			}
			newName = selected
		default:
			return errors.New("expected 0, 1, or 2 arguments")
		}

		if currentName == "" || newName == "" {
			return errors.New("both from-note-path and to-note-path are required")
		}
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
	moveCmd.Flags().BoolVar(&moveSelect, "ls", false, "select the note to move interactively")
	moveCmd.Flags().BoolVar(&moveSelect, "select", false, "select the note to move interactively")
	moveCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	rootCmd.AddCommand(moveCmd)
}
