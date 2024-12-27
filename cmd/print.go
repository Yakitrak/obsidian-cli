package cmd

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var shouldRenderMarkdown bool
var printCmd = &cobra.Command{
	Use:     "print",
	Aliases: []string{"p"},
	Short:   "Print contents of note",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteName := args[0]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		params := actions.PrintParams{
			NoteName: noteName,
		}
		contents, err := actions.PrintNote(&vault, &note, params)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(contents)
	},
}

func init() {
	printCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(printCmd)
}
