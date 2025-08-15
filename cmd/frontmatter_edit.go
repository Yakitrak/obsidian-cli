package cmd

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmKey string
var fmValue string

var frontmatterEditCmd = &cobra.Command{
	Use:   "edit <note>",
	Short: "Edit YAML frontmatter key in a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		params := actions.FrontmatterEditParams{
			NoteName: note,
			Key:      fmKey,
			Value:    fmValue,
		}
		if err := actions.EditFrontmatter(&v, params); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	frontmatterEditCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterEditCmd.Flags().StringVarP(&fmKey, "key", "k", "", "frontmatter key to set")
	frontmatterEditCmd.Flags().StringVarP(&fmValue, "value", "V", "", "frontmatter value (YAML)")
	frontmatterEditCmd.MarkFlagRequired("key")
	frontmatterEditCmd.MarkFlagRequired("value")
	frontmatterCmd.AddCommand(frontmatterEditCmd)
}
