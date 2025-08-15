package cmd

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmRemoveKey string

var frontmatterRemoveCmd = &cobra.Command{
	Use:   "remove <note>",
	Short: "Remove a YAML frontmatter key (and its value) from a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		if err := actions.RemoveFrontmatterKey(&v, note, fmRemoveKey); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	frontmatterRemoveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterRemoveCmd.Flags().StringVarP(&fmRemoveKey, "key", "k", "", "frontmatter key to remove")
	frontmatterRemoveCmd.MarkFlagRequired("key")
	frontmatterCmd.AddCommand(frontmatterRemoveCmd)
}
