package cmd

import (
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var listVaultsCmd = &cobra.Command{
	Use:   "list-vaults",
	Short: "List vault mappings stored in obsidian-cli preferences",
	Run: func(cmd *cobra.Command, args []string) {
		vaults, defaultName, err := obsidian.ListPreferenceVaults()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Manual vaults (obsidian-cli preferences):")
		if len(vaults) == 0 {
			fmt.Println("  (none)")
		} else {
			names := make([]string, 0, len(vaults))
			for name := range vaults {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				path := vaults[name].Path
				if name == defaultName {
					fmt.Printf("  %s (default) -> %s\n", name, path)
					continue
				}
				fmt.Printf("  %s -> %s\n", name, path)
			}
		}

		fmt.Println()
		fmt.Println("Obsidian app vaults (obsidian.json):")
		appVaults, err := obsidian.ListObsidianVaults()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read Obsidian vaults: %v\n", err)
			return
		}

		if len(appVaults) == 0 {
			fmt.Println("  (none)")
			return
		}

		keys := make([]string, 0, len(appVaults))
		for key := range appVaults {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			entry := appVaults[key]
			if entry.Path == "" {
				continue
			}
			name := filepath.Base(entry.Path)
			fmt.Printf("  %s -> %s\n", name, entry.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(listVaultsCmd)
}
