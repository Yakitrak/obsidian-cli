package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	ncIncludeBacklinks   bool
	ncIncludeNeighbors   bool
	ncIncludeFrontmatter bool
	ncIncludeTags        bool
	ncNeighborLimit      int
	ncBacklinksLimit     int
)

type simpleNoteContext struct {
	Path        string                 `json:"path"`
	Title       string                 `json:"title,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Graph       map[string]interface{} `json:"graph,omitempty"`
	Community   map[string]interface{} `json:"community,omitempty"`
	Neighbors   map[string][]string    `json:"neighbors,omitempty"`
	Backlinks   []obsidian.Backlink    `json:"backlinks,omitempty"`
}

var graphNoteContextCmd = &cobra.Command{
	Use:   "note-context",
	Short: "Print note context (graph + community) for one or more files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(contextFilesFlag) == 0 {
			return fmt.Errorf("--files is required")
		}

		selectedVault := vaultName
		if selectedVault == "" {
			v := &obsidian.Vault{}
			name, err := v.DefaultName()
			if err != nil {
				return err
			}
			selectedVault = name
		}

		vault := obsidian.Vault{Name: selectedVault}
		note := obsidian.Note{}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		analysis, err := actions.GraphAnalysis(&vault, &note, actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: graphSkipAnchors,
					SkipEmbeds:  graphSkipEmbeds,
				},
				IncludeTags:       ncIncludeTags,
				MinDegree:         graphMinDegree,
				MutualOnly:        graphMutualOnly,
				RecencyCascade:    graphRecencyCascade,
				RecencyCascadeSet: true,
			},
		})
		if err != nil {
			return err
		}

		reverse := buildReverseNeighborsCLI(analysis.Nodes)
		communityLookup := obsidian.CommunityMembershipLookup(analysis.Communities)

		normalizedTargets := make([]string, 0, len(contextFilesFlag))
		for _, f := range contextFilesFlag {
			normalizedTargets = append(normalizedTargets, obsidian.NormalizePath(obsidian.AddMdSuffix(f)))
		}

		var backlinks map[string][]obsidian.Backlink
		if ncIncludeBacklinks {
			blOptions := obsidian.WikilinkOptions{SkipAnchors: graphSkipAnchors, SkipEmbeds: graphSkipEmbeds}
			backlinks, err = obsidian.CollectBacklinks(vaultPath, &note, normalizedTargets, blOptions, nil)
			if err != nil {
				return fmt.Errorf("error collecting backlinks: %w", err)
			}
		}

		contexts := make([]simpleNoteContext, 0, len(normalizedTargets))
		for _, target := range normalizedTargets {
			ctx := simpleNoteContext{Path: target}
			node, ok := analysis.Nodes[target]
			if !ok {
				ctx.Error = "not found in graph (respect include/exclude filters)"
				contexts = append(contexts, ctx)
				continue
			}
			ctx.Title = node.Title
			if ncIncludeTags {
				ctx.Tags = node.Tags
			}
			if ncIncludeFrontmatter {
				// best-effort: we don't have parsed frontmatter here
				ctx.Frontmatter = map[string]interface{}{}
			}

			ctx.Graph = map[string]interface{}{
				"inbound":   node.Inbound,
				"outbound":  node.Outbound,
				"hub":       node.Hub,
				"authority": node.Authority,
			}

			if comm := communityLookup[target]; comm != nil {
				ctx.Community = map[string]interface{}{
					"id":              comm.ID,
					"size":            len(comm.Nodes),
					"fractionOfVault": float64(len(comm.Nodes)) / float64(analysis.Stats.NodeCount),
					"anchor":          comm.Anchor,
					"density":         comm.Density,
					"recency":         comm.Recency,
				}
				if ncIncludeTags {
					ctx.Community["topTags"] = comm.TopTags
				}
			}

			if ncIncludeNeighbors {
				linksOut := node.Neighbors
				if len(linksOut) > ncNeighborLimit && ncNeighborLimit > 0 {
					linksOut = linksOut[:ncNeighborLimit]
				}
				linksIn := reverse[target]
				if len(linksIn) > ncNeighborLimit && ncNeighborLimit > 0 {
					linksIn = linksIn[:ncNeighborLimit]
				}
				ctx.Neighbors = map[string][]string{
					"linksOut": linksOut,
					"linksIn":  linksIn,
				}
			}

			if ncIncludeBacklinks && backlinks != nil {
				if bl := backlinks[target]; len(bl) > 0 {
					if len(bl) > ncBacklinksLimit && ncBacklinksLimit > 0 {
						bl = bl[:ncBacklinksLimit]
					}
					ctx.Backlinks = bl
				}
			}

			contexts = append(contexts, ctx)
		}

		payload := map[string]interface{}{
			"contexts": contexts,
			"count":    len(contexts),
		}
		if graphTimings {
			payload["timings"] = analysis.Timings
		}
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
		return nil
	},
}

func buildReverseNeighborsCLI(nodes map[string]obsidian.GraphNode) map[string][]string {
	reverse := make(map[string][]string, len(nodes))
	for path := range nodes {
		reverse[path] = nil
	}
	for src, node := range nodes {
		for _, dst := range node.Neighbors {
			reverse[dst] = append(reverse[dst], src)
		}
	}
	for path := range reverse {
		sort.Strings(reverse[path])
	}
	return reverse
}

var contextFilesFlag []string

func init() {
	graphNoteContextCmd.Flags().StringSliceVar(&contextFilesFlag, "files", nil, "comma-separated list of files (required)")
	graphNoteContextCmd.Flags().BoolVar(&ncIncludeBacklinks, "backlinks", true, "include backlinks")
	graphNoteContextCmd.Flags().BoolVar(&ncIncludeNeighbors, "neighbors", true, "include neighbors")
	graphNoteContextCmd.Flags().BoolVar(&ncIncludeFrontmatter, "frontmatter", false, "include frontmatter (best-effort)")
	graphNoteContextCmd.Flags().BoolVar(&ncIncludeTags, "tags", true, "include tags")
	graphNoteContextCmd.Flags().IntVar(&ncNeighborLimit, "neighbor-limit", 50, "max neighbors per direction (0 = all)")
	graphNoteContextCmd.Flags().IntVar(&ncBacklinksLimit, "backlinks-limit", 50, "max backlinks (0 = all)")

	graphCmd.AddCommand(graphNoteContextCmd)
}
