package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var createNoteCmd = &cobra.Command{
	Use:     "create <note-path>",
	Aliases: []string{"c"},
	Short:   "Creates note in vault",
	Long: `Creates a new note in your Obsidian vault.

By default, if the note already exists, Obsidian will create a new note
with a numeric suffix. Use --append to add to an existing note, or
--overwrite to replace its contents.`,
	Example: `  # Create an empty note
  obsidian-cli create "New Note"

  # Create with content
  obsidian-cli create "Ideas" --content "My brilliant idea"

  # Append to existing note
  obsidian-cli create "Log" --content "Entry" --append

  # Create and open in Obsidian
  obsidian-cli create "Draft" --open

  # Create and open in $EDITOR
  obsidian-cli create "Draft" --open --editor`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		noteName := args[0]
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			return fmt.Errorf("failed to parse --editor flag: %w", err)
		}
		params := actions.CreateParams{
			NoteName:        noteName,
			Content:         content,
			ShouldAppend:    shouldAppend,
			ShouldOverwrite: shouldOverwrite,
			ShouldOpen:      shouldOpen,
			UseEditor:       useEditor,
		}
		return actions.CreateNote(&vault, &uri, params)
	},
}

func init() {
	createNoteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	createNoteCmd.Flags().BoolVarP(&shouldOpen, "open", "", false, "open created note")
	createNoteCmd.Flags().StringVarP(&content, "content", "c", "", "text to add to note")
	createNoteCmd.Flags().BoolVarP(&shouldAppend, "append", "a", false, "append to note")
	createNoteCmd.Flags().BoolVarP(&shouldOverwrite, "overwrite", "o", false, "overwrite note")
	createNoteCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	createNoteCmd.MarkFlagsMutuallyExclusive("append", "overwrite")
	rootCmd.AddCommand(createNoteCmd)
}
