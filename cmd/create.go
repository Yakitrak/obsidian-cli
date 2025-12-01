package cmd

import (
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var createNoteCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "Creates note in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		noteName := args[0]
		params := actions.CreateParams{
			NoteName:        noteName,
			Content:         content,
			ShouldAppend:    shouldAppend,
			ShouldOverwrite: shouldOverwrite,
			ShouldOpen:      shouldOpen,
		}
		err := actions.CreateNote(&vault, &uri, params)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	createNoteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	createNoteCmd.Flags().BoolVarP(&shouldOpen, "open", "", false, "open created note")
	createNoteCmd.Flags().StringVarP(&content, "content", "c", "", "text to add to note")
	createNoteCmd.Flags().BoolVarP(&shouldAppend, "append", "a", false, "append to note")
	createNoteCmd.Flags().BoolVarP(&shouldOverwrite, "overwrite", "o", false, "overwrite note")
	createNoteCmd.MarkFlagsMutuallyExclusive("append", "overwrite")
	noteCmd.AddCommand(createNoteCmd)
}
