package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var dailyDryRun bool

var DailyCmd = &cobra.Command{
	Use:     "daily",
	Aliases: []string{"d"},
	Short:   "Creates or opens daily note in vault",
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}

		if dailyDryRun {
			u, err := actions.PlanDailyNote(&vault, &uri)
			if err != nil {
				return err
			}
			fmt.Println("Daily dry-run:")
			fmt.Println(u)
			return nil
		}

		return actions.DailyNote(&vault, &uri)
	},
}

func init() {
	DailyCmd.Flags().BoolVar(&dailyDryRun, "dry-run", false, "print the Obsidian URI without opening it")
	DailyCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(DailyCmd)
}
