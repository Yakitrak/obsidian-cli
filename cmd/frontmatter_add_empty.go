package cmd

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmAddEmptyKey string

var frontmatterAddEmptyCmd = &cobra.Command{
	Use:   "add-empty <note>",
	Short: "Add an empty YAML frontmatter key to a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		if err := actions.AddEmptyFrontmatterKey(&v, note, fmAddEmptyKey); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	frontmatterAddEmptyCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterAddEmptyCmd.Flags().StringVarP(&fmAddEmptyKey, "key", "k", "", "frontmatter key to add as empty")
	frontmatterAddEmptyCmd.MarkFlagRequired("key")
	frontmatterCmd.AddCommand(frontmatterAddEmptyCmd)
}
