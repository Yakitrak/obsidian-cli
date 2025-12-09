package cmd

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var DailyCmd = &cobra.Command{
	Use:     "daily",
	Aliases: []string{"d"},
	Short:   "Creates or opens daily note in vault",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		useEditor, _ := cmd.Flags().GetBool("editor")
		err := actions.DailyNote(&vault, &uri, useEditor)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	DailyCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	DailyCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(DailyCmd)
}
