package cmd

import (
	"fmt"
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	printPathOnly   bool
	printDefaultCmd = &cobra.Command{
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

			if printPathOnly {
				fmt.Print(path)
				return
			}

			fmt.Println("Default vault name: ", name)
			fmt.Println("Default vault path: ", path)
		},
	}
)

func init() {
	printDefaultCmd.Flags().BoolVar(&printPathOnly, "path-only", false, "print only the vault path")
	rootCmd.AddCommand(printDefaultCmd)
}
