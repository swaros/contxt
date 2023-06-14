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

 // this file contains the token types as a simple set
// of definitions for a basic parsing environment.
// it can be used as a base for a more complex parsing environment

package linehack

import (
	"regexp"
	"strconv"
	"strings"
)

// this is the basic token type
// it is used to scan the input string
type ScanToken struct {
	Pos   int    //reported position in line
	Value string // the value of the token
}

// this is the interface that all token types must implement
// it is used to check if a token is of a certain type
type TokenSelfProvider interface {
	ItsMe(token ScanToken) bool // check if the token is of this type
	SetValue(token ScanToken)   // set the value of the token
	Copy() TokenSelfProvider    // copy the token
}

// this is the interface that all token types must implement
// the could be misleading by reading a single character.
// so for example = could be == or =
// this interface is used to check if the token is of a different type
// the token iteself to not care about any other possibility
// instead: any other maybe matching token like != or == or >= or <=
// will be tested with the neighboring token.
// so that means that the other token itself checks if he is ment instead of the current token
type tokenCouldBeAnother interface {
	maybeWrong() bool   // reports the possibility of being a different token
	wrongValue() string // returns the value current value.
}

// this is the interface that all token types must implement
// that is ment for start any conditional check.
// like IF, WHILE, FOR, SWITCH, CASE, DEFAULT
type ConditionalCheck interface {
	conditionalCheck() bool
	getToken() ScanToken
}

type Condition interface {
	compareValues(interface{}, interface{}) bool
}

type ConditionChain interface {
	IsStillValid([]bool) bool
}

type ThenBehaviour interface {
	ThenBehaviour() bool
}

type ElseBehaviour interface {
	ElseBehaviour() bool
}

type Runable interface {
	Run(...interface{}) (interface{}, error) // run the token
	GetPos() int                             // returns the position of the token
}

// this is the token, for any "not found" token
type trashToken struct {
	Pos   int
	Value string
}

func (t *trashToken) ItsMe(token ScanToken) bool {
	return false
}

func (t *trashToken) SetValue(token ScanToken) {
	t.Pos = token.Pos
	t.Value = token.Value
}

func (t *trashToken) Copy() TokenSelfProvider {
	return &trashToken{Pos: t.Pos, Value: t.Value}
}

// the conditional IF token
type If struct {
	Pos int
}

func (i *If) ItsMe(token ScanToken) bool {
	return token.Value == "if"
}

func (i *If) SetValue(token ScanToken) {
	i.Pos = token.Pos
}

func (i *If) Copy() TokenSelfProvider {
	return &If{Pos: i.Pos}
}

// the conditionalCheck interface is implemented
func (i *If) conditionalCheck() bool {
	return true
}

func (i *If) getToken() ScanToken {
	return ScanToken{Pos: i.Pos, Value: "if"}
}

type TBracketOpen struct {
	Pos int
}

func (t *TBracketOpen) ItsMe(token ScanToken) bool {
	return token.Value == "("
}

func (t *TBracketOpen) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TBracketOpen) Copy() TokenSelfProvider {
	return &TBracketOpen{Pos: t.Pos}
}

type TBracketClose struct {
	Pos int
}

func (t *TBracketClose) ItsMe(token ScanToken) bool {
	return token.Value == ")"
}

func (t *TBracketClose) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TBracketClose) Copy() TokenSelfProvider {
	return &TBracketClose{Pos: t.Pos}
}

type TCurlyOpen struct {
	Pos int
}

func (t *TCurlyOpen) ItsMe(token ScanToken) bool {
	return token.Value == "{"
}

func (t *TCurlyOpen) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TCurlyOpen) Copy() TokenSelfProvider {
	return &TCurlyOpen{Pos: t.Pos}
}

type TCurlyClose struct {
	Pos int
}

func (t *TCurlyClose) ItsMe(token ScanToken) bool {
	return token.Value == "}"
}

func (t *TCurlyClose) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TCurlyClose) Copy() TokenSelfProvider {
	return &TCurlyClose{Pos: t.Pos}
}

type TSemiColon struct {
	Pos int
}

func (t *TSemiColon) ItsMe(token ScanToken) bool {
	return token.Value == ";"
}

func (t *TSemiColon) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TSemiColon) Copy() TokenSelfProvider {
	return &TSemiColon{Pos: t.Pos}
}

type Then struct {
	Pos int
}

func (t *Then) ItsMe(token ScanToken) bool {
	// we accept here indepent of the case
	return strings.ToLower(token.Value) == "then"
}

func (t *Then) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *Then) Copy() TokenSelfProvider {
	return &Then{Pos: t.Pos}
}

func (t *Then) ThenBehaviour() bool {
	return true
}

type Else struct {
	Pos int
}

func (t *Else) ItsMe(token ScanToken) bool {
	// we accept here indepent of the case
	return strings.ToLower(token.Value) == "else"
}

func (t *Else) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *Else) ElseBehaviour() bool {
	return true
}

func (t *Else) Copy() TokenSelfProvider {
	return &Else{Pos: t.Pos}
}

type TSet struct {
	Pos int
}

func (t *TSet) ItsMe(token ScanToken) bool {
	return token.Value == "set"
}

func (t *TSet) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TSet) Copy() TokenSelfProvider {
	return &TSet{Pos: t.Pos}
}

type TString struct {
	Value string
	Pos   int
}

func (t *TString) ItsMe(token ScanToken) bool {
	return token.Value[0] == '"' && token.Value[len(token.Value)-1] == '"'
}

func (t *TString) SetValue(token ScanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

func (t *TString) Copy() TokenSelfProvider {
	return &TString{Value: t.Value, Pos: t.Pos}
}

type TInt struct {
	Value int
	Pos   int
}

func (t *TInt) ItsMe(token ScanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[0-9]+$`).MatchString(token.Value)
}

func (t *TInt) SetValue(token ScanToken) {
	t.Pos = token.Pos
	if i, err := strconv.Atoi(token.Value); err == nil {
		t.Value = i
	}
}

func (t *TInt) Copy() TokenSelfProvider {
	return &TInt{Value: t.Value, Pos: t.Pos}
}

type TFloat struct {
	Value float64
	Pos   int
}

func (t *TFloat) ItsMe(token ScanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[0-9]+\.[0-9]+$`).MatchString(token.Value)
}

func (t *TFloat) SetValue(token ScanToken) {
	t.Pos = token.Pos
	if i, err := strconv.ParseFloat(token.Value, 64); err == nil {
		t.Value = i
	}
}

func (t *TFloat) Copy() TokenSelfProvider {
	return &TFloat{Value: t.Value, Pos: t.Pos}
}

type TBool struct {
	Value bool
	Pos   int
}

func (t *TBool) ItsMe(token ScanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && (token.Value == "true" || token.Value == "false")
}

func (t *TBool) SetValue(token ScanToken) {
	t.Pos = token.Pos
	if token.Value == "true" {
		t.Value = true
	} else {
		t.Value = false
	}
}

func (t *TBool) Copy() TokenSelfProvider {
	return &TBool{Value: t.Value, Pos: t.Pos}
}

type TPrint struct {
	Pos int
}

func (t *TPrint) ItsMe(token ScanToken) bool {
	return token.Value == "print"
}

func (t *TPrint) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TPrint) Copy() TokenSelfProvider {
	return &TPrint{Pos: t.Pos}
}

type TVar struct {
	Value string
	Pos   int
}

func (t *TVar) ItsMe(token ScanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value)
}

func (t *TVar) SetValue(token ScanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

func (t *TVar) Copy() TokenSelfProvider {
	return &TVar{Value: t.Value, Pos: t.Pos}
}

type TPlus struct {
	Pos int
}

func (t *TPlus) ItsMe(token ScanToken) bool {
	return token.Value == "+"
}

func (t *TPlus) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TPlus) maybeWrong() bool {
	return true
}

func (t *TPlus) wrongValue() string {
	return "+"
}

func (t *TPlus) Copy() TokenSelfProvider {
	return &TPlus{Pos: t.Pos}
}

type TMinus struct {
	Pos int
}

func (t *TMinus) ItsMe(token ScanToken) bool {
	return token.Value == "-"
}

func (t *TMinus) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TMinus) maybeWrong() bool {
	return true
}

func (t *TMinus) wrongValue() string {
	return "-"
}

func (t *TMinus) Copy() TokenSelfProvider {
	return &TMinus{Pos: t.Pos}
}

type TMultiply struct {
	Pos int
}

func (t *TMultiply) ItsMe(token ScanToken) bool {
	return token.Value == "*"
}

func (t *TMultiply) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TMultiply) maybeWrong() bool {
	return true
}

func (t *TMultiply) wrongValue() string {
	return "*"
}

func (t *TMultiply) Copy() TokenSelfProvider {
	return &TMultiply{Pos: t.Pos}
}

type TDivide struct {
	Pos int
}

func (t *TDivide) ItsMe(token ScanToken) bool {
	return token.Value == "/"
}

func (t *TDivide) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TDivide) maybeWrong() bool {
	return true
}

func (t *TDivide) wrongValue() string {
	return "/"
}

func (t *TDivide) Copy() TokenSelfProvider {
	return &TDivide{Pos: t.Pos}
}

type TModulo struct {
	Pos int
}

func (t *TModulo) ItsMe(token ScanToken) bool {
	return token.Value == "%"
}

func (t *TModulo) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TModulo) maybeWrong() bool {
	return true
}

func (t *TModulo) wrongValue() string {
	return "%"
}

func (t *TModulo) Copy() TokenSelfProvider {
	return &TModulo{Pos: t.Pos}
}

type TEqual struct {
	Pos int
}

func (t *TEqual) ItsMe(token ScanToken) bool {
	return token.Value == "=="
}

func (t *TEqual) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TEqual) Copy() TokenSelfProvider {
	return &TEqual{Pos: t.Pos}
}

func (t *TEqual) compareValues(a, b interface{}) bool {
	switch a.(type) {
	case int:
		return a.(int) == b.(int)
	case float64:
		return a.(float64) == b.(float64)
	case string:

		return strings.EqualFold(a.(string), b.(string))

	case bool:
		return a.(bool) == b.(bool)
	}
	return a == b
}

type TNotEqual struct {
	Pos int
}

func (t *TNotEqual) ItsMe(token ScanToken) bool {
	return token.Value == "!="
}

func (t *TNotEqual) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TNotEqual) compareValues(a, b interface{}) bool {
	return a != b
}

func (t *TNotEqual) Copy() TokenSelfProvider {
	return &TNotEqual{Pos: t.Pos}
}

type TLess struct {
	Pos int
}

func (t *TLess) ItsMe(token ScanToken) bool {
	return token.Value == "<"
}

func (t *TLess) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TLess) compareValues(a, b interface{}) bool {
	return a.(int) < b.(int)
}

func (t *TLess) Copy() TokenSelfProvider {
	return &TLess{Pos: t.Pos}
}

type TLessOrEqual struct {
	Pos int
}

func (t *TLessOrEqual) ItsMe(token ScanToken) bool {
	return token.Value == "<="
}

func (t *TLessOrEqual) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TLessOrEqual) compareValues(a, b interface{}) bool {
	return a.(int) <= b.(int)
}

func (t *TLessOrEqual) Copy() TokenSelfProvider {
	return &TLessOrEqual{Pos: t.Pos}
}

type TGreater struct {
	Pos int
}

func (t *TGreater) ItsMe(token ScanToken) bool {
	return token.Value == ">"
}

func (t *TGreater) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TGreater) compareValues(a, b interface{}) bool {
	return a.(int) > b.(int)
}

func (t *TGreater) Copy() TokenSelfProvider {
	return &TGreater{Pos: t.Pos}
}

type TGreaterOrEqual struct {
	Pos int
}

func (t *TGreaterOrEqual) ItsMe(token ScanToken) bool {
	return token.Value == ">="
}

func (t *TGreaterOrEqual) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TGreaterOrEqual) compareValues(a, b interface{}) bool {
	return a.(int) >= b.(int)
}

func (t *TGreaterOrEqual) Copy() TokenSelfProvider {
	return &TGreaterOrEqual{Pos: t.Pos}
}

type TAnd struct {
	Pos int
}

func (t *TAnd) ItsMe(token ScanToken) bool {
	return token.Value == "&&"
}

func (t *TAnd) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TAnd) IsStillValid(args []bool) bool {
	for _, arg := range args {
		if !arg {
			return false
		}
	}
	return true
}

func (t *TAnd) Copy() TokenSelfProvider {
	return &TAnd{Pos: t.Pos}
}

type TXor struct {
	Pos int
}

func (t *TXor) ItsMe(token ScanToken) bool {
	return token.Value == "^"
}

func (t *TXor) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TXor) IsStillValid(args []bool) bool {
	count := 0
	for _, arg := range args {
		if arg {
			count++
		}
	}
	return count == 1
}

func (t *TXor) Copy() TokenSelfProvider {
	return &TXor{Pos: t.Pos}
}

type TOrPrecedence struct {
	Pos int
}

func (t *TOrPrecedence) ItsMe(token ScanToken) bool {
	return token.Value == "|"
}

func (t *TOrPrecedence) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TOrPrecedence) maybeWrong() bool {
	return true
}

func (t *TOrPrecedence) wrongValue() string {
	return "|"
}

func (t *TOrPrecedence) Copy() TokenSelfProvider {
	return &TOrPrecedence{Pos: t.Pos}
}

type TOr struct {
	Pos int
}

func (t *TOr) ItsMe(token ScanToken) bool {
	return token.Value == "||"
}

func (t *TOr) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TOr) IsStillValid(args []bool) bool {
	for _, arg := range args {
		if arg {
			return true
		}
	}
	return false
}

func (t *TOr) Copy() TokenSelfProvider {
	return &TOr{Pos: t.Pos}
}

type TNot struct {
	Pos int
}

func (t *TNot) ItsMe(token ScanToken) bool {
	return token.Value == "!"
}

func (t *TNot) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TNot) maybeWrong() bool {
	return true
}

func (t *TNot) wrongValue() string {
	return "!"
}

func (t *TNot) Copy() TokenSelfProvider {
	return &TNot{Pos: t.Pos}
}

type TAssign struct {
	Pos int
}

func (t *TAssign) ItsMe(token ScanToken) bool {
	return token.Value == "="
}

func (t *TAssign) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TAssign) maybeWrong() bool {
	return true
}

func (t *TAssign) wrongValue() string {
	return "="
}

func (t *TAssign) Copy() TokenSelfProvider {
	return &TAssign{Pos: t.Pos}
}

type TAssignPlus struct {
	Pos int
}

func (t *TAssignPlus) ItsMe(token ScanToken) bool {
	return token.Value == "+="
}

func (t *TAssignPlus) SetValue(token ScanToken) {
	t.Pos = token.Pos
}

func (t *TAssignPlus) Copy() TokenSelfProvider {
	return &TAssignPlus{Pos: t.Pos}
}

type TVariable struct {
	Value string
	Pos   int
}

func (t *TVariable) ItsMe(token ScanToken) bool {
	return token.Value[0] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value)
}

func (t *TVariable) SetValue(token ScanToken) {
	t.Value = token.Value
	t.Pos = token.Pos
}

func (t *TVariable) Copy() TokenSelfProvider {
	return &TVariable{Value: t.Value, Pos: t.Pos}
}

type TPrefixedVariable struct {
	Value string
	Pos   int
}

func (t *TPrefixedVariable) ItsMe(token ScanToken) bool {
	return token.Value[0] == '$' && token.Value[1] != '"' && token.Value[len(token.Value)-1] != '"' && regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(token.Value[1:])
}

func (t *TPrefixedVariable) SetValue(token ScanToken) {
	t.Value = token.Value[1:]
	t.Pos = token.Pos
}

func (t *TPrefixedVariable) Copy() TokenSelfProvider {
	return &TPrefixedVariable{Value: t.Value, Pos: t.Pos}
}
