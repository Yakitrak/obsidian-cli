package cmd

import (
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"log"
)

var printDefaultCmd = &cobra.Command{
	Use:     "print-default",
	Aliases: []string{"pd"},
	Short:   "prints default vault name and path",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{}
		name, err := vault.DefaultName()
		if err != nil {
			log.Fatal(err)
		}
		path, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default vault name: ", name)
		fmt.Println("Default vault path: ", path)
	},
}

func init() {
	vaultCmd.AddCommand(printDefaultCmd)
}
