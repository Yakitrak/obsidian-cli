package cmd

import (
	"fmt"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var installIgnoreCmd = &cobra.Command{
	Use:   "install-ignore",
	Short: "Write the default .obsidianignore into the vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				return fmt.Errorf("failed to get default vault name: %w", err)
			}
			vaultName = defaultName
		}

		vault := &obsidian.Vault{Name: vaultName}
		path, err := actions.InstallDefaultIgnore(vault, actions.InstallIgnoreOptions{Force: force})
		if err != nil {
			return err
		}

		fmt.Printf("Wrote default ignore file to %s\n", path)
		return nil
	},
}

func init() {
	installIgnoreCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	installIgnoreCmd.Flags().Bool("force", false, "Overwrite an existing .obsidianignore")
	vaultCmd.AddCommand(installIgnoreCmd)
}
