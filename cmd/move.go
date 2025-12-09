package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var moveCmd = &cobra.Command{
	Use:     "move",
	Aliases: []string{"m"},
	Short:   "Move or rename note in vault and updated corresponding links",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		currentName := args[0]
		newName := args[1]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		useEditor, _ := cmd.Flags().GetBool("editor")
		params := actions.MoveParams{
			CurrentNoteName: currentName,
			NewNoteName:     newName,
			ShouldOpen:      shouldOpen,
			UseEditor:       useEditor,
		}
		err := actions.MoveNote(&vault, &note, &uri, params)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	moveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	moveCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	rootCmd.AddCommand(moveCmd)
}
