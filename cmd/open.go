package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/Yakitrak/obsidian-cli/utils"
	"github.com/spf13/cobra"
)

var vaultName string
var openVaultCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Opens note in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteName := args[0]
		uri := pkg.OpenNote(utils.UriConstructor, utils.GetDefaultVault(vaultName), noteName)
		utils.UriExecute(uri)

	},
}

func init() {
	openVaultCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(openVaultCmd)
}
