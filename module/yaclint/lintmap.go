package yaclint

type LintChunk struct {
	ChunkNr int
	Removed []*MatchToken
	Added   []*MatchToken
}

type LintMap struct {
	Chunks []*LintChunk
}

// AddChunk adds a new chunk to the lint map
func (l *LintMap) AddChunk(chunk LintChunk) {
	l.Chunks = append(l.Chunks, &chunk)
}

// GetTokensFromSequence returns all tokens from the given sequence
func (l *LintMap) GetTokensFromSequence(seq int) []*MatchToken {
	var tokens []*MatchToken
	l.walkAll(func(token *MatchToken, added bool) {
		if token.SequenceNr == seq {
			tokens = append(tokens, token)
		}
	})
	return tokens
}

// GetTokensFromSequenceAndIndex returns all tokens from the given sequence and index
func (l *LintMap) GetTokensFromSequenceAndIndex(seq int, index int) []*MatchToken {
	var tokens []*MatchToken
	for _, token := range l.GetTokensFromSequence(seq) {
		if token.indexNr == index {
			tokens = append(tokens, token)
		}

	}
	return tokens
}

// GetMatchById returns the match token with the given id
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

// walkAll walks through all tokens and calls the given handler
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
