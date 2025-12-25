package cmd

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var setDefaultCmd = &cobra.Command{
	Use:     "set-default <vault>",
	Aliases: []string{"sd"},
	Short:   "Sets default vault",
	Long: `Sets the default vault for all commands.

The vault name must match exactly as it appears in Obsidian.
Once set, you won't need to specify --vault for each command.`,
	Example: `  # Set default vault
  obsidian-cli set-default "My Vault"

  # Verify it worked
  obsidian-cli print-default`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		v := obsidian.Vault{Name: name}
		if err := v.SetDefaultName(name); err != nil {
			return err
		}
		path, err := v.Path()
		if err != nil {
			return err
		}
		fmt.Println("Default vault set to: ", name)
		fmt.Println("Default vault path set to: ", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}
