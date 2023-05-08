package ctxout

import (
	"regexp"
	"strconv"
	"strings"
)

// general markup parser. this defines a single markup
// and NOT the markup itself
// so a markup is created by openeing the markup with START-TOKEN + something + END-TOKEN
// and closing it with START-TOKEN + CLOSE-IDENT + something + END-TOKEN
// the default markup is <something> and </something>
type Markup struct {
	startToken rune // the identifier for starting a markup
	endToken   rune // the identifier for ending a markup
	closeIdent rune // the identifier for closing a markup
}

// Parsed is a single parsed markup. it can be a plain text or a markup
type Parsed struct {
	IsMarkup bool   // flag if is plain text or a markup
	Text     string // the text of the parsed markup. this is on a markup like <something> the something. on a plain text it is the plain text
}

// NewMarkup creates a new markup parser with the default markup
func NewMarkup() *Markup {
	return &Markup{
		startToken: '<',
		endToken:   '>',
		closeIdent: '/', // the identifier for closing a markup
	}
}

// GetMarkupIntValue returns the int value of a markup
// for example if the markup is <something value='123'> and the key is value, it returns 123
func (p *Parsed) GetMarkupIntValue(key string) (int, bool) {
	var result int
	found := false
	cmpStr := " " + key + `='(\d+)'` // compose the regex
	re := regexp.MustCompile(cmpStr)
	newStrs := re.FindAllStringSubmatch(p.Text, -1)
	for _, s := range newStrs {
		result = p.toInt(s[1])
		found = true
	}
	return result, found
}

// GetMarkupStringValue returns the string value of a markup
// for example if the markup is <something value='123'> and the key is value, it returns 123
func (p *Parsed) GetMarkupStringValue(key string) (string, bool) {
	var result string
	found := false
	cmpStr := " " + key + `='([^']*)'` // compose the regex
	re := regexp.MustCompile(cmpStr)
	newStrs := re.FindAllStringSubmatch(p.Text, -1)
	for _, s := range newStrs {
		result = s[1]
		found = true
	}
	return result, found
}

// toInt converts a string to an int
func (p *Parsed) toInt(s string) int {
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

// GetProperty returns the value of a property of a markup
func (p *Parsed) GetProperty(propertie string, defaultValue interface{}) interface{} {
	if strings.Contains(p.Text, propertie) {
		switch defaultValue.(type) {
		case int:
			if v, f := p.GetMarkupIntValue(propertie); f {
				return v
			} else {
				return defaultValue
			}
		case string:
			if v, f := p.GetMarkupStringValue(propertie); f {
				return v
			} else {
				return defaultValue
			}
		default:
			return defaultValue
		}
	} else {
		return defaultValue
	}
}

// SetStartToken sets the start token of a markup
// like < so the markup would looks like <something/>
func (m *Markup) SetStartToken(token rune) *Markup {
	m.startToken = token
	return m
}

// SetEndToken sets the end token of a markup
// like > so the markup would looks like <something/>
func (m *Markup) SetEndToken(token rune) *Markup {
	m.endToken = token
	return m
}

// SetCloseIdent sets the close identifier of a markup
// like / so the markup would looks like <something/>
func (m *Markup) SetCloseIdent(token rune) *Markup {
	m.closeIdent = token
	return m
}

// Parse parses a string and returns a slice of Parsed elements
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

// getStag returns the start tag of a markup
// for example if the startToken is < and the token is something, it returns <something
func (m *Markup) getStartToken(token string) string {
	return string(m.startToken) + token
}

// getEndToken returns the end tag of a markup
// for example if the startToken is < and the token is something, it returns </something
func (m *Markup) getEndToken(token string) string {
	return string(m.startToken) + string(m.closeIdent) + token
}

// BuildInnerSlice builds a slice of Parsed elements from a slice of Parsed elements
// it searches for the outerMarkup and returns the inner slice of Parsed elements
// and the outer slice of Parsed elements
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

// BuildInnerSliceEach builds a slice of Parsed elements from a slice of Parsed elements
// it searches for the outerMarkup and returns the inner slice of Parsed elements
// and the outer slice of Parsed elements
// it calls the handleFn for each inner slice of Parsed elements
// if the handleFn returns false, the iteration stops
func (m *Markup) BuildInnerSliceEach(parsed []Parsed, outerMarkup string, handleFn func(markup []Parsed) bool) []Parsed {
	var result []Parsed                    // the result
	inInnerBlock := false                  // flag if we are in the inner block
	startM := m.getStartToken(outerMarkup) // the start token of the outer markup
	endM := m.getEndToken(outerMarkup)     // the end token of the outer markup
	for _, p := range parsed {             // iterate over all parsed elements
		if p.IsMarkup { // if the element is a markup
			if strings.HasPrefix(p.Text, endM) { // if we hit the end token of the outer markup
				inInnerBlock = false                      // we are not in the inner block anymore
				if handleFn != nil && !handleFn(result) { // if the handleFn returns false, we stop the iteration
					return result // return the result
				}
				result = []Parsed{} // reset the result after each handle
			}
			if inInnerBlock { // if we are in the inner block
				result = append(result, p) // add the element to the result
			} else {
				inInnerBlock = strings.HasPrefix(p.Text, startM) // if we hit the start token of the outer markup, we are in the inner block
			}
		} else {
			if inInnerBlock { // if we are in the inner block
				result = append(result, p) // add the element to the result
			}
		}
	}
	if handleFn != nil { // if we have a handleFn
		handleFn(result) // call it with the last result
	}
	return result
}

// splitByMarks splits a string by the markups
// it returns a slice of strings and a bool if it found any markups
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
