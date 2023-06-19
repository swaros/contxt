package yaclint

type LintChunk struct {
	ChunkNr int
	Removed []*MatchToken
	Added   []*MatchToken
}

type LintMap struct {
	Chunks []*LintChunk
}

func (l *LintMap) AddChunk(chunk LintChunk) {
	l.Chunks = append(l.Chunks, &chunk)
}

func (l *LintMap) GetMatchById(id string) *MatchToken {
	for _, chunk := range l.Chunks {
		for _, add := range chunk.Added {
			if add.UuId == id {
				return add
			}
		}
		for _, rm := range chunk.Removed {
			if rm.UuId == id {
				return rm
			}
		}
	}
	return nil
}
