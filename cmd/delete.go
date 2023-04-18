package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"log"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete node in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		notePath := args[0]
		err := pkg.DeleteNote(vaultName, notePath)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	deleteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(deleteCmd)
}
