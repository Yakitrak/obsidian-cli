package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	suppressTags []string
	noSuppress   bool
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "List files in vault with contents formatted for LLM consumption, such as tag:a-tag or find:a-filename-pattern",
	Long: `List files in your Obsidian vault with contents formatted for LLM consumption.
Similar to the list command, but outputs file contents in a format optimized for LLMs.

By default, files tagged with "no-prompt" are excluded from output. This can be controlled with --suppress-tags and --no-suppress flags.

Examples:
  obsidian-cli prompt Notes  					 						# the Notes folder
  obsidian-cli prompt find:joe  									# Filename containing "joe"
  obsidian-cli prompt find:'n/s joe'  						# Notes in folder starting with "n" whose name contains a word starting with "s" and a word starting with "joe"
  obsidian-cli prompt tag:career-pathing 					# Notes tagged with "career-pathing"
  obsidian-cli prompt tag:"career-pathing" -d 2 	# Notes tagged with "career-pathing", notes they link to, and the notes those link to
  obsidian-cli prompt find:project -d 1 --skip-anchors   # Notes whose names match pattern plus directly linked notes, excluding anchors
  obsidian-cli prompt find:notes -d 1 --skip-embeds      # Notes whose names match pattern plus directly linked notes, excluding embedded links
  obsidian-cli prompt find:docs -d 1 --skip-anchors --skip-embeds  # Skip both anchored and embedded links
  obsidian-cli prompt tag:foo --suppress-tags private,draft     # Find tag:foo but exclude files with private or draft tags
  obsidian-cli prompt Notes --no-suppress                        # Don't exclude any tags (including no-prompt)
	`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Enable debug output if debug flag is set
		actions.Debug = debug

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
		inputs, expr, err := actions.ParseInputsWithExpression(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Configure suppressed tags
		var suppressedTags []string
		if !noSuppress {
			// Default suppression
			suppressedTags = append(suppressedTags, "no-prompt")
			// Add any additional suppressed tags
			suppressedTags = append(suppressedTags, suppressTags...)
		} else if len(suppressTags) > 0 {
			// Only use explicitly specified tags when --no-suppress is used
			suppressedTags = suppressTags
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
			fmt.Fprintf(os.Stderr, "Searching with: %q\n", strings.Join(args, " "))
			if len(suppressedTags) > 0 {
				fmt.Fprintf(os.Stderr, "Suppressing files with tags: %v\n", suppressedTags)
			}
		}

		// Print the vault header
		fmt.Fprintf(&output, "<obsidian-vault name=\"%s\">\n\n", vaultName)

		type match struct {
			path    string
			content string
		}
		var matches []match

		onMatch := func(file string) {
			printMu.Lock()
			defer printMu.Unlock()
			if uniqueFiles[file] {
				return
			}
			uniqueFiles[file] = true

			filePath := filepath.Join(vaultPath, file)
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", file, err)
				return
			}

			if isTerminalMode {
				fmt.Fprintf(os.Stderr, "Found %s\n", file)
			}

			matches = append(matches, match{path: file, content: string(content)})
		}

		params := actions.ListParams{
			Inputs:         inputs,
			MaxDepth:       maxDepth,
			SkipAnchors:    skipAnchors,
			SkipEmbeds:     skipEmbeds,
			AbsolutePaths:  absolutePaths,
			Expression:     expr,
			SuppressedTags: suppressedTags,
			OnMatch:        onMatch,
		}

		var backlinks map[string][]obsidian.Backlink
		var primaryMatches []string
		if includeBacklinks {
			params.IncludeBacklinks = true
			params.Backlinks = &backlinks
			params.PrimaryMatches = &primaryMatches
		}

		_, err = actions.ListFiles(&vault, &note, params)
		if err != nil {
			log.Fatal(err)
		}

		if includeBacklinks {
			primaries := primaryMatches
			if len(primaries) == 0 {
				for _, m := range matches {
					primaries = append(primaries, m.path)
				}
			}

			seen := make(map[string]bool)
			for _, m := range matches {
				seen[obsidian.NormalizePath(obsidian.AddMdSuffix(m.path))] = true
			}

			for _, p := range primaries {
				key := obsidian.NormalizePath(obsidian.AddMdSuffix(p))
				for _, bl := range backlinks[key] {
					ref := obsidian.NormalizePath(bl.Referrer)
					if seen[ref] {
						continue
					}
					content, err := ioutil.ReadFile(filepath.Join(vaultPath, ref))
					if err != nil {
						log.Printf("Error reading backlink referrer %s: %v", ref, err)
						continue
					}
					matches = append(matches, match{path: ref, content: string(content)})
					seen[ref] = true
				}
			}
		}

		for _, m := range matches {
			fmt.Fprintf(&output, "<file path=\"%s\">\n%s\n</file>\n", m.path, m.content)
			fmt.Fprintf(&output, "\n")
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
	promptCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum depth for following wikilinks (0 means don't follow)")
	promptCmd.Flags().BoolVar(&skipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	promptCmd.Flags().BoolVar(&skipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")
	promptCmd.Flags().BoolVarP(&absolutePaths, "absolute", "a", false, "print absolute paths")
	promptCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output")
	promptCmd.Flags().StringSliceVar(&suppressTags, "suppress-tags", nil, "additional tags to suppress/exclude from output (comma-separated)")
	promptCmd.Flags().BoolVar(&noSuppress, "no-suppress", false, "disable all tag suppression, including default no-prompt tag")
	promptCmd.Flags().BoolVar(&includeBacklinks, "backlinks", false, "include first-degree backlinks for each matched file")
	rootCmd.AddCommand(promptCmd)
}
