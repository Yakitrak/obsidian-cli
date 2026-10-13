package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Yakitrak/notesmd-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var removeVaultCmd = &cobra.Command{
	Use:     "remove-vault <name|path>",
	Aliases: []string{"rv"},
	Short:   "Unregister a vault",
	Long:    "Removes a vault from the Obsidian config. Does not delete any files on disk.\n\n" + vaultRegistryWarning,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErrln("Warning: " + vaultRegistryWarning)
		input := args[0]

		name, err := obsidian.RemoveVault(input)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Vault %q removed\n", name)

		if err := obsidian.ClearDefaultIfMatch(name); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not clear default vault:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeVaultCmd)
}
