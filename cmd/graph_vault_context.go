package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	vcMaxCommunities    int
	vcCommunityTopNotes int
	vcCommunityTopTags  int
)

type simpleCommunity struct {
	ID              string                    `json:"id"`
	Size            int                       `json:"size"`
	FractionOfVault float64                   `json:"fractionOfVault"`
	Anchor          string                    `json:"anchor,omitempty"`
	Density         float64                   `json:"density,omitempty"`
	TopTags         []obsidian.TagCount       `json:"topTags,omitempty"`
	TopAuthority    []obsidian.AuthorityScore `json:"topAuthority,omitempty"`
	AuthorityStats  *obsidian.AuthorityStats  `json:"authorityStats,omitempty"`
	Recency         *obsidian.GraphRecency    `json:"recency,omitempty"`
}

type vaultContextPayload struct {
	Stats       obsidian.GraphStatsSummary  `json:"stats"`
	OrphanCount int                         `json:"orphanCount"`
	Orphans     []string                    `json:"orphans,omitempty"`
	Components  [][]string                  `json:"components,omitempty"`
	Communities []simpleCommunity           `json:"communities,omitempty"`
	Timings     obsidian.GraphTimingsMillis `json:"timings,omitempty"`
}

var graphVaultContextCmd = &cobra.Command{
	Use:   "vault-context",
	Short: "Print a JSON vault context (communities, stats, recency)",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		analysis, err := actions.GraphAnalysis(&vault, &note, actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: graphSkipAnchors,
					SkipEmbeds:  graphSkipEmbeds,
				},
				IncludeTags:       true,
				MinDegree:         graphMinDegree,
				MutualOnly:        graphMutualOnly,
				RecencyCascade:    graphRecencyCascade,
				RecencyCascadeSet: true,
			},
		})
		if err != nil {
			return err
		}

		if vcMaxCommunities <= 0 || vcMaxCommunities > len(analysis.Communities) {
			vcMaxCommunities = len(analysis.Communities)
		}

		if vcCommunityTopNotes <= 0 {
			vcCommunityTopNotes = 5
		}
		if vcCommunityTopTags <= 0 {
			vcCommunityTopTags = 5
		}

		comms := make([]simpleCommunity, 0, vcMaxCommunities)
		for idx, comm := range analysis.Communities {
			if idx >= vcMaxCommunities {
				break
			}
			topTags := comm.TopTags
			if len(topTags) > vcCommunityTopTags {
				topTags = topTags[:vcCommunityTopTags]
			}
			topAuthority := comm.TopAuthority
			if len(topAuthority) > vcCommunityTopNotes {
				topAuthority = topAuthority[:vcCommunityTopNotes]
			}
			c := simpleCommunity{
				ID:              comm.ID,
				Size:            len(comm.Nodes),
				FractionOfVault: float64(len(comm.Nodes)) / float64(analysis.Stats.NodeCount),
				Anchor:          comm.Anchor,
				Density:         comm.Density,
				TopTags:         topTags,
				TopAuthority:    topAuthority,
				AuthorityStats:  comm.AuthorityStats,
				Recency:         comm.Recency,
			}
			comms = append(comms, c)
		}

		payload := vaultContextPayload{
			Stats:       analysis.Stats,
			OrphanCount: len(analysis.Orphans),
			Orphans:     analysis.Orphans,
			Components:  analysis.WeakComponents,
			Communities: comms,
		}
		if graphTimings {
			payload.Timings = analysis.Timings.ToMillis()
		}

		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
		return nil
	},
}

func init() {
	graphVaultContextCmd.Flags().IntVar(&vcMaxCommunities, "max-communities", 25, "max communities to include (0 = all)")
	graphVaultContextCmd.Flags().IntVar(&vcCommunityTopNotes, "community-top-notes", 5, "top authority notes per community")
	graphVaultContextCmd.Flags().IntVar(&vcCommunityTopTags, "community-top-tags", 5, "top tags per community")

	graphCmd.AddCommand(graphVaultContextCmd)
}
