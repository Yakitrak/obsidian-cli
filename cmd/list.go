package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	followLinks   bool
	maxDepth      int
	absolutePaths bool
	debug         bool
)

// isTerminal returns true if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List files in vault with various filtering options",
	Long: `List files in your Obsidian vault with various filtering options:
- File paths (exact matches)
- Tag-based filtering (tag:some-tag)
- Fuzzy search (search:query)
- Optional recursive wikilink following

Examples:
  obsidian-cli list tag:career-pathing "./Notes/Ideas.md" search:TLS
  obsidian-cli list tag:"some-tag" tag:'another-tag'
  obsidian-cli list "./Notes" search:project`,
	Args: cobra.ArbitraryArgs,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	listCmd.Flags().BoolVarP(&followLinks, "follow", "f", false, "follow wikilinks recursively")
	listCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum depth for following wikilinks (0 means don't follow)")
	listCmd.Flags().BoolVarP(&absolutePaths, "absolute", "a", false, "print absolute paths")
	listCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	actions.Debug = debug

	// If maxDepth is greater than 0, enable followLinks
	if maxDepth > 0 {
		followLinks = true
	}

	// If no vault name is provided, get the default vault name
	if vaultName == "" {
		vault := &obsidian.Vault{}
		defaultName, err := vault.DefaultName()
		if err != nil {
			log.Fatal(err)
		}
		vaultName = defaultName
	}

	vault := obsidian.Vault{Name: vaultName}
	note := obsidian.Note{}

	// Parse inputs
	var inputs []actions.ListInput
	for _, arg := range args {
		if strings.HasPrefix(arg, "tag:") {
			// Handle quoted tags
			tag := strings.TrimPrefix(arg, "tag:")
			if strings.HasPrefix(tag, "\"") && strings.HasSuffix(tag, "\"") {
				tag = strings.Trim(tag, "\"")
			}
			inputs = append(inputs, actions.ListInput{
				Type:  actions.InputTypeTag,
				Value: tag,
			})
		} else if strings.HasPrefix(arg, "find:") {
			// Handle find input
			searchTerm := strings.TrimPrefix(arg, "find:")
			if strings.HasPrefix(searchTerm, "\"") && strings.HasSuffix(searchTerm, "\"") {
				searchTerm = strings.Trim(searchTerm, "\"")
			}
			inputs = append(inputs, actions.ListInput{
				Type:  actions.InputTypeFind,
				Value: searchTerm,
			})
		} else {
			// Handle file paths
			inputs = append(inputs, actions.ListInput{
				Type:  actions.InputTypeFile,
				Value: arg,
			})
		}
	}

	// Create a map to track unique files
	uniqueFiles := make(map[string]bool)
	// Create a mutex to safely print files
	var printMu sync.Mutex

	// Get vault path for absolute paths
	vaultPath, err := vault.Path()
	if err != nil {
		log.Fatal(err)
	}

	// Call ListFiles with a callback to print files as they're found
	_, err = actions.ListFiles(&vault, &note, actions.ListParams{
		Inputs:        inputs,
		FollowLinks:   followLinks,
		MaxDepth:      maxDepth,
		AbsolutePaths: absolutePaths,
		OnMatch: func(file string) {
			printMu.Lock()
			if !uniqueFiles[file] {
				uniqueFiles[file] = true
				path := file
				if absolutePaths {
					absPath, err := filepath.Abs(filepath.Join(vaultPath, file))
					if err == nil {
						path = absPath
					}
				}

				if isTerminal() {
					// When printing to terminal, never quote
					fmt.Printf("%s\n", path)
				} else {
					// When piped, always quote for safety
					fmt.Printf("%q\n", path)
				}
			}
			printMu.Unlock()
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// No need to print files here since they're printed by the callback
	return nil
}
