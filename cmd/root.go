package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "obsidian-cli",
	Short:   "obsidian-cli - CLI to open, search, move, create, delete and update notes",
	Version: "v0.2.0",
	Long:    "obsidian-cli - CLI to open, search, move, create, delete and update notes",
}

func Execute() {
	maybeRewriteArgsForTargetNames()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func maybeRewriteArgsForTargetNames() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}
	first := strings.TrimSpace(args[0])
	if first == "" || strings.HasPrefix(first, "-") {
		return
	}
	if isKnownRootCommand(first) {
		return
	}

	cfg, err := obsidian.LoadTargets()
	if err != nil {
		return
	}
	if _, ok := cfg[first]; !ok {
		return
	}

	// Treat: obsidian-cli <targetName> ... as: obsidian-cli target <targetName> ...
	rootCmd.SetArgs(append([]string{"target"}, args...))
}

func isKnownRootCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return true
		}
		for _, a := range c.Aliases {
			if a == name {
				return true
			}
		}
	}
	return false
}
