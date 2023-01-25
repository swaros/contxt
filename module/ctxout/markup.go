package ctxout

import (
	"regexp"
	"strings"
)

type Markup struct {
	startToken   rune
	endToken     rune
	closingToken rune
}

type MarkupParser struct {
	HandleErrors bool          // flag if we stop on errors
	Entries      []MarkupEntry // contains all found markups
	LeftString   string        // at least the part of the string, until the first markup
}

type MarkupEntry struct {
	Text       string
	Properties []Markup
	Parsed     string
}

type Parsed struct {
	IsMarkup bool // flag if is plain text or a markup
	Text     string
}

func NewMarkup() *Markup {
	return &Markup{
		startToken:   '<',
		endToken:     '>',
		closingToken: '/',
	}
}

func (m *Markup) SetStartToken(token rune) {
	m.startToken = token
}

func (m *Markup) SetEndToken(token rune) {
	m.endToken = token
}

func (m *Markup) Parse(orig string) []Parsed {
	var pars []Parsed
	//var parsed MarkupParser
	searchString := orig                               // we need a copy of the origin string, so we can cut them after any search hit
	if markups, found := m.splitByMarks(orig); found { // first extract the markups, and iterate over them, if we found some
		for _, markup := range markups { // iterate over all markups
			// we ignore empty markups
			if markup != "" {
				strs := strings.Split(searchString, markup) // split the string by the markup
				if len(strs) > 0 {                          // if we have a part before the markup
					if strs[0] != "" { // if the part before the markup is not empty
						pars = append(pars, Parsed{IsMarkup: false, Text: strs[0]}) // add the part before the markup
					}

					pars = append(pars, Parsed{IsMarkup: true, Text: markup}) // add the markup
					searchString = strs[1]                                    // set the new search string to the part after the markup
				}
			}

		}
		if searchString != "" { // if we have a part after the last markup
			pars = append(pars, Parsed{IsMarkup: false, Text: searchString}) // add the part after the last markup
		}
	}
	return pars
}

func (m *Markup) splitByMarks(orig string) ([]string, bool) {
	var result []string
	found := false
	cmpStr := string(m.startToken) + "[^" + string(m.endToken) + "]+" + string(m.endToken) // compose the regex
	re := regexp.MustCompile(cmpStr)                                                       // compile and panic on bad things happens
	newStrs := re.FindAllString(orig, -1)                                                  // use regex to find all patterns
	for _, s := range newStrs {                                                            // get all markups
		found = true
		result = append(result, s)

	}

	return result, found
}
