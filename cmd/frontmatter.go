package cmd

import (
	"errors"
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmPrint bool
var fmEdit bool
var fmDelete bool
var fmKey string
var fmValue string
var fmSelect bool

var frontmatterCmd = &cobra.Command{
	Use:     "frontmatter [note]",
	Aliases: []string{"fm"},
	Short:   "View or modify note frontmatter",
	Long: `View or modify YAML frontmatter in a note.

Use --print to display frontmatter, --edit to modify a key,
or --delete to remove a key.

Examples:
  obsidian-cli frontmatter "My Note" --print
  obsidian-cli frontmatter "My Note" --edit --key "status" --value "done"
  obsidian-cli frontmatter "My Note" --delete --key "draft"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		noteName := ""

		if len(args) > 0 && !fmSelect {
			noteName = args[0]
		} else {
			if !fmSelect {
				return errors.New("note path required (or use --ls)")
			}
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

		params := actions.FrontmatterParams{
			NoteName: noteName,
			Print:    fmPrint,
			Edit:     fmEdit,
			Delete:   fmDelete,
			Key:      fmKey,
			Value:    fmValue,
		}

		output, err := actions.Frontmatter(&vault, &note, params)
		if err != nil {
			return err
		}

		if output != "" {
			fmt.Print(output)
		}
		return nil
	},
}

func init() {
	frontmatterCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	frontmatterCmd.Flags().BoolVar(&fmSelect, "ls", false, "select a note interactively")
	frontmatterCmd.Flags().BoolVar(&fmSelect, "select", false, "select a note interactively")
	frontmatterCmd.Flags().BoolVarP(&fmPrint, "print", "p", false, "print frontmatter")
	frontmatterCmd.Flags().BoolVarP(&fmEdit, "edit", "e", false, "edit a frontmatter key")
	frontmatterCmd.Flags().BoolVarP(&fmDelete, "delete", "d", false, "delete a frontmatter key")
	frontmatterCmd.Flags().StringVarP(&fmKey, "key", "k", "", "key to edit or delete")
	frontmatterCmd.Flags().StringVar(&fmValue, "value", "", "value to set (required for --edit)")
	rootCmd.AddCommand(frontmatterCmd)
}
