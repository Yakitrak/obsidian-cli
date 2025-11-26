package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var shouldOpen bool
var moveOverwrite bool
var moveUpdateBacklinks bool
var moveToFolder string

var moveCmd = &cobra.Command{
	Use:     "move <source> <target> | move --to-folder <folder> <sources...>",
	Aliases: []string{"m"},
	Short:   "Move or rename notes within the vault; backlinks are skipped by default",
	Args: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(moveToFolder) != "" {
			if len(args) < 1 {
				return fmt.Errorf("at least one source is required when using --to-folder")
			}
			return nil
		}
		if len(args) != 2 {
			return fmt.Errorf("move requires source and target (or use --to-folder)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}

		var moves []actions.MoveRequest
		if strings.TrimSpace(moveToFolder) != "" {
			for _, src := range args {
				base := filepath.Base(src)
				moves = append(moves, actions.MoveRequest{
					Source: src,
					Target: filepath.Join(moveToFolder, base),
				})
			}
		} else {
			moves = append(moves, actions.MoveRequest{Source: args[0], Target: args[1]})
		}

		summary, err := actions.MoveNotes(&vault, &uri, actions.MoveParams{
			Moves:           moves,
			Overwrite:       moveOverwrite,
			UpdateBacklinks: moveUpdateBacklinks,
			ShouldOpen:      shouldOpen,
		})
		if err != nil {
			return err
		}

		for _, res := range summary.Results {
			fmt.Fprintf(cmd.OutOrStdout(), "Moved %s -> %s (git history preserved: %t, link updates: %d)\n", res.Source, res.Target, res.GitHistoryPreserved, res.LinkUpdates)
		}
		if len(summary.Skipped) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Skipped: %s\n", strings.Join(summary.Skipped, ", "))
		}
		if summary.TotalLinkUpdates > 0 && len(summary.Results) > 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "Total link updates: %d\n", summary.TotalLinkUpdates)
		}
		return nil
	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	moveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	moveCmd.Flags().BoolVar(&moveOverwrite, "overwrite", false, "overwrite target note if it exists")
	moveCmd.Flags().BoolVar(&moveUpdateBacklinks, "update-backlinks", false, "rewrite backlinks to point to the moved note (default: false)")
	moveCmd.Flags().StringVar(&moveToFolder, "to-folder", "", "move one or more notes into the specified folder (preserves filenames)")
	rootCmd.AddCommand(moveCmd)
}
