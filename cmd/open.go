package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"github.com/spf13/cobra"
	"log"
)

var vaultName string
var OpenVaultCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Opens note in vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultOp := vault.Vault{Name: vaultName}
		noteName := args[0]
		createUri, err := actions.OpenNote(&vaultOp, noteName)
		if err != nil {
			log.Fatal(err)
		}
		err = uri.Execute(createUri)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	OpenVaultCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(OpenVaultCmd)
}
