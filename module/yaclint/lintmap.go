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

func (l *LintMap) GetTokensFromSequence(seq int) []*MatchToken {
	var tokens []*MatchToken
	l.walkAll(func(token *MatchToken, added bool) {
		if token.SequenceNr == seq {
			tokens = append(tokens, token)
		}
	})
	return tokens
}

func (l *LintMap) GetTokensFromSequenceAndIndex(seq int, index int) []*MatchToken {
	var tokens []*MatchToken
	for _, token := range l.GetTokensFromSequence(seq) {
		if token.indexNr == index {
			tokens = append(tokens, token)
		}

	}
	return tokens
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

func (l *LintMap) walkAll(hndl func(token *MatchToken, added bool)) {
	for _, chunk := range l.Chunks {
		for _, add := range chunk.Added {
			hndl(add, true)
		}
		for _, rm := range chunk.Removed {
			hndl(rm, false)
		}
	}
}
