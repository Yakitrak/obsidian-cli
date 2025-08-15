package cmd

import (
	"log"
	"strings"

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

		// Support adding multiple empty keys at once using comma-separated --key when --value is omitted
		if fmValue == "" && strings.Contains(fmKey, ",") {
			parts := strings.Split(fmKey, ",")
			for _, p := range parts {
				k := strings.TrimSpace(p)
				if k == "" {
					continue
				}
				params := actions.FrontmatterEditParams{
					NoteName: note,
					Key:      k,
					Value:    "",
				}
				if err := actions.EditFrontmatter(&v, params); err != nil {
					log.Fatal(err)
				}
			}
			return
		}

		// Default: single key edit (value may be empty or provided)
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
	if err := frontmatterEditCmd.MarkFlagRequired("key"); err != nil {
		log.Fatalf("failed to mark --key as required: %v", err)
	}
	frontmatterCmd.AddCommand(frontmatterEditCmd)
}
