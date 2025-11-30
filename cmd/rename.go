package cmd

import (
	"fmt"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var renameNoBacklinks bool
var renameOverwrite bool

var renameCmd = &cobra.Command{
	Use:   "rename <source> <target>",
	Short: "Rename a note and update backlinks, preserving history in git vaults",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		params := actions.RenameParams{
			Source:          args[0],
			Target:          args[1],
			Overwrite:       renameOverwrite,
			UpdateBacklinks: !renameNoBacklinks,
		}

		result, err := actions.RenameNote(&vault, params)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Renamed to %s; link updates: %d; git history preserved: %t\n", result.RenamedPath, result.LinkUpdates, result.GitHistoryPreserved)
		if len(result.Skipped) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Skipped: %s\n", strings.Join(result.Skipped, ", "))
		}
		return nil
	},
}

func init() {
	renameCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	renameCmd.Flags().BoolVar(&renameOverwrite, "overwrite", false, "overwrite target note if it exists")
	renameCmd.Flags().BoolVar(&renameNoBacklinks, "no-backlinks", false, "skip rewriting backlinks to the renamed note")
	noteCmd.AddCommand(renameCmd)
}
