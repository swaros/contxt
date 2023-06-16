package yaclint

import "strings"

type MatchToken struct {
	KeyWord   string
	Value     interface{}
	Type      string
	Added     bool
	SeqneceNr int
}

func NewMatchToken(line string, SeqneceNr int, added bool) MatchToken {
	var matchToken MatchToken
	matchToken.Type = "undefined"
	matchToken.Added = added
	matchToken.SeqneceNr = SeqneceNr
	jsonLineParts := strings.Split(line, ":")
	if len(jsonLineParts) > 1 {
		matchToken.KeyWord = jsonLineParts[0]
		matchToken.Value = jsonLineParts[1]
		matchToken.detectValueType()
	} else {
		matchToken.KeyWord = line
		matchToken.Value = ""

	}
	return matchToken
}

func (m *MatchToken) Compare(token MatchToken) bool {
	if m.KeyWord == token.KeyWord && m.Type == token.Type {
		return true
	}
	return false
}

func (m *MatchToken) CompareValue(token MatchToken) bool {
	if m.KeyWord == token.KeyWord && m.Type == token.Type && m.Value == token.Value {
		return true
	}
	return false
}

func (m *MatchToken) IsValid() bool {
	if m.KeyWord != "" && m.Type != "" {
		return true
	}
	return false
}

func (m *MatchToken) IsCounterPart(token *MatchToken) bool {
	if token.IsValid() && m.KeyWord == token.KeyWord && m.Type == token.Type && m.Added != token.Added {
		return true
	}
	return false
}

func (m *MatchToken) detectValueType() {
	switch m.Value.(type) {
	case string:
		m.Type = "string"
	case int:
		m.Type = "int"
	case bool:
		m.Type = "bool"
	case float64:
		m.Type = "float64"
	default:
		m.Type = "undefined"
	}
}
