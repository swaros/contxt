// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
