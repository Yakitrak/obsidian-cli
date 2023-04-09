package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var moveCmd = &cobra.Command{
	Use:     "move",
	Aliases: []string{"m"},
	Short:   "Move or rename note in vault and updated corresponding links",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		current := args[0]
		new := args[1]
		uri := pkg.MoveNote(utils.UriConstructor, utils.FindVaultPathFromName, utils.MoveNote, utils.UpdateLinksInVault, utils.GetDefaultVault(vaultName), current, new)
		if shouldOpen {
			utils.UriExecute(uri)

		}
	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	rootCmd.AddCommand(moveCmd)
}
