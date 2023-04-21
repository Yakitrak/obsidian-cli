package cmd

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var printDefaultCmd = &cobra.Command{
	Use:     "print-default",
	Aliases: []string{"pd"},
	Short:   "prints default Obsidian obsidian name and path",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{}
		name, err := v.DefaultName()
		if err != nil {
			log.Fatal(err)
		}
		path, err := v.Path()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default obsidian name: ", name)
		fmt.Println("Default obsidian path: ", path)
	},
}

func init() {
	rootCmd.AddCommand(printDefaultCmd)
}
