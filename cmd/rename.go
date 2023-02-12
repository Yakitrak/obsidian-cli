package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg"

	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:     "rename",
	Aliases: []string{"r"},
	Short:   "rename file in Obsidian and updated corresponding links",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		uri := pkg.RenameNote(vaultName, args[0], args[1])
		fmt.Println(uri)
		// utils.UriExecute(uri)
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
