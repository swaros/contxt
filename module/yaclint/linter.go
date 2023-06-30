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
	"reflect"
	"strings"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

const (
	IssueLevelInfo  = 2
	IssueLevelWarn  = 5
	IssueLevelError = 10
)

type Linter struct {
	config            *yacl.ConfigModel // the config model that we need to verify
	lMap              LintMap           // contains the diff chunks
	diffFound         bool              // true if we found a diff. that is just a sign, that an SOME diff is found, not that the config is invalid
	highestIssueLevel int               // the highest issue level found
	structhandler     yamc.StructDef    // the struct handler for the config file. keeps the struct definition
	ldlogger          DirtyLoggerDef    // quick and dirty logger for the linter
}

func NewLinter(config yacl.ConfigModel) *Linter {
	return &Linter{
		config:            &config,
		highestIssueLevel: 0,
	}

}

// getUnstructMap loads the config file as generic map and returns it
// as map[string]interface{} and as string (the yaml/json representation)
func (l *Linter) getUnstructMap(loader yamc.DataReader) (map[string]interface{}, error) {
	fileName := l.config.GetLoadedFile() // the file name of the config file
	l.Trace("(re)Loading file:", fileName)
	m := make(map[string]interface{}) // generic map to load the file for comparison
	err := loader.FileDecode(fileName, &m)
	if err != nil {
		return nil, err
	}
	_, err = loader.Marshal(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// getStructSource creates the yaml/json representation of the config file
func (l *Linter) getStructSource(loader yamc.DataReader) (map[string]interface{}, error) {
	cYamc, cerr := l.config.GetAsYmac() // get the configuration as yamc object
	if cerr != nil {
		return nil, cerr
	}

	structData := cYamc.GetData() // get the source as map[string]interface{}
	return structData, nil
}

// init4read is a helper function that initializes the linter for reading the config file.
func (l *Linter) init4read() (string, string, error) {
	yamcLoader := l.config.GetLastUsedReader() // the last used reader from the config
	if yamcLoader == nil {
		return "", "", fmt.Errorf("no reader found. the config needs to be loaded first")
	}
	l.Trace("found used data reader: ", reflect.TypeOf(yamcLoader))
	l.structhandler = *yamcLoader.GetFields() // get the struct handler from the reader. must be done before the unstructed map is loaded

	l.Trace("init4read: structhandler Init: ", l.structhandler.Init)

	unStructData, err1 := l.getUnstructMap(yamcLoader)
	if err1 != nil {
		return "", "", err1
	}

	structData, err2 := l.getStructSource(yamcLoader)
	if err2 != nil {
		return "", "", err2
	}

	niceUnstructed := pretty.CompareConfig.Sprint(unStructData)
	niceStructed := pretty.CompareConfig.Sprint(structData)
	return niceUnstructed, niceStructed, nil
}

// GetDiff returns the diff between the config file and the structed config file.
// The diff is returned as string.
func (l *Linter) GetDiff() (string, error) {
	unstructedSrc, structedSrc, err := l.init4read()
	if err != nil {
		return "", err
	}
	return diff.Diff(unstructedSrc, structedSrc), nil

}

// Verify is the main function of the linter. It will verify the config file
// against the structed config file. It will return an error if the config file
// is not valid.
func (l *Linter) Verify() error {
	unstructSource, structSource, err := l.init4read()
	l.Trace("Verify: unstructSource:\n", unstructSource)
	l.Trace("Verify: structSource:\n", structSource)
	if err != nil {
		return err
	}

	freeChnk := strings.Split(unstructSource, "\n")
	orgiChnk := strings.Split(structSource, "\n")

	chunk := diff.DiffChunks(freeChnk, orgiChnk)
	l.Trace("Verify: chunk count:", len(chunk))
	l.chunkWorker(chunk)
	l.Trace("Verify: diff found:", l.diffFound)
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

// report if we found an issue with the config file, that should not be ignored.
func (l *Linter) HasError() bool {
	return l.highestIssueLevel >= IssueLevelError
}

// report if we found an issue with the config file, that is not so important, but be warned.
func (l *Linter) HasWarning() bool {
	return l.highestIssueLevel >= IssueLevelWarn
}

// report if we found an issue with the config file, that is most usual, like type conversion. (what is difficult to avoid)
func (l *Linter) HasInfo() bool {
	return l.highestIssueLevel >= IssueLevelInfo
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
	l.Trace("chunkWorker: chunk count:", len(chunks))

	// we need to remember all the keys that are found in the config file.
	// so we can build a ordered list of keys.
	// this is needed to get the full path of the key. including the parent keys.
	var keysAdded []string
	var keysRemoved []string

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

	l.structhandler.SetAllowedTagSearch(true)
	// iterate over all chunks.
	for chunkIndex, c := range chunks {
		temporaryChunk := LintChunk{}
		needToBeAdded := false
		if len(keysAdded) > 0 {
			l.Trace("chunkWorker: -- index -- add - (", len(keysAdded), ")", keysAdded)
		}

		if len(keysRemoved) > 0 {
			l.Trace("chunkWorker: -- index -- rm - (", len(keysRemoved), ")", keysRemoved)
		}

		for _, line := range c.Added {
			// ignore any single open or close bracket
			if isDelimerType(line) != ValueString {
				continue
			}

			keyStr, _, _ := getTokenParts(line)
			keysAdded = append(keysAdded, keyStr)

			l.structhandler.SetIndexSlice(keysAdded)

			changeNr4Add++
			addToken := NewMatchToken(l.structhandler, l.Trace, &l.lMap, line, changeNr4Add, sequenceNr, true)
			temporaryChunk.Added = append(temporaryChunk.Added, &addToken)
			needToBeAdded = true
		}
		for _, line := range c.Deleted {
			// ignore any single open or close bracket
			if isDelimerType(line) != ValueString {
				continue
			}
			keyStr, _, _ := getTokenParts(line)
			keysRemoved = append(keysRemoved, keyStr)

			l.structhandler.SetIndexSlice(keysRemoved)

			changeNr4Rm++
			//fmt.Println("DELETED:"+line, " --->index[", indexNr, "] seq[", sequenceNr, "]", "chunk[", chunkIndex, "]", "change[", changeNr4Rm, "]")

			rmToken := NewMatchToken(l.structhandler, l.Trace, &l.lMap, line, changeNr4Rm, sequenceNr, false)
			rmToken.TraceFunc = l.Trace
			temporaryChunk.Removed = append(temporaryChunk.Removed, &rmToken)
			needToBeAdded = true
		}

		// on equal we need to reset the indices for removed and added changes, so we can track the
		// changes in the next chunk, and get the context what of these removas and adds are just changes
		if len(c.Equal) > 0 {
			changeNr4Add = 0
			changeNr4Rm = 0
			sequenceNr += len(c.Equal)
			// we need to add the equal lines to the keys, so we can track the context of the keys
			// this needs to be done for booth. added and removed keys.
			// ignore any single open or close bracket
			for _, line := range c.Equal {
				if isDelimerType(line) != ValueString {
					continue
				}
				keyStr, _, _ := getTokenParts(line)
				keysAdded = append(keysAdded, keyStr)
				keysRemoved = append(keysRemoved, keyStr)
			}

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
	// first we try to find the pair in the best case. that means, we do not have a match depending the keyword.
	// we also are in the same chunk, so we be sure, this keyword is the pair.
	bestmatchTokens := l.lMap.GetTokensFromSequenceAndIndex(tkn.SequenceNr, tkn.IndexNr)
	l.findPairInPairMap(tkn, bestmatchTokens, "findPairsHelper: try optimal case. compare ")

	// we did not find a pair in the best case, so we need to find a pair in the worst case.
	// that means we have to seach in all other chunks, if we find a pair.
	// that can be lead to false positives, but we can not do anything else.
	if tkn.PairToken == nil {
		// for the trace, we try so intent the output, so it is more readable
		//l.Trace("findPairsHelper:    try fallback", tkn, " in best case search, try to find a pair without sequence number")
		moreTokens := l.lMap.GetTokensFromSequence(tkn.SequenceNr)
		l.findPairInPairMap(tkn, moreTokens, "findPairsHelper:    retry with sequence only. ")

	}
}

func (l *Linter) findPairInPairMap(tkn *MatchToken, tkns []*MatchToken, traceMsg string) {
	if len(tkns) > 0 {
		for _, bestmatch := range tkns {
			if bestmatch.Added != tkn.Added {
				match := tkn.IsPair(bestmatch) // {bestmatch, tkn}
				if traceMsg != "" {
					l.Trace(traceMsg, tkn, " and ", bestmatch, " (", match, ")")
				}

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

// GetIssue will execute the reportFn for all tokens that have the same or higher level as the given level.
func (l *Linter) GetIssue(level int, reportFn func(token *MatchToken)) {
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if token.Status >= level {
				reportFn(token)
			}
		})
	}
}

func (l *Linter) Errors() []string {
	return l.filterIssueBylevel(IssueLevelError)
}

func (l *Linter) Warnings() []string {
	return l.filterIssueBylevel(IssueLevelWarn)
}

func (l *Linter) Infos() []string {
	return l.filterIssueBylevel(IssueLevelInfo)
}

func (l *Linter) filterIssueBylevel(equalOrHigherLevel int) []string {
	var out []string
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if token.Status >= equalOrHigherLevel {
				add := "[-]"
				if added {
					add = "[+]"
				}
				out = append(out, add+token.ToIssueString())
			}
		})
	}
	return out
}

// PrintIssues will print the issues found in the diff.
// This is an report function and show any issue greater than 0.
func (l *Linter) PrintIssues() string {
	outPut := ""
	if l.diffFound {
		l.walkAll(func(token *MatchToken, added bool) {
			if token.Status > 0 {
				add := "[-]"
				if added {
					add = "[+]"
				}
				outPut += add + token.ToIssueString() + "\n"
			}
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

func (l *Linter) SetDirtyLogger(logger DirtyLoggerDef) {
	l.ldlogger = logger
}

// Trace is a helper function to trace the linter workflow.
// this might help to debug the linter.
func (l *Linter) Trace(arg ...interface{}) {
	if l.ldlogger != nil {
		l.ldlogger.Trace(arg...)
	}
}

func (l *Linter) GetTrace(orFind ...string) string {
	if l.ldlogger != nil {
		return strings.Join(l.ldlogger.GetTrace(orFind...), "\n")
	}
	return ""
}
