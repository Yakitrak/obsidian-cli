package cmd

import (
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var removeVaultCmd = &cobra.Command{
	Use: "remove [name]",
	Aliases: []string{
		"remove-vault",
	},
	Short: "Remove a vault path from obsidian-cli preferences",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		vault := obsidian.Vault{Name: name}
		if err := vault.RemoveFromPreferences(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Vault '%s' removed from preferences\n", name)
	},
}

func init() {
	vaultCmd.AddCommand(removeVaultCmd)
}
