package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	temp2 "github.com/Yakitrak/obsidian-cli/temp"
	"github.com/Yakitrak/obsidian-cli/utils/temp"
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

		uri, err := pkg.MoveNote(temp.UriConstructor, temp2.FindVaultPathFromConfig, temp2.MoveNote, temp2.UpdateLinksInVault, temp2.GetDefaultVault(vaultName), currentName, newName)
		if err != nil {
			log.Fatal(err)
		}
		if shouldOpen {
			err := temp.UriExecute(uri)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	rootCmd.AddCommand(moveCmd)
}
