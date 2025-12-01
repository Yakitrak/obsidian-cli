package cmd

import (
	"fmt"
	"log"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var graphIgnoreCmd = &cobra.Command{
	Use:   "ignore [patterns...]",
	Short: "Set graph ignore patterns for this vault (.obsidian-cli/config.json)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedVault := vaultName
		if selectedVault == "" {
			v := &obsidian.Vault{}
			name, err := v.DefaultName()
			if err != nil {
				return err
			}
			selectedVault = name
		}

		vault := &obsidian.Vault{Name: selectedVault}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		cfg, err := obsidian.LoadVaultGraphConfig(vaultPath)
		if err != nil {
			return err
		}
		cfg.GraphIgnore = args

		if err := obsidian.SaveVaultGraphConfig(vaultPath, cfg); err != nil {
			return err
		}

		fmt.Printf("Saved graph ignore patterns for vault %q (%s): %v\n", selectedVault, vaultPath, args)
		return nil
	},
}

func init() {
	graphIgnoreCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	graphCmd.AddCommand(graphIgnoreCmd)
	if err := graphIgnoreCmd.MarkFlagRequired("vault"); err != nil {
		log.Printf("MarkFlagRequired failed: %v", err)
	}
}
