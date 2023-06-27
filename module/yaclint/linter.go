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
	"github.com/swaros/contxt/module/yamc"
)

type Linter struct {
	config            *yacl.ConfigModel // the config model that we need to verify
	lMap              LintMap           // contains the diff chunks
	diffFound         bool              // true if we found a diff. that is just a sign, that an SOME diff is found, not that the config is invalid
	highestIssueLevel int               // the highest issue level found
	structhandler     yamc.StructDef    // the struct handler for the config file. keeps the struct definition
}

func NewLinter(config yacl.ConfigModel) *Linter {
	return &Linter{
		config:            &config,
		highestIssueLevel: 0,
	}

}

// getUnstructMap loads the config file as generic map and returns it
// as map[string]interface{} and as string (the yaml/json representation)
func (l *Linter) getUnstructMap(loader yamc.DataReader) (map[string]interface{}, string, error) {
	fileName := l.config.GetLoadedFile() // the file name of the config file
	m := make(map[string]interface{})    // generic map to load the file for comparison
	err := loader.FileDecode(fileName, &m)
	if err != nil {
		return nil, "", err
	}
	bytes, err := loader.Marshal(m)
	if err != nil {
		return nil, "", err
	}
	return m, string(bytes), nil
}

// getStructSource creates the yaml/json representation of the config file
func (l *Linter) getStructSource(loader yamc.DataReader) (string, error) {
	cYamc, cerr := l.config.GetAsYmac() // get the configuration as yamc object
	if cerr != nil {
		return "", cerr
	}

	structData := cYamc.GetData()               // get the source as string from the yamc object
	cbytes, ccerr := loader.Marshal(structData) // encode the source to bytes
	if ccerr != nil {
		return "", ccerr
	}
	return string(cbytes), nil
}

// init4read is a helper function that initializes the linter for reading the config file.
func (l *Linter) init4read() (yamc.DataReader, string, string, error) {
	yamcLoader := l.config.GetLastUsedReader() // the last used reader from the config
	if yamcLoader == nil {
		return nil, "", "", fmt.Errorf("no reader found. the config needs to be loaded first")
	}
	l.structhandler = *yamcLoader.GetFields() // get the struct handler from the reader. must be done before the unstructed map is loaded
	_, unstructSource, err1 := l.getUnstructMap(yamcLoader)
	if err1 != nil {
		return nil, "", "", err1
	}

	structSource, err2 := l.getStructSource(yamcLoader)
	if err2 != nil {
		return nil, "", "", err2
	}
	return yamcLoader, unstructSource, structSource, nil
}

// GetDiff returns the diff between the config file and the structed config file.
// The diff is returned as string.
func (l *Linter) GetDiff() (string, error) {
	_, unstructedSrc, structedSrc, err := l.init4read()
	if err != nil {
		return "", err
	}
	return diff.Diff(unstructedSrc, structedSrc), nil

}

// Verify is the main function of the linter. It will verify the config file
// against the structed config file. It will return an error if the config file
// is not valid.
func (l *Linter) Verify() error {

	_, unstructSource, structSource, err := l.init4read()
	if err != nil {
		return err
	}

	freeChnk := strings.Split(unstructSource, "\n")
	orgiChnk := strings.Split(structSource, "\n")

	chunk := diff.DiffChunks(freeChnk, orgiChnk)
	l.chunkWorker(chunk)
	if !l.diffFound { // no diff found, so no need to go further
		return nil
	}
	l.findPairs()
	l.valueVerify()

	return nil
}

// GetHighestIssueLevel returns the highest issue level found.
func (l *Linter) GetHighestIssueLevel() int {
	return l.highestIssueLevel
}

// if the lint fails and do not report any error, the config could be just invalid for structure parsing.
// this can happen while the config file is tryed to be loaded, but it is not readable.
// this can be the case if the config is injected with an reference of an pointer,
// or it is an array, a map[string]string or an interface{}.
// if this was happens, this function will return the reason why the parsing failed.
func (l *Linter) HaveParsingError() (string, bool) {
	if l.structhandler.IgnoredBecauseOf != "" {
		return l.structhandler.IgnoredBecauseOf, true
	}
	return "", false
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

			addToken := NewMatchToken(l.structhandler, &l.lMap, line, changeNr4Add, sequenceNr, true)
			temporaryChunk.Added = append(temporaryChunk.Added, &addToken)
			needToBeAdded = true
		}
		for _, line := range c.Deleted {
			changeNr4Rm++
			//fmt.Println("DELETED:"+line, " --->index[", indexNr, "] seq[", sequenceNr, "]", "chunk[", chunkIndex, "]", "change[", changeNr4Rm, "]")

			rmToken := NewMatchToken(l.structhandler, &l.lMap, line, changeNr4Rm, sequenceNr, false)
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

// find the token that is the pair of the current token
// so it must be deleted in the previous chunk with the same sequence number
func (l *Linter) findPairsHelper(tkn *MatchToken) {
	bestmatchTokens := l.lMap.GetTokensFromSequenceAndIndex(tkn.SequenceNr, tkn.IndexNr)
	if len(bestmatchTokens) > 0 {
		for _, bestmatch := range bestmatchTokens {
			if bestmatch.Added != tkn.Added {
				tkn.IsPair(bestmatch) // {bestmatch, tkn}
			}
		}
	}
}

// findPairs is a worker that is called if a diff is found.
// it will find the pairs of the diff chunks.
// a pair is a removed and an added line that are the same.
func (l *Linter) findPairs() {
	if l.diffFound {
		// depends on the reult of the diff, we need to find the pairs
		// in the diff chunks. and they have the matching part in the previous diff chunk.
		// so we need to find the matching part in the previous chunk.

		for _, chunk := range l.lMap.Chunks {
			for _, add := range chunk.Added {
				l.findPairsHelper(add)
			}
			// the deletes
			for _, rm := range chunk.Removed {
				l.findPairsHelper(rm)
			}

		}
	}
}

// verify the values of the tokens.
// if the values are not the same, we have an issue.
// this is done after all pairs are found.
// so we can verify the values of the pairs.
// also we detect the highest issue level.
func (l *Linter) valueVerify() {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			token.VerifyValue()
			if added {
				if l.highestIssueLevel < token.Status {
					l.highestIssueLevel = token.Status
				}
			} else {
				// no pair on the removed side.
				// so we have an entry that is not defined in the struct.
				if token.PairToken == nil {
					token.Status = UnknownEntry
				}
				if l.highestIssueLevel < token.Status {
					l.highestIssueLevel = token.Status
				}
			}

		})
	}

}

// walkAll will walk over all tokens in the diff they have equal or higher level as the given level.
func (l *Linter) ReportDiffStartedAt(level int, reportFn func(token *MatchToken)) {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if token.Status >= level {
				reportFn(token)
			}
		})
	}
}

// PrintIssues will print the issues found in the diff.
// This is an report function.
func (l *Linter) PrintIssues() string {
	outPut := ""
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			outPut += token.ToIssueString()
		})
	}
	return outPut
}

func (l *Linter) WalkIssues(hndlFn func(token *MatchToken, added bool)) {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if token.Status > 0 {
				hndlFn(token, added)
			}
		})
	}
}

// proxy to walk all
func (l *Linter) walkAll(hndl func(token *MatchToken, added bool)) {
	l.lMap.walkAll(hndl)
}
