package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var shouldRenderMarkdown bool
var printCmd = &cobra.Command{
	Use:     "print <note>",
	Aliases: []string{"p"},
	Short:   "Print contents of note",
	Long: `Prints the contents of a note to stdout.

Useful for piping note contents to other commands, or quickly viewing
a note without opening Obsidian.`,
	Example: `  # Print a note
  obsidian-cli print "Meeting Notes"

  # Print note in subfolder
  obsidian-cli print "Projects/readme"

  # Pipe to grep
  obsidian-cli print "Todo" | grep "TODO"

  # Copy to clipboard (macOS)
  obsidian-cli print "Template" | pbcopy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteName := args[0]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		params := actions.PrintParams{
			NoteName: noteName,
		}
		contents, err := actions.PrintNote(&vault, &note, params)
		if err != nil {
			return err
		}
		fmt.Println(contents)
		return nil
	},
}

func init() {
	printCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(printCmd)
}
