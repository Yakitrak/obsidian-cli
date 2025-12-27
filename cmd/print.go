package cmd

import (
	"errors"
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var shouldRenderMarkdown bool
var printSelect bool
var printCmd = &cobra.Command{
	Use:     "print [note-path]",
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
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		noteName := ""
		if len(args) > 0 && !printSelect {
			noteName = args[0]
		} else {
			if _, err := vault.DefaultName(); err != nil {
				return err
			}
			vaultPath, err := vault.Path()
			if err != nil {
				return err
			}
			selected, err := pickExistingNotePath(vaultPath)
			if err != nil {
				return err
			}
			noteName = selected
		}
		if noteName == "" {
			return errors.New("no note selected")
		}
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
	printCmd.Flags().BoolVar(&printSelect, "ls", false, "select a note interactively")
	printCmd.Flags().BoolVar(&printSelect, "select", false, "select a note interactively")
	rootCmd.AddCommand(printCmd)
}
