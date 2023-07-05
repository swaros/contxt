package yaclint

type LintChunk struct {
	ChunkNr int
	Removed []*MatchToken
	Added   []*MatchToken
}

type LintMap struct {
	Chunks []*LintChunk
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
		if token.IndexNr == index {
			tokens = append(tokens, token)
		}

	}
	return tokens
}

// find tokens by keypath over all chunks
func (l *LintMap) GetTokensByTokenPath(fromToken *MatchToken) []*MatchToken {
	var tokens []*MatchToken
	l.walkAll(func(token *MatchToken, added bool) {
		if token.KeyPath != "" && token.KeyPath == fromToken.KeyPath {
			tokens = append(tokens, token)
		}
	})
	return tokens
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
