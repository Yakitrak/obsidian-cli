package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "obs",
	Short: "obs - CLI to search and open notes in obsidian",
	Long: `obs - CLI to search and open notes in obsidian
   
It can open a vault, open, search and create notes in your vault(s)`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
