package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "notesmd-cli",
	Short:   "Interact with Obsidian vaults from the terminal",
	Version: "v0.3.7",
	Long:    "Interact with Obsidian vaults from the terminal",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
