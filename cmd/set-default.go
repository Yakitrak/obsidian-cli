package cmd

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/spf13/cobra"
)

var setDefaultCmd = &cobra.Command{
	Use:     "set-default",
	Aliases: []string{"sd"},
	Short:   "Sets default Obsidian vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		v := handler.Vault{Name: name}
		err := v.SetDefaultName(name)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Default vault set to: ", name)
		}

	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}
