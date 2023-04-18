package cmd

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/spf13/cobra"
)

var printDefaultCmd = &cobra.Command{
	Use:     "print-default",
	Aliases: []string{"pd"},
	Short:   "prints default Obsidian vault name and path",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		v := handler.Vault{}
		name, err := v.DefaultName()
		if err != nil {
			panic(err)
		}
		path, err := v.Path()
		if err != nil {
			panic(err)
		}
		fmt.Println("Default vault name: ", name)
		fmt.Println("Default vault path: ", path)
	},
}

func init() {
	rootCmd.AddCommand(printDefaultCmd)
}
