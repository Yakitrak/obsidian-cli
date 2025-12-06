package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "obsidian-cli",
	Short:   "obsidian-cli - CLI to open, search, move, create, delete and update notes",
	Version: "v0.1.8",
	Long:    "obsidian-cli - CLI to open, search, move, create, delete and update notes",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(vaultCmd)
	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(propertiesCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(OpenVaultCmd)
	rootCmd.AddCommand(createNoteCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(DailyCmd)
	rootCmd.AddCommand(printCmd)
	rootCmd.AddCommand(installIgnoreCmd)
	rootCmd.AddCommand(printDefaultCmd)
	rootCmd.AddCommand(mcpCmd)
}
