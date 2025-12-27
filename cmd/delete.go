package cmd

import (
	"errors"
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var deleteForce bool
var deleteDryRun bool
var deleteSelect bool

var deleteCmd = &cobra.Command{
	Use:     "delete [note-path]",
	Aliases: []string{"del"},
	Short:   "Delete note in vault",
	Long: `Delete a note from the vault.

If other notes link to the note, you'll be prompted to confirm.
Use --force to skip confirmation (recommended for scripts).`,
	Example: `  # Delete a note (prompts if linked)
  obsidian-cli delete "old-note"

  # Force delete without prompt
  obsidian-cli delete "temp" --force

  # Delete from specific vault
  obsidian-cli delete "note" --vault "Archive"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		notePath := ""

		if len(args) > 0 && !deleteSelect {
			notePath = args[0]
		} else {
			if !deleteSelect {
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
			notePath = selected
		}
		if notePath == "" {
			return errors.New("no note selected")
		}
		params := actions.DeleteParams{
			NotePath: notePath,
			Force:    deleteForce,
		}

		if deleteDryRun {
			plan, err := actions.PlanDeleteNote(&vault, params)
			if err != nil {
				return err
			}
			fmt.Println("Delete dry-run:")
			fmt.Printf("  vault: %s\n", plan.VaultName)
			fmt.Printf("  path: %s\n", plan.AbsoluteNotePath)
			return nil
		}
		return actions.DeleteNote(&vault, &note, params)
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "preview which file would be deleted without deleting it")
	deleteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation if the note has incoming links")
	deleteCmd.Flags().BoolVar(&deleteSelect, "ls", false, "select a note interactively")
	deleteCmd.Flags().BoolVar(&deleteSelect, "select", false, "select a note interactively")
	rootCmd.AddCommand(deleteCmd)
}
