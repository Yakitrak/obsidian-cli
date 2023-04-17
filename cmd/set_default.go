package cmd

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils/temp"

	"github.com/spf13/cobra"
)

var setDefaultCmd = &cobra.Command{
	Use:     "set-default",
	Aliases: []string{"sd"},
	Short:   "Sets default Obsidian vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Println("Setting vault location as default: ", name)
		temp.SetDefaultVault(name)
	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}
