package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	maxDepth         int
	skipAnchors      bool
	skipEmbeds       bool
	absolutePaths    bool
	debug            bool
	includeBacklinks bool
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
  obsidian-cli list tag:"career-pathing" -d 2      # Notes tagged with "career-pathing", notes they link to, and the notes those link to
  obsidian-cli list find:project -d 1 --skip-anchors # Notes whose names match pattern plus directly linked notes, excluding anchors
  obsidian-cli list find:notes -d 1 --skip-embeds    # Notes whose names match pattern plus directly linked notes, excluding embedded links`,
	Args: cobra.ArbitraryArgs,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	listCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum depth for following wikilinks (0 means don't follow)")
	listCmd.Flags().BoolVar(&skipAnchors, "skip-anchors", false, "skip wikilinks that contain anchors (e.g. [[Note#Section]])")
	listCmd.Flags().BoolVar(&skipEmbeds, "skip-embeds", false, "skip embedded wikilinks (e.g. ![[Embedded Note]])")
	listCmd.Flags().BoolVarP(&absolutePaths, "absolute", "a", false, "print absolute paths")
	listCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output")
	listCmd.Flags().BoolVar(&includeBacklinks, "backlinks", false, "include first-degree backlinks for each matched file")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	actions.Debug = debug

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

	// Parse inputs using the helper function
	inputs, expr, err := actions.ParseInputsWithExpression(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return err
	}

	// If no inputs were provided, use a wildcard
	if len(inputs) == 0 {
		inputs = append(inputs, actions.ListInput{
			Type:  actions.InputTypeFile,
			Value: "*",
		})
	}

	if debug {
		log.Printf("Number of parsed inputs: %d", len(inputs))
	}

	// Mutex for printing when needed
	var printMu sync.Mutex

	// Get vault path for absolute paths
	vaultPath, err := vault.Path()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Printf("Vault path: %s", vaultPath)
	}

	params := actions.ListParams{
		Inputs:        inputs,
		MaxDepth:      maxDepth,
		SkipAnchors:   skipAnchors,
		SkipEmbeds:    skipEmbeds,
		AbsolutePaths: absolutePaths,
		Expression:    expr,
	}

	var backlinks map[string][]obsidian.Backlink
	var primaryMatches []string

	if includeBacklinks {
		params.IncludeBacklinks = true
		params.Backlinks = &backlinks
		params.PrimaryMatches = &primaryMatches
	}

	matches, err := actions.ListFiles(&vault, &note, params)
	if err != nil {
		log.Fatal(err)
	}

	seen := make(map[string]bool)
	printPath := func(p string) {
		printMu.Lock()
		defer printMu.Unlock()
		if seen[p] {
			return
		}
		seen[p] = true
		path := p
		if absolutePaths {
			if absPath, err := filepath.Abs(filepath.Join(vaultPath, p)); err == nil {
				path = absPath
			}
		}
		if isTerminal() {
			fmt.Printf("%s\n", path)
		} else {
			fmt.Printf("%q\n", path)
		}
	}

	if includeBacklinks {
		primaries := primaryMatches
		if len(primaries) == 0 {
			primaries = matches
		}

		for _, file := range matches {
			printPath(file)
		}

		for _, file := range primaries {
			key := obsidian.NormalizePath(obsidian.AddMdSuffix(file))
			for _, bl := range backlinks[key] {
				printPath(bl.Referrer)
			}
		}

		return nil
	}

	for _, file := range matches {
		printPath(file)
	}

	return nil
}
