package yaclint

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/swaros/contxt/module/yamc"
)

const (
	// -- info level --
	Unset                    = -1 // the value is not set. the initial value. nothing was tryed to match
	PerfectMatch             = 0  // the value and type matches. should most times not happen, because we compare the default values with actual values. so a diff is common
	ValueMatchButTypeDiffers = 1  // the value matches but in different type like "1.4" and 1.4 (valid because of yaml parser type conversion)
	ValueNotMatch            = 2  // the value is not matching (still valid because of yaml parser type conversion. we compare the default values with actual values. so a diff is common)

	// now the  types they are mostly real issues in the config (they should be greater then 5)
	// the default issue Errorlevel is 10. so we can use the default errorlevel for the most common issues

	// -- warning level --
	MissingEntry = 5 // the entry is missing. this entry is defined in struct but not in config. depends on the implementation if this is an issue.
	// -- error level --
	WrongType    = 11 // the type is wrong. different from the strct definition, and also no type conversion is possible
	UnknownEntry = 12 // the entry is is in the config but not in the struct

	IdentObject = 1 // the token is an object
	IdentArray  = 2 // the token is an array
	IdentValue  = 3 // the token is a value

	ValueString     = 0 // the token is not Open or Close Object or Array and not KeyValue
	KeyValue        = 1 // the token is a key value pair like "key: value"
	OpenObject      = 2 // the token is an object like "{"
	CloseObject     = 3 // the token is an object like "}"
	OpenArray       = 4 // the token is an object like "["
	CloseArray      = 5 // the token is an object like "]"
	KeyedOpenObject = 6 // the token is an object like "key: {"
	KeyedOpenArray  = 7 // the token is an object like "key: ["

)

type MatchToken struct {
	KeyWord    string
	OrginKey   string
	KeyPath    string
	Value      interface{}
	Type       string
	Added      bool
	SequenceNr int
	IndexNr    int
	Status     int
	PairToken  *MatchToken
	ParentLint *LintMap
	TraceFunc  func(args ...interface{})
}

func NewMatchToken(structDef yamc.StructDef, traceFn func(args ...interface{}), parent *LintMap, line string, indexNr int, seqNr int, added bool) MatchToken {
	traceFn("NewMatchToken:parse: ", line)
	var matchToken MatchToken
	matchToken.TraceFunc = traceFn
	matchToken.ParentLint = parent
	matchToken.Type = "undefined"
	matchToken.Added = added
	matchToken.SequenceNr = seqNr
	matchToken.IndexNr = indexNr

	matchToken.Status = -1

	rKeyWod, rValue, rWithValue := getTokenParts(line)
	matchToken.Value = rValue
	matchToken.OrginKey = rKeyWod
	matchToken.KeyWord, matchToken.KeyPath = matchToken.getNameOf(structDef, rKeyWod)
	if rWithValue {
		matchToken.detectValueType()
	}
	matchToken.trace("NewMatchToken:", matchToken.ToString())
	return matchToken
}

// we will read the keyname and the value from the line
// they comes as  string. that is "nicely" formated.
// so we have to deal with any sinlgle line and do not have an context.
// so {key:"value"} will be come in 3 lines, and we handle any line as a single line
// we will get as line argument:
// 1. : {
// 2.  key: "value"
// 3. }
// any key have a leading space, also if they are in the root of the struct
// so this one space have to removed anytime
// then we have to take care about the colon. the colon is the separator between key and value
// the value itself can have a comma as separator. but the comma is not part of the value
// so we have to deal with the comma as separator and the colon as separator
func getTokenParts(token string) (string, string, bool) {
	if strings.Contains(token, ":") {
		parts := strings.Split(token, ":")
		// we only remove the ONE leading space. no trim here
		key := strings.TrimPrefix(parts[0], " ")
		value := strings.TrimSpace(parts[1])
		value = strings.TrimSuffix(value, ",")
		// if we have a value, that is equal "{" we found an object
		// so there is no value
		if value == "{" {
			return key, "", false
		}
		// if we have a value, that is equal "[" we found an array
		// so there is no value
		if value == "[" {
			return key, "", false
		}
		return key, value, true
	}
	return token, "", false

}

func isJustAToken(str string) bool {
	typed := detectTypeFromString(str)
	if typed == OpenArray || typed == OpenObject || typed == CloseArray || typed == CloseObject {
		return true
	}
	return false
}

func isObjectOrArray(str string) bool {
	typed := detectTypeFromString(str)
	if typed == OpenObject ||
		typed == OpenArray ||
		typed == KeyedOpenObject ||
		typed == CloseArray ||
		typed == CloseObject ||
		typed == KeyedOpenArray {
		return true
	}
	return false
}

// we will detect the type of the value
// return the type as int flag
// OpenObject, CloseObject, OpenArray, CloseArray, KeyValue, ValueString
func detectTypeFromString(str string) int {
	// first trim the string
	haveSep := strings.Contains(str, ":")
	str = strings.TrimSpace(str)
	if str == "{" {
		return OpenObject
	}
	if str == "}" || str == "}," {
		return CloseObject
	}
	if str == "[" {
		return OpenArray
	}
	if str == "]" || str == "]," {
		return CloseArray
	}

	if haveSep && strings.Contains(str, " {") {
		return KeyedOpenObject
	}
	if haveSep && strings.Contains(str, " [") {
		return KeyedOpenArray
	}

	if haveSep {
		return KeyValue
	}
	return ValueString
}

func (m *MatchToken) trace(args ...interface{}) {
	if m.TraceFunc != nil {
		m.TraceFunc(args...)
	}
}

func (m *MatchToken) getNameOf(structDef yamc.StructDef, check string) (string, string) {
	if structDef.Fields != nil && len(structDef.Fields) > 0 {
		if field, err := structDef.GetField(check); err == nil {
			m.trace("MatchToken.getNameOf: [", check, "] => [", field.Name, "] into [", field.OrginalTag.TagRenamed, "] @", field.Path)
			if field.OrginalTag.TagRenamed == "" {
				return field.Name, field.Path
			}
			return field.OrginalTag.TagRenamed, field.Path
		}
	}
	m.trace("MatchToken:", m, " [", check, "] !No Tag found!")
	return check, check
}

// IsPair checks if the given token is a pair to this token
// it checks if the keyword is the same and the added flag is different
// if so it sets the pair token property and returns true
func (m *MatchToken) IsPair(token *MatchToken) bool {

	keyVerified := false
	// savest way to check if the keypath is the same
	if m.KeyPath != "" && token.KeyPath != "" && m.KeyPath == token.KeyPath {
		keyVerified = true
	} else {
		// if we do not have a keypath we check if the keyword is the same.
		// but then the OriginKey must be the also the same, or we mix up keys in a different path (like "a.b" and "b")
		if m.KeyWord == token.KeyWord && m.OrginKey == token.OrginKey {
			keyVerified = true
		}
	}
	if keyVerified && m.Added != token.Added {
		m.PairToken = token
		m.trace("MatchToken:", m, " [", m.keyToString(), "] is pair to [", token.keyToString(), "] index ", token.IndexNr, " sequence ", token.SequenceNr)
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
	//m.detectValueType()
	if m.PairToken == nil {
		m.Status = MissingEntry
	} else {
		pairMatch := m.PairToken
		if pairMatch == nil {
			m.Status = MissingEntry
		} else {
			// values matching are difficult, because of the type conversion of the yaml parser
			if m.Value == pairMatch.Value {
				m.Status = PerfectMatch
			} else {
				// if the value is a string and the pair token is a number, we try to convert the string to a number
				// so we do the lazy way and convert all values to string and compare them
				mStr := fmt.Sprintf("%v", m.Value)
				pairStr := fmt.Sprintf("%v", pairMatch.Value)
				if mStr == pairStr {
					m.Status = ValueMatchButTypeDiffers
				} else {
					m.Status = ValueNotMatch
				}

			}
		}
	}

	return m.Status
}

func (m *MatchToken) keyToString() string {
	return fmt.Sprintf("%s (%s)", m.KeyWord, m.KeyPath)
}

// ToIssueString returns a string representation of the issue
func (m *MatchToken) ToIssueString() string {

	// compose a readable string
	// depending on the issue level
	// the issue level is the status property
	switch m.Status {
	case ValueMatchButTypeDiffers:
		return fmt.Sprintf(
			"ValueMatchButTypeDiffers: level[%d] @%s ['%s' != '%s']",
			m.Status,
			m.keyToString(),
			m.Type,
			m.PairToken.Type,
		)

	case ValueNotMatch:
		return fmt.Sprintf(
			"ValuesNotMatching: level[%d] @%s vs @%s ['%v' != '%v']",
			m.Status,
			m.keyToString(),
			m.PairToken.keyToString(),
			m.Value,
			m.PairToken.Value,
		)

	case MissingEntry:
		return fmt.Sprintf(
			"MissingEntry: level[%d] @%s",
			m.Status,
			m.keyToString(),
		)

	case WrongType:
		return fmt.Sprintf(
			"WrongType: level[%d] @%s ['%s' != '%s']",
			m.Status,
			m.keyToString(),
			m.Type,
			m.PairToken.Type,
		)

	case UnknownEntry:
		return fmt.Sprintf(
			"UnknownEntry: level[%d] @%s",
			m.Status,
			m.keyToString(),
		)

	case PerfectMatch:
		return fmt.Sprintf(
			"PerfectMatch: level[%d] @%s",
			m.Status,
			m.keyToString(),
		)

	default:
		return fmt.Sprintf("Unknown: level[%d] @%s", m.Status, m.keyToString())

	}
}

func (m *MatchToken) ToString() string {
	addStr := "[-]"
	if m.Added {
		addStr = "[+]"
	}
	if m.PairToken == nil {
		return fmt.Sprintf("%s %s: [%d] val[%v] indx[%d] seq[%d] (%s)",
			addStr,
			m.keyToString(),
			m.Status,
			m.Value,
			m.IndexNr,
			m.SequenceNr,
			m.Type)
	}
	return fmt.Sprintf("%s %s: [%d] val[%v] (%s)pval[%v] indx[%d] seq[%d] (%s)",
		addStr,
		m.keyToString(),
		m.Status,
		m.Value,
		m.PairToken.keyToString(),
		m.PairToken.Value,
		m.IndexNr,
		m.SequenceNr,
		m.Type)
}

func (m *MatchToken) trimString() {
	escaped := strings.Replace(m.Value.(string), "\"", "", -1)
	m.Value = strings.Trim(escaped, " ")

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
	case string: // this is the case most of the time, because we do not parse the data. we parse the diff report
		m.Value = DetectedValueFromString(m.Value.(string))
	}

	// and again after the conversion
	switch m.Value.(type) {
	case string:
		m.trimString()
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

func DetectedValueFromString(str string) interface{} {
	// we keep the quotes if the string contains quotes. to get the right value, CleanValue() should be used
	if strings.Contains(str, "\"") {
		return str
	}

	str = strings.TrimLeft(str, " ")
	if str == "true" {
		return true
	}
	if str == "false" {
		return false
	}
	if i, err := strconv.Atoi(str); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return str
}
