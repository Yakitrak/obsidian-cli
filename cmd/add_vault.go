package cmd

import (
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

var (
	addVaultForce bool
)

var addVaultCmd = &cobra.Command{
	Use:   "add-vault [name] [path]",
	Short: "Add or override a vault path in obsidian-cli preferences",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		path := args[1]

		vault := obsidian.Vault{Name: name}
		if err := vault.SavePathToPreferences(path, addVaultForce); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Vault '%s' saved at %s\n", name, path)
	},
}

func init() {
	addVaultCmd.Flags().BoolVar(&addVaultForce, "force", false, "overwrite existing path for this vault name")
	rootCmd.AddCommand(addVaultCmd)
}
