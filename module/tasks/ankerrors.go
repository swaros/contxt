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
