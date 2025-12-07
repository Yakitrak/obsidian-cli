package embeddings

import "sort"

// RankConfig tunes the blend of embedding and graph signals.
type RankConfig struct {
	DirectLinkBonus float64
	TwoHopBonus     float64
	SharedTagBonus  float64
	Alpha           float64 // weight for embedding similarity
	Beta            float64 // weight for graph score
}

// RankRelated blends embedding similarity with lightweight graph features.
func RankRelated(src NoteID, cands []SimilarNote, g Graph, cfg RankConfig) []SimilarNote {
	if g == nil {
		for i := range cands {
			cands[i].FinalScore = cands[i].Score
		}
		sort.Slice(cands, func(i, j int) bool {
			return cands[i].FinalScore > cands[j].FinalScore
		})
		return cands
	}

	out := map[NoteID]struct{}{}
	for _, id := range g.Outgoing(src) {
		out[id] = struct{}{}
	}
	in := map[NoteID]struct{}{}
	for _, id := range g.Incoming(src) {
		in[id] = struct{}{}
	}

	srcTags := map[string]struct{}{}
	for _, t := range g.Tags(src) {
		srcTags[t] = struct{}{}
	}

	for i := range cands {
		id := cands[i].ID
		var graphScore float64

		if _, ok := out[id]; ok {
			graphScore += cfg.DirectLinkBonus
		}
		if _, ok := in[id]; ok {
			graphScore += cfg.DirectLinkBonus
		}

		if len(out) > 0 {
			for _, n := range g.Outgoing(id) {
				if _, ok := out[n]; ok {
					graphScore += cfg.TwoHopBonus
					break
				}
			}
		}

		if len(srcTags) > 0 {
			for _, t := range g.Tags(id) {
				if _, ok := srcTags[t]; ok {
					graphScore += cfg.SharedTagBonus
					break
				}
			}
		}

		cands[i].GraphScore = graphScore
		cands[i].FinalScore = cfg.Alpha*cands[i].Score + cfg.Beta*graphScore
	}

	sort.Slice(cands, func(i, j int) bool {
		return cands[i].FinalScore > cands[j].FinalScore
	})

	return cands
}
