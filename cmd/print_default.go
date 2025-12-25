package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var printPathOnly bool
var printDefaultCmd = &cobra.Command{
	Use:     "print-default",
	Aliases: []string{"pd"},
	Short:   "Prints default vault name and path",
	Long: `Shows the currently configured default vault.

Use --path-only to output just the path, useful for scripting.`,
	Example: `  # Show default vault info
  obsidian-cli print-default

  # Get just the path (for scripts)
  obsidian-cli print-default --path-only`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{}
		name, err := vault.DefaultName()
		if err != nil {
			return err
		}
		path, err := vault.Path()
		if err != nil {
			return err
		}

		if printPathOnly {
			fmt.Print(path)
			return nil
		}

		fmt.Println("Default vault name: ", name)
		fmt.Println("Default vault path: ", path)
		return nil
	},
}

func init() {
	printDefaultCmd.Flags().BoolVar(&printPathOnly, "path-only", false, "print only the vault path")
	rootCmd.AddCommand(printDefaultCmd)
}
