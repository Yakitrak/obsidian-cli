package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var renameCmd = &cobra.Command{
	Use:     "rename",
	Aliases: []string{"r"},
	Short:   "rename note in Obsidian and updated corresponding links",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		current := args[0]
		new := args[1]
		uri := pkg.RenameNote(vaultName, current, new)
		if shouldOpen {
			utils.UriExecute(uri)

		}
	},
}

func init() {
	renameCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	rootCmd.AddCommand(renameCmd)
}
