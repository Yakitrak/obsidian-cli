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
	Short:   "List files in vault with various filtering options, such as tag:a-tag or find:a-filename-pattern",
	Long: `List files in your Obsidian vault with various filtering options:

Examples:
  obsidian-cli list Notes                          # the Notes folder
  obsidian-cli list find:joe                       # Filename containing "joe"
  obsidian-cli list find:'n/s joe'                 # Notes in folder starting with "n" whose name contains a word starting with "s" and a word starting with "joe"
  obsidian-cli list tag:career-pathing             # Notes tagged with "career-pathing"
  obsidian-cli list tag:"career-pathing" -d 2      # Notes tagged with "career-pathing", notes they link to, and the notes those link to`,
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
		if debug {
			log.Printf("Using default vault name: %s", vaultName)
		}
	}

	vault := obsidian.Vault{Name: vaultName}
	note := obsidian.Note{}

	if debug {
		log.Printf("Number of input args: %d", len(args))
	}

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

	// If no inputs provided, add a wildcard input
	if len(inputs) == 0 {
		inputs = append(inputs, actions.ListInput{
			Type:  actions.InputTypeFile,
			Value: "*",
		})
	}

	if debug {
		log.Printf("Number of parsed inputs: %d", len(inputs))
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

	if debug {
		log.Printf("Vault path: %s", vaultPath)
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

	return nil
}
