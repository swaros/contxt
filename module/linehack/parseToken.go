package linehack

import (
	"regexp"
	"strconv"
)

type scanToken struct {
	Pos   int
	Value string
}

type tokenSelfProvider interface {
	itsMe(token scanToken) bool
	setValue(token scanToken)
}

type tokenCouldBeAnother interface {
	maybeWrong() bool
	wrongValue() string
}

type trashToken struct {
	Pos   int
	Value string
}

func (t *trashToken) itsMe(token scanToken) bool {
	return false
}

func (t *trashToken) setValue(token scanToken) {
	t.Pos = token.Pos
	t.Value = token.Value
}

type If struct {
	Pos int
}

func (i *If) itsMe(token scanToken) bool {
	return token.Value == "if"
}

func (i *If) setValue(token scanToken) {
	i.Pos = token.Pos
}

type tBracketOpen struct {
	Pos int
}

func (t *tBracketOpen) itsMe(token scanToken) bool {
	return token.Value == "("
}

func (t *tBracketOpen) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tBracketClose struct {
	Pos int
}

func (t *tBracketClose) itsMe(token scanToken) bool {
	return token.Value == ")"
}

func (t *tBracketClose) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tCurlyOpen struct {
	Pos int
}

func (t *tCurlyOpen) itsMe(token scanToken) bool {
	return token.Value == "{"
}

func (t *tCurlyOpen) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tCurlyClose struct {
	Pos int
}

func (t *tCurlyClose) itsMe(token scanToken) bool {
	return token.Value == "}"
}

func (t *tCurlyClose) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tSemiColon struct {
	Pos int
}

func (t *tSemiColon) itsMe(token scanToken) bool {
	return token.Value == ";"
}

func (t *tSemiColon) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tThen struct {
	Pos int
}

func (t *tThen) itsMe(token scanToken) bool {
	return token.Value == "then"
}

func (t *tThen) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tElse struct {
	Pos int
}

func (t *tElse) itsMe(token scanToken) bool {
	return token.Value == "else"
}

func (t *tElse) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tSet struct {
	Pos int
}

func (t *tSet) itsMe(token scanToken) bool {
	return token.Value == "set"
}

func (t *tSet) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tString struct {
	Value string
	Pos   int
}

func (t *tString) itsMe(token scanToken) bool {
	return token.Value[0] == '"' && token.Value[len(token.Value)-1] == '"'
}

func (t *tString) setValue(token scanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

type tInt struct {
	Value int
	Pos   int
}

func (t *tInt) itsMe(token scanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[0-9]+$`).MatchString(token.Value)
}

func (t *tInt) setValue(token scanToken) {
	t.Pos = token.Pos
	if i, err := strconv.Atoi(token.Value); err == nil {
		t.Value = i
	}
}

type tFloat struct {
	Value float64
	Pos   int
}

func (t *tFloat) itsMe(token scanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[0-9]+\.[0-9]+$`).MatchString(token.Value)
}

func (t *tFloat) setValue(token scanToken) {
	t.Pos = token.Pos
	if i, err := strconv.ParseFloat(token.Value, 64); err == nil {
		t.Value = i
	}
}

type tBool struct {
	Value bool
	Pos   int
}

func (t *tBool) itsMe(token scanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && (token.Value == "true" || token.Value == "false")
}

func (t *tBool) setValue(token scanToken) {
	t.Pos = token.Pos
	if token.Value == "true" {
		t.Value = true
	} else {
		t.Value = false
	}
}

type tPrint struct {
	Pos int
}

func (t *tPrint) itsMe(token scanToken) bool {
	return token.Value == "print"
}

func (t *tPrint) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tVar struct {
	Value string
	Pos   int
}

func (t *tVar) itsMe(token scanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value)
}

func (t *tVar) setValue(token scanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

type tPlus struct {
	Pos int
}

func (t *tPlus) itsMe(token scanToken) bool {
	return token.Value == "+"
}

func (t *tPlus) setValue(token scanToken) {
	t.Pos = token.Pos
}

func (t *tPlus) maybeWrong() bool {
	return true
}

func (t *tPlus) wrongValue() string {
	return "+"
}

type tMinus struct {
	Pos int
}

func (t *tMinus) itsMe(token scanToken) bool {
	return token.Value == "-"
}

func (t *tMinus) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tMultiply struct {
	Pos int
}

func (t *tMultiply) itsMe(token scanToken) bool {
	return token.Value == "*"
}

func (t *tMultiply) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tDivide struct {
	Pos int
}

func (t *tDivide) itsMe(token scanToken) bool {
	return token.Value == "/"
}

func (t *tDivide) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tModulo struct {
	Pos int
}

func (t *tModulo) itsMe(token scanToken) bool {
	return token.Value == "%"
}

func (t *tModulo) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tEqual struct {
	Pos int
}

func (t *tEqual) itsMe(token scanToken) bool {
	return token.Value == "=="
}

func (t *tEqual) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tNotEqual struct {
	Pos int
}

func (t *tNotEqual) itsMe(token scanToken) bool {
	return token.Value == "!="
}

func (t *tNotEqual) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tLess struct {
	Pos int
}

func (t *tLess) itsMe(token scanToken) bool {
	return token.Value == "<"
}

func (t *tLess) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tLessOrEqual struct {
	Pos int
}

func (t *tLessOrEqual) itsMe(token scanToken) bool {
	return token.Value == "<="
}

func (t *tLessOrEqual) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tGreater struct {
	Pos int
}

func (t *tGreater) itsMe(token scanToken) bool {
	return token.Value == ">"
}

func (t *tGreater) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tGreaterOrEqual struct {
	Pos int
}

func (t *tGreaterOrEqual) itsMe(token scanToken) bool {
	return token.Value == ">="
}

func (t *tGreaterOrEqual) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tAnd struct {
	Pos int
}

func (t *tAnd) itsMe(token scanToken) bool {
	return token.Value == "&&"
}

func (t *tAnd) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tXor struct {
	Pos int
}

func (t *tXor) itsMe(token scanToken) bool {
	return token.Value == "^"
}

func (t *tXor) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tOrPrecedence struct {
	Pos int
}

func (t *tOrPrecedence) itsMe(token scanToken) bool {
	return token.Value == "|"
}

func (t *tOrPrecedence) setValue(token scanToken) {
	t.Pos = token.Pos
}

func (t *tOrPrecedence) maybeWrong() bool {
	return true
}

func (t *tOrPrecedence) wrongValue() string {
	return "|"
}

type tOr struct {
	Pos int
}

func (t *tOr) itsMe(token scanToken) bool {
	return token.Value == "||"
}

func (t *tOr) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tNot struct {
	Pos int
}

func (t *tNot) itsMe(token scanToken) bool {
	return token.Value == "!"
}

func (t *tNot) setValue(token scanToken) {
	t.Pos = token.Pos
}

func (t *tNot) maybeWrong() bool {
	return true
}

func (t *tNot) wrongValue() string {
	return "!"
}

type tAssign struct {
	Pos int
}

func (t *tAssign) itsMe(token scanToken) bool {
	return token.Value == "="
}

func (t *tAssign) setValue(token scanToken) {
	t.Pos = token.Pos
}

func (t *tAssign) maybeWrong() bool {
	return true
}

func (t *tAssign) wrongValue() string {
	return "="
}

type tAssignPlus struct {
	Pos int
}

func (t *tAssignPlus) itsMe(token scanToken) bool {
	return token.Value == "+="
}

func (t *tAssignPlus) setValue(token scanToken) {
	t.Pos = token.Pos
}

type tVariable struct {
	Value string
	Pos   int
}

func (t *tVariable) itsMe(token scanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value)
}

func (t *tVariable) setValue(token scanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

type tPrefixedVariable struct {
	Value string
	Pos   int
}

func (t *tPrefixedVariable) itsMe(token scanToken) bool {
	return token.Value[0] == '$' && token.Value[1] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value[1:])
}

func (t *tPrefixedVariable) setValue(token scanToken) {
	t.Value = token.Value[1:]
	t.Pos = token.Pos
}
