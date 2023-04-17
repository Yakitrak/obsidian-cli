package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	temp2 "github.com/Yakitrak/obsidian-cli/temp/vaults"
	"github.com/Yakitrak/obsidian-cli/utils/temp"
	"github.com/spf13/cobra"
	"log"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var createNoteCmd = &cobra.Command{
	Use:       "create",
	Aliases:   []string{"c"},
	Short:     "Creates note in vault",
	ValidArgs: []string{"vault", "text", "silent", "append", "overwrite"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteName := args[0]
		uri := pkg.CreateNote(temp.UriConstructor, temp2.GetDefaultVault(vaultName), noteName, content, shouldAppend, shouldOverwrite)
		err := temp.UriExecute(uri)
		if err != nil {
			log.Fatal(err)
		}
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
