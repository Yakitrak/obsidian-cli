package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"
	"github.com/spf13/cobra"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var createNoteCmd = &cobra.Command{
	Use:       "create",
	Aliases:   []string{"c"},
	Short:     "creates note in Obsidian",
	ValidArgs: []string{"vault", "text", "silent", "append", "overwrite"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteName := args[0]
		uri := pkg.CreateNote(vaultName, noteName, content, shouldAppend, shouldOverwrite)
		utils.UriExecute(uri)
	},
}

func init() {
	createNoteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	createNoteCmd.Flags().StringVarP(&content, "content", "c", "", "text to add to note")
	createNoteCmd.Flags().BoolVarP(&shouldAppend, "append", "a", false, "append to note")
	createNoteCmd.Flags().BoolVarP(&shouldOverwrite, "overwrite", "o", false, "overwrite note")
	createNoteCmd.MarkFlagsMutuallyExclusive("append", "overwrite")
	rootCmd.AddCommand(createNoteCmd)
}
