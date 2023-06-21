package yaclint

import (
	"strings"

	"github.com/google/uuid"
)

const (
	Unset                    = -1 // the value is not set. the initial value. nothing was tryed to match
	PerfectMatch             = 0  // the value and type matches. should most times not happen, because we compare the default values with actual values. so a diff is common
	ValueMatchButTypeDiffers = 1  // the value matches but in different type like "1.4" and 1.4 (valid because of yaml parser type conversion)
	ValueNotMatch            = 2  // the value is not matching (still valid because of yaml parser type conversion. we compare the default values with actual values. so a diff is common)

	// now the  types they are mostly real issues in the config (they should be greater then 9)
	// the default issue Errorlevel is 10. so we can use the default errorlevel for the most common issues
	IssueLevelError = 10

	MissingEntry = 10 // the entry is missing. this entry is defined in struct but not in config. als no omitempty tag is set in struct
	WrongType    = 11 // the type is wrong. different from the strct definition, and also no type conversion is possible
	UnknownEntry = 12 // the entry is is in the config but not in the struct

)

type MatchToken struct {
	UuId       string
	KeyWord    string
	Value      interface{}
	Type       string
	Added      bool
	SequenceNr int
	indexNr    int
	Status     int
	PairToken  *MatchToken
	ParentLint *LintMap
}

func NewMatchToken(parent *LintMap, line string, indexNr int, seqNr int, added bool) MatchToken {
	var matchToken MatchToken
	matchToken.ParentLint = parent
	matchToken.UuId = uuid.New().String()
	matchToken.Type = "undefined"
	matchToken.Added = added
	matchToken.SequenceNr = seqNr
	matchToken.indexNr = indexNr

	matchToken.Status = -1
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

// IsPair checks if the given token is a pair to this token
// it checks if the keyword is the same and the added flag is different
// if so it sets the pair token property and returns true
func (m *MatchToken) IsPair(token *MatchToken) bool {
	if m.IsValid() && token.IsValid() && m.KeyWord == token.KeyWord && m.Added != token.Added {
		m.PairToken = token
		return true
	}
	return false
}

// VerifyValue checks if the value is matching the pair token
// if so it sets the status property and returns the status
// the status represents the issue level
func (m *MatchToken) VerifyValue() int {
	if m.Status != -1 {
		return m.Status
	}
	m.detectValueType()
	if m.PairToken == nil {
		m.Status = MissingEntry
	} else {
		pairMatch := m.PairToken
		if pairMatch == nil {
			m.Status = MissingEntry
		} else {
			if m.Type == pairMatch.Type {
				if m.Value == pairMatch.Value {
					m.Status = PerfectMatch
				} else {
					m.Status = ValueNotMatch
				}
			} else {
				m.Status = WrongType
			}
		}
	}

	return m.Status
}

// IsValid checks if the token is valid and can be used for further processing
func (m *MatchToken) IsValid() bool {
	if m.KeyWord != "" && m.Type != "" {
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
