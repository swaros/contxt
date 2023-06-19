package yaclint

import (
	"strings"

	"github.com/google/uuid"
)

const (
	Unset                    = -1 // the value is not set
	PerfectMatch             = 0  // the value and type matches
	ValueMatchButTypeDiffers = 1  // the value matches butin different type like "1.4" and 1.4 (valid because of yaml parser type conversion)
	MissingEntry             = 2  // the entry is missing. this entry is defined in struct but not in config. als no omitempty tag is set in struct
	UnknownEntry             = 3  // the entry is is in the config but not in the struct
	WrongType                = 4  // the type is wrong. different from the strct definition, and also no type conversion is possible
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

func (m *MatchToken) IsPair(token *MatchToken) bool {
	if m.KeyWord == token.KeyWord && m.Added != token.Added {
		m.PairToken = token
		return true
	}
	return false
}

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
					m.Status = ValueMatchButTypeDiffers
				}
			} else {
				m.Status = WrongType
			}
		}
	}

	return m.Status
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
