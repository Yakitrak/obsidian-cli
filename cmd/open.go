package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var vaultName string
var OpenVaultCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Opens note in obsidian",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultOp := obsidian.Vault{Name: vaultName}
		uriManager := obsidian.Uri{}
		noteName := args[0]
		params := actions.OpenParams{NoteName: noteName}
		err := actions.OpenNote(&vaultOp, &uriManager, params)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	OpenVaultCmd.Flags().StringVarP(&vaultName, "obsidian", "v", "", "obsidian name (not required if default is set)")
	rootCmd.AddCommand(OpenVaultCmd)
}
