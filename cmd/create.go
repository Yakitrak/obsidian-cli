package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var createNoteCmd = &cobra.Command{
	Use:       "create",
	Aliases:   []string{"c"},
	Short:     "Creates note in obsidian",
	ValidArgs: []string{"obsidian", "text", "silent", "append", "overwrite"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultOp := obsidian.Vault{Name: vaultName}
		noteName := args[0]
		uriManager := obsidian.Uri{}
		params := actions.CreateParams{
			NoteName:        noteName,
			Content:         content,
			ShouldAppend:    shouldAppend,
			ShouldOverwrite: shouldOverwrite,
		}
		err := actions.CreateNote(&vaultOp, &uriManager, params)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	createNoteCmd.Flags().StringVarP(&vaultName, "obsidian", "v", "", "obsidian name")
	createNoteCmd.Flags().StringVarP(&content, "content", "c", "", "text to add to note")
	createNoteCmd.Flags().BoolVarP(&shouldAppend, "append", "a", false, "append to note")
	createNoteCmd.Flags().BoolVarP(&shouldOverwrite, "overwrite", "o", false, "overwrite note")
	createNoteCmd.MarkFlagsMutuallyExclusive("append", "overwrite")
	rootCmd.AddCommand(createNoteCmd)
}
