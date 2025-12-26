package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var deleteForce bool
var deleteDryRun bool

var deleteCmd = &cobra.Command{
	Use:     "delete <note-path>",
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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		notePath := args[0]
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
	rootCmd.AddCommand(deleteCmd)
}
