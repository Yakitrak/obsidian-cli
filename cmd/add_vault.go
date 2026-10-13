package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Yakitrak/notesmd-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var addVaultCmd = &cobra.Command{
	Use:     "add-vault <path>",
	Aliases: []string{"av"},
	Short:   "Register a vault directory",
	Long:    "Registers a directory as an Obsidian vault. Creates the Obsidian config file if it does not exist.\n\n" + vaultRegistryWarning,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErrln("Warning: " + vaultRegistryWarning)
		absPath, err := obsidian.AddVault(args[0])
		if err != nil {
			log.Fatal(err)
		}

		name := filepath.Base(absPath)
		fmt.Printf("Vault %q registered at: %s\n", name, absPath)

		setDefault, _ := cmd.Flags().GetBool("set-default")
		if setDefault {
			v := obsidian.Vault{Name: name}
			if err := v.SetDefaultName(name); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Default vault set to:", name)
		}
	},
}

func init() {
	addVaultCmd.Flags().Bool("set-default", false, "set the added vault as the default")
	rootCmd.AddCommand(addVaultCmd)
}
