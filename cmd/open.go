package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	temp2 "github.com/Yakitrak/obsidian-cli/temp"
	"github.com/Yakitrak/obsidian-cli/utils/temp"
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
		noteName := args[0]
		uri := pkg.OpenNote(temp.UriConstructor, temp2.GetDefaultVault(vaultName), noteName)
		err := temp.UriExecute(uri)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	OpenVaultCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(OpenVaultCmd)
}
