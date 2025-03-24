package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "List files in vault with contents formatted for LLM consumption, such as tag:a-tag or find:a-filename-pattern",
	Long: `List files in your Obsidian vault with contents formatted for LLM consumption.
Similar to the list command, but outputs file contents in a format optimized for LLMs.

Examples:
  obsidian-cli prompt Notes  					 						# the Notes folder
  obsidian-cli prompt find:joe  									# Filename containing "joe"
  obsidian-cli prompt find:'n/s joe'  						# Notes in folder starting with "n" whose name contains a word starting with "s" and a word starting with "joe"
  obsidian-cli prompt tag:career-pathing 					# Notes tagged with "career-pathing"
  obsidian-cli prompt tag:"career-pathing" -d 2 	# Notes tagged with "career-pathing", notes they link to, and the notes those link to
  obsidian-cli prompt find:project -f --skip-anchors   # Notes containing "project" and notes they link to, excluding links with section anchors
  obsidian-cli prompt find:notes -f --skip-embeds      # Notes containing "notes" and notes they link to, excluding embedded links
  obsidian-cli prompt find:docs -f --skip-anchors --skip-embeds  # Skip both anchored and embedded links
	`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug output if debug flag is set
		actions.Debug = debug
		
		// If maxDepth is greater than 0, enable followLinks
		if maxDepth > 0 {
			followLinks = true
		}
		
		// Check if any inputs were provided
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: at least one input is required\n")
			fmt.Fprintf(os.Stderr, "Run 'obsidian-cli prompt --help' for usage.\n")
			os.Exit(1)
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

		// Parse inputs using the helper function
		inputs, err := actions.ParseInputs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Create a map to track unique files
		uniqueFiles := make(map[string]bool)
		// Create a mutex to safely print files
		var printMu sync.Mutex

		// Get vault path
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		// Buffer to store output
		var output strings.Builder

		// Print initial search message if in terminal mode
		isTerminalMode := isTerminal()
		if isTerminalMode {
			for _, input := range inputs {
				switch input.Type {
				case actions.InputTypeTag:
					fmt.Fprintf(os.Stderr, "Searching for tag %q\n", input.Value)
				case actions.InputTypeFind:
					fmt.Fprintf(os.Stderr, "Searching for %q\n", input.Value)
				case actions.InputTypeFile:
					fmt.Fprintf(os.Stderr, "Including file %q\n", input.Value)
				}
			}
		}

		// Print the vault header
		fmt.Fprintf(&output, "<obsidian-vault name=\"%s\">\n\n", vaultName)

		// Call ListFiles with a callback to print files as they're found
		_, err = actions.ListFiles(&vault, &note, actions.ListParams{
			Inputs:        inputs,
			FollowLinks:   followLinks,
			MaxDepth:      maxDepth,
			SkipAnchors:   skipAnchors,
			SkipEmbeds:    skipEmbeds,
			AbsolutePaths: absolutePaths,
			OnMatch: func(file string) {
				printMu.Lock()
				if !uniqueFiles[file] {
					uniqueFiles[file] = true

					// Read the file contents
					filePath := filepath.Join(vaultPath, file)
					content, err := ioutil.ReadFile(filePath)
					if err != nil {
						log.Printf("Error reading file %s: %v", file, err)
						return
					}

					// Print progress message if in terminal mode
					if isTerminalMode {
						fmt.Fprintf(os.Stderr, "Found %s\n", file)
					}

					// Add to output buffer
					fmt.Fprintf(&output, "<file path=\"%s\">\n%s\n</file>\n\n", file, string(content))
				}
				printMu.Unlock()
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		// Print the closing vault tag
		fmt.Fprintf(&output, "</obsidian-vault>")

		// If terminal, copy to clipboard and notify on stderr
		if isTerminalMode {
			err := clipboard.WriteAll(output.String())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(os.Stderr, "Copied %d files to clipboard in LLM-friendly format\n", len(uniqueFiles))
		} else {
			// If piped, print to stdout
			fmt.Print(output.String())
		}
	},
}

func init() {
	promptCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	promptCmd.Flags().BoolVarP(&followLinks, "follow", "f", false, "follow wikilinks recursively")
	promptCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum depth for following wikilinks (0 means don't follow)")
	promptCmd.Flags().BoolVar(&skipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	promptCmd.Flags().BoolVar(&skipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")
	promptCmd.Flags().BoolVarP(&absolutePaths, "absolute", "a", false, "print absolute paths")
	promptCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.AddCommand(promptCmd)
}
