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
			fmt.Println(freeStyle)
			cYamc, cerr := l.config.GetAsYmac() // get the configuration as yamc object
			if cerr == nil {
				fmt.Println("-----------------")
				configStyle := cYamc.GetData()                   // get the source as string from the yamc object
				cbytes, ccerr := yamcLoader.Marshal(configStyle) // encode the source to bytes
				if ccerr == nil {
					fmt.Println(string(cbytes))
					differ := diff.Diff(freeStyle, string(cbytes))
					fmt.Println("-----------------")
					fmt.Println(differ)
					fmt.Println("-----------------")

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
	for fg, c := range chunks {
		sequenceNr := 0 // this number is inreased for any matching line. so we are able to find the line in the original file.
		temporaryChunk := LintChunk{}
		needToBeAdded := false
		for sg, line := range c.Added {

			fmt.Println("ADDED:"+line, " --->", fg, sequenceNr)

			addToken := NewMatchToken(&l.lMap, line, sg, true)
			temporaryChunk.Added = append(temporaryChunk.Added, &addToken)
			needToBeAdded = true
		}
		for sg, line := range c.Deleted {

			fmt.Println("DELETED:"+line, " --->", fg, sequenceNr)

			rmToken := NewMatchToken(&l.lMap, line, sg, false)
			temporaryChunk.Removed = append(temporaryChunk.Removed, &rmToken)
			needToBeAdded = true
		}

		for _, line := range c.Equal {
			sequenceNr++ // anytime a match is reported, we are out of any add remove section

			fmt.Println("EQUAL:"+line, " --->", fg, sequenceNr)

		}
		if needToBeAdded {
			temporaryChunk.ChunkNr = fg
			lintResult.Chunks = append(lintResult.Chunks, &LintChunk{
				ChunkNr: fg,
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

		lastChunk := &LintChunk{}
		for _, chunk := range l.lMap.Chunks {
			for _, add := range chunk.Added {
				if lastChunk.ChunkNr+1 == chunk.ChunkNr {
					for _, rm := range lastChunk.Removed {
						if add.IsPair(rm) {
							fmt.Println("FOUND PAIR: ", add.KeyWord, " ---> ", add.KeyWord)
						}
					}
				}
			}
			lastChunk = chunk
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
					fmt.Println("ValueMatchButTypeDiffers: ", token.Type, " ---> ", token.PairToken.Type)

				case MissingEntry:
					fmt.Println("MissingEntry: ", token.KeyWord)

				}
			}
		})
	}
}

// helper function to walk all tokens without writeing the same loop again and again
func (l *Linter) walkAll(hndl func(token *MatchToken, added bool)) {
	for _, chunk := range l.lMap.Chunks {
		for _, add := range chunk.Added {
			hndl(add, true)
		}
		for _, rm := range chunk.Removed {
			hndl(rm, false)
		}
	}
}
