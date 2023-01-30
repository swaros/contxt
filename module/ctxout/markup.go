package ctxout

import (
	"regexp"
	"strconv"
	"strings"
)

type Markup struct {
	startToken rune
	endToken   rune
	closeIdent string // the identifier for closing a markup
}

type Parsed struct {
	IsMarkup bool // flag if is plain text or a markup
	Text     string
}

func NewMarkup() *Markup {
	return &Markup{
		startToken: '<',
		endToken:   '>',
		closeIdent: "/", // the identifier for closing a markup
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
				strs := strings.SplitN(searchString, markup, 2) // split the string by the markup
				if len(strs) > 0 {                              // if we have a part before the markup
					if strs[0] != "" { // if the part before the markup is not empty
						pars = append(pars, Parsed{IsMarkup: false, Text: strs[0]}) // add the part before the markup
					}

					pars = append(pars, Parsed{IsMarkup: true, Text: markup}) // add the markup
					searchString = strings.Join(strs[1:], "")                 // set the new search string to the part after the markup
				}
			}

		}
		if searchString != "" { // if we have a part after the last markup
			pars = append(pars, Parsed{IsMarkup: false, Text: searchString}) // add the part after the last markup
		}
	}
	return pars
}

func (m *Markup) getStartToken(token string) string {
	return string(m.startToken) + token
}

func (m *Markup) getEndToken(token string) string {
	return string(m.startToken) + m.closeIdent + token
}

func (m *Markup) BuildInnerSlice(parsed []Parsed, outerMarkup string) ([]Parsed, []Parsed) {
	var result []Parsed
	var outer []Parsed
	inInnerBlock := false
	startM := m.getStartToken(outerMarkup)
	endM := m.getEndToken(outerMarkup)
	for _, p := range parsed {
		if p.IsMarkup {
			if strings.HasPrefix(p.Text, endM) {
				inInnerBlock = false
				outer = append(outer, p)
				return result, outer
			}
			if inInnerBlock {
				result = append(result, p)
			} else {
				inInnerBlock = strings.HasPrefix(p.Text, startM)
				if inInnerBlock {
					outer = append(outer, p) // if we hit this once, we have a outer markup
				}
			}
		} else {
			if inInnerBlock {
				result = append(result, p)
			}
		}
	}
	return result, outer
}

func (m *Markup) BuildInnerSliceEach(parsed []Parsed, outerMarkup string, handleFn func(markup []Parsed) bool) []Parsed {
	var result []Parsed
	inInnerBlock := false
	startM := m.getStartToken(outerMarkup)
	endM := m.getEndToken(outerMarkup)
	for _, p := range parsed {
		if p.IsMarkup {
			if strings.HasPrefix(p.Text, endM) {
				inInnerBlock = false
				if handleFn != nil && !handleFn(result) { // if the handleFn returns false, we stop the iteration
					return result
				}
				result = []Parsed{} // reset the result after each handle
			}
			if inInnerBlock {
				result = append(result, p)
			} else {
				inInnerBlock = strings.HasPrefix(p.Text, startM)
			}
		} else {
			if inInnerBlock {
				result = append(result, p)
			}
		}
	}
	if handleFn != nil {
		handleFn(result)
	}
	return result
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

func (m *Markup) GetMarkupIntValue(markup string, key string) int {
	var result int
	cmpStr := key + `='(\d+)'` // compose the regex
	re := regexp.MustCompile(cmpStr)
	newStrs := re.FindAllStringSubmatch(markup, -1)
	for _, s := range newStrs {
		result = m.toInt(s[1])
	}
	return result
}

func (m *Markup) GetMarkupStringValue(markup string, key string) string {
	var result string
	cmpStr := key + `='([^']*)'` // compose the regex
	re := regexp.MustCompile(cmpStr)
	newStrs := re.FindAllStringSubmatch(markup, -1)
	for _, s := range newStrs {
		result = s[1]
	}
	return result
}

func (m *Markup) toInt(s string) int {
	var result int
	if s != "" {
		var e error
		result, e = strconv.Atoi(s)
		if e != nil {
			result = 0
		}
	}
	return result
}
