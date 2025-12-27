package cmd

import (
	"errors"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var vaultName string
var openSelect bool
var OpenVaultCmd = &cobra.Command{
	Use:     "open [note-path]",
	Aliases: []string{"o"},
	Short:   "Opens note in vault by note name",
	Long: `Opens a note in Obsidian by name or path.

The note name can be just the filename or a path relative to the vault root.
The .md extension is optional.`,
	Example: `  # Open a note by name
  obsidian-cli open "Meeting Notes"

  # Open a note in a subfolder
  obsidian-cli open "Projects/my-project"

  # Open in a specific vault
  obsidian-cli open "Daily" --vault "Work"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		noteName := ""

		if len(args) > 0 && !openSelect {
			noteName = args[0]
		} else {
			if _, err := vault.DefaultName(); err != nil {
				return err
			}
			vaultPath, err := vault.Path()
			if err != nil {
				return err
			}
			selected, err := pickNotePathOrNew(vaultPath)
			if err != nil {
				return err
			}
			noteName = selected
		}

		if noteName == "" {
			return errors.New("no note selected")
		}
		params := actions.OpenParams{NoteName: noteName}
		return actions.OpenNote(&vault, &uri, params)
	},
}

func init() {
	OpenVaultCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	OpenVaultCmd.Flags().BoolVar(&openSelect, "ls", false, "select a note interactively")
	OpenVaultCmd.Flags().BoolVar(&openSelect, "select", false, "select a note interactively")
	rootCmd.AddCommand(OpenVaultCmd)
}
