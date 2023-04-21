package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete node in obsidian",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultOp := obsidian.Vault{Name: vaultName}
		noteManager := note.Manager{}
		notePath := args[0]
		params := actions.DeleteParams{NotePath: notePath}
		err := actions.DeleteNote(&vaultOp, &noteManager, params)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	deleteCmd.Flags().StringVarP(&vaultName, "obsidian", "v", "", "obsidian name (not required if default is set)")
	rootCmd.AddCommand(deleteCmd)
}
