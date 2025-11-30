package cmd

import (
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var setDefaultCmd = &cobra.Command{
	Use:     "set-default",
	Aliases: []string{"sd"},
	Short:   "Sets default vault",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		v := obsidian.Vault{Name: name}
		err := v.SetDefaultName(name)
		if err != nil {
			log.Fatal(err)
		}
		path, err := v.Path()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default vault set to: ", name)
		fmt.Println("Default vault path set to: ", path)

	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}
