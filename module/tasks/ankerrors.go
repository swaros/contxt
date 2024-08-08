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

package tasks

import (
	"fmt"

	"github.com/mattn/anko/parser"
)

type VerifiedLines struct {
	Line string
	Err  error
}

type AnkVerifier struct {
}

func NewAnkVerifier() *AnkVerifier {
	return &AnkVerifier{}
}

// VerifyLines verifies the given lines of script
// it just uses the parser to get at least syntax errors
func (av *AnkVerifier) VerifyLines(lines []string) ([]VerifiedLines, error) {
	var result []VerifiedLines
	for _, line := range lines {

		stmt, err := parser.ParseSrc(line)
		scr := line
		if stmt != nil {
			scr = fmt.Sprintf("%d:%s", stmt.Position().Column, line)
		}

		result = append(result, VerifiedLines{Line: scr, Err: err})
	}
	return result, nil
}
