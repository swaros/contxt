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

import (
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/diff"
	"github.com/swaros/contxt/module/yacl"
)

type Linter struct {
	config    yacl.ConfigModel // the config model that we need to verify
	lMap      LintMap          // contains the diff chunks
	diffFound bool             // true if we found a diff. that is just a sign, that an SOME diff is found, not that the config is invalid
}

func NewLinter(config yacl.ConfigModel) *Linter {

	return &Linter{
		config: config,
	}

}

func (l *Linter) Verify() error {
	fileName := l.config.GetLoadedFile()

	m := make(map[string]interface{})          // generic map to load the file for comparison
	yamcLoader := l.config.GetLastUsedReader() // the last used reader from the config
	if yamcLoader == nil {
		return fmt.Errorf("no reader found. the config needs to be loaded first")

	}
	if err := yamcLoader.FileDecode(fileName, &m); err != nil { // decode the file to the generic map
		return err
	} else {
		bytes, err := yamcLoader.Marshal(m) // encode the generic map to source
		if err == nil {
			freeStyle := string(bytes) // the source as string
			//fmt.Println(freeStyle)
			cYamc, cerr := l.config.GetAsYmac() // get the configuration as yamc object
			if cerr == nil {
				//fmt.Println("-----------------")
				configStyle := cYamc.GetData()                   // get the source as string from the yamc object
				cbytes, ccerr := yamcLoader.Marshal(configStyle) // encode the source to bytes
				if ccerr == nil {
					//fmt.Println(string(cbytes))
					//differ := diff.Diff(freeStyle, string(cbytes))
					//fmt.Println("-----------------")
					//fmt.Println(differ)
					//fmt.Println("-----------------")

					freeChnk := strings.Split(freeStyle, "\n")
					orgiChnk := strings.Split(string(cbytes), "\n")

					chunk := diff.DiffChunks(freeChnk, orgiChnk)
					l.chunkWorker(chunk)
					l.FindPairs()
					l.ValueVerify()
				} else {
					return ccerr
				}
			} else {
				return cerr
			}
		} else {
			return err
		}
	}
	return nil
}

// chunkWorker is a worker that is called for each chunk that is found.
// in the diff. It will create a LintMap that contains the chunks
// for later investigation, if needed.
// if no diff found at all, that is all what we need to do.
func (l *Linter) chunkWorker(chunks []diff.Chunk) {
	lintResult := LintMap{}
	foundDiff := false
	// the diff package reports any change as an added and a removed line.
	// thats fine for printing a diff, but not for our needs.
	// we need to get the context what add and remove is just a change.
	// for this we need to count the changes over all chunks.
	// in the end, we have an safe match, if we have the same sequence number and the same change number.

	sequenceNr := 0
	changeNr4Add := 0 // track the index of added lines
	changeNr4Rm := 0  // track the index of removed lines

	// iterate over all chunks.
	for chunkIndex, c := range chunks {
		temporaryChunk := LintChunk{}
		needToBeAdded := false

		for _, line := range c.Added {
			changeNr4Add++
			//fmt.Println("ADDED:"+line, " ---> index[", indexNr, "] seq[", sequenceNr, "]", "chunk[", chunkIndex, "]", "change[", changeNr4Add, "]")

			addToken := NewMatchToken(&l.lMap, line, changeNr4Add, sequenceNr, true)
			temporaryChunk.Added = append(temporaryChunk.Added, &addToken)
			needToBeAdded = true
		}
		for _, line := range c.Deleted {
			changeNr4Rm++
			//fmt.Println("DELETED:"+line, " --->index[", indexNr, "] seq[", sequenceNr, "]", "chunk[", chunkIndex, "]", "change[", changeNr4Rm, "]")

			rmToken := NewMatchToken(&l.lMap, line, changeNr4Rm, sequenceNr, false)
			temporaryChunk.Removed = append(temporaryChunk.Removed, &rmToken)
			needToBeAdded = true
		}

		// on equal we need to reset the indices for removed and added changes, so we can track the
		// changes in the next chunk, and get the context what of these removas and adds are just changes
		if len(c.Equal) > 0 {
			changeNr4Add = 0
			changeNr4Rm = 0
			sequenceNr += len(c.Equal)
		}

		if needToBeAdded {
			temporaryChunk.ChunkNr = chunkIndex
			lintResult.Chunks = append(lintResult.Chunks, &LintChunk{
				ChunkNr: chunkIndex,
				Added:   temporaryChunk.Added,
				Removed: temporaryChunk.Removed,
			})
			foundDiff = true
		}
	}
	l.diffFound = foundDiff
	// only if we found a diff, we need to store the chunks
	// so we can verify the if the diffs matter
	if foundDiff {
		l.lMap = lintResult
	}
}

func (l *Linter) FindPairs() {
	if l.diffFound {
		// depends on the reult of the diff, we need to find the pairs
		// in the diff chunks. and they have the matching part in the previous diff chunk.
		// so we need to find the matching part in the previous chunk.

		for _, chunk := range l.lMap.Chunks {
			for _, add := range chunk.Added {
				foundmatch := false
				bestmatchTokens := l.lMap.GetTokensFromSequenceAndIndex(add.SequenceNr, add.indexNr)
				if len(bestmatchTokens) > 0 {
					for _, bestmatch := range bestmatchTokens {
						if bestmatch.Added != add.Added {
							if add.IsPair(bestmatch) {
								fmt.Println("FOUND PAIR: ", add.KeyWord, " ---> ", bestmatch.KeyWord)
								foundmatch = true
							}
						}
					}
				}
				if !foundmatch {
					// fallback.
					for _, seqTkn := range l.lMap.GetTokensFromSequence(add.SequenceNr) {
						if add.IsPair(seqTkn) {
							fmt.Println("fallback .... FOUND PAIR IN SEQUENCE: ", add.KeyWord, " ---> ", add.KeyWord)
						}
					}
				}

			}

		}
	}
}

func (l *Linter) ValueVerify() {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			token.VerifyValue()
		})
	}

}

func (l *Linter) PrintDiff() {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if added {
				switch token.Status {
				case ValueMatchButTypeDiffers:
					fmt.Println("   ValueMatchButTypeDiffers: ", token.Type, " ---> ", token.PairToken.Type)

				case MissingEntry:
					fmt.Println("   MissingEntry: ", token.KeyWord)

				}
			}
		})
	}
}

// proxy to walk all
func (l *Linter) walkAll(hndl func(token *MatchToken, added bool)) {
	l.lMap.walkAll(hndl)
}
