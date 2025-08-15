package cmd

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmClearKey string

var frontmatterClearCmd = &cobra.Command{
	Use:   "clear <note>",
	Short: "Clear the content of a YAML frontmatter key in a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		if err := actions.ClearFrontmatter(&v, note, fmClearKey); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	frontmatterClearCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterClearCmd.Flags().StringVarP(&fmClearKey, "key", "k", "", "frontmatter key to clear")
	frontmatterClearCmd.MarkFlagRequired("key")
	frontmatterCmd.AddCommand(frontmatterClearCmd)
}
