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

 package linehack

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/scanner"
	"unicode"
)

type Parser struct {
	PrintByFmt          bool
	trashTokenCount     int
	trashTokenTrace     []string
	source              map[int]TokenSelfProvider
	stringVariables     map[string]string
	variableRequester   func(string) (interface{}, error) // is is the function that is asking for a variable value by name
	useTokens           []TokenSelfProvider
	neighborTokens      []TokenSelfProvider
	ConditionStartToken string
	ConditionEndToken   string
	execCallback        func(TokenSelfProvider, []TokenSelfProvider) (interface{}, error)
	maxPosition         int
}

type IdentPrefixes struct {
	Rune     rune
	CallBack func(TokenSelfProvider, interface{}) (interface{}, error)
}

func NewParser() *Parser {
	return &Parser{
		stringVariables: make(map[string]string),
	}
}

func BasicTokens() []TokenSelfProvider {
	return []TokenSelfProvider{
		&If{},
		&TBracketOpen{},
		&TBracketClose{},
		&TCurlyOpen{},
		&TCurlyClose{},
		&Then{},
		&Else{},
		&TVariable{},
		&TString{},
		&TEqual{},
		&TNotEqual{},
		&TLess{},
		&TLessOrEqual{},
		&TGreater{},
		&TGreaterOrEqual{},
		&TAnd{},
		&TOr{},
		&TBool{},
		&TNot{},
		&TOrPrecedence{},
		&TXor{},
		&TPrefixedVariable{},
	}
}

func (p *Parser) Parse(line string) {
	p.defaultsIfNotSet()
	tokens := p.lineScan(line)
	p.source = p.parseToken(tokens)
	p.Println("------------------")
	p.printDegugTokenMap(p.source)

}

func (p *Parser) Copy(start int, end int) *Parser {
	newParser := NewParser()
	newParser.source = p.getTokensInBetweenWithIndex(start, end)
	newParser.useTokens = p.useTokens
	newParser.neighborTokens = p.neighborTokens
	newParser.ConditionStartToken = p.ConditionStartToken
	newParser.ConditionEndToken = p.ConditionEndToken
	newParser.execCallback = p.execCallback
	newParser.maxPosition = p.maxPosition
	newParser.stringVariables = p.stringVariables
	newParser.variableRequester = p.variableRequester
	newParser.PrintByFmt = p.PrintByFmt

	return newParser
}

// defaultIfNotSet sets the default values for the parser
// if the values are not set
func (p *Parser) defaultsIfNotSet() {
	if p.ConditionStartToken == "" {
		p.ConditionStartToken = "(" // the default start token for conditions
	}
	if p.ConditionEndToken == "" {

		p.ConditionEndToken = ")" // the default end token for conditions
	}

	if p.useTokens == nil {
		p.useTokens = BasicTokens()
	}
}

// SetUseTokens sets the tokens that are used in the line
// and have to be parsed.
func (p *Parser) SetUseTokens(tokens []TokenSelfProvider) {
	p.useTokens = tokens
}

// AddUseToken adds a token that is used in the line
func (p *Parser) AddUseToken(token ...TokenSelfProvider) {
	p.useTokens = append(p.useTokens, token...)
}

// AddUseToken adds a token that is used in the line
func (p *Parser) AddUseTokenFirst(token ...TokenSelfProvider) {
	p.useTokens = append(token, p.useTokens...)
}

// SetNeighborTokens sets the tokens that are required to recheck
// if they are assigned, but depending on neighbors, it turns
// out they are are other tokens ment. like == and =, or > and >=
func (p *Parser) SetNeighborTokens(tokens []TokenSelfProvider) {
	p.neighborTokens = tokens
}

func (p *Parser) SetVariableRequester(requester func(string) (interface{}, error)) {
	p.variableRequester = requester
}

func (p *Parser) SetExecCallBack(callback func(TokenSelfProvider, []TokenSelfProvider) (interface{}, error)) {
	p.execCallback = callback
}

func (p *Parser) getVariableValue(name string) (interface{}, error) {
	if p.variableRequester != nil {
		return p.variableRequester(name)
	}
	return nil, fmt.Errorf("variable requester is not set")
}

func (p *Parser) getTokensInBetween(start int, end int) []TokenSelfProvider {
	var tokens []TokenSelfProvider
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos > start && pos < end {
			tokens = append(tokens, token)
		}
		return true
	})
	return tokens
}

func (p *Parser) getTokensInBetweenWithIndex(start int, end int) map[int]TokenSelfProvider {
	tokens := make(map[int]TokenSelfProvider)
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos > start && pos < end {
			tokens[pos] = token
		}
		return true
	})
	return tokens
}

func (p *Parser) findNextTokenOfTypes(startPos int, types []ScanToken) (int, TokenSelfProvider) {
	var pos int = -1
	var token TokenSelfProvider
	p.iteratemapInKeyorder(p.source, func(position int, t TokenSelfProvider) bool {
		if position > startPos {
			for _, typ := range types {
				if t.ItsMe(typ) {
					pos = position
					token = t
					return true
				}
			}
		}
		return true
	})
	return pos, token
}

func (p *Parser) findNextThenToken(startPos int) (int, TokenSelfProvider) {
	var posValue int = -1
	var tokenReturn TokenSelfProvider
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos > startPos {
			switch token.(type) {
			case ThenBehaviour:
				posValue = pos
				tokenReturn = token
				return true
			}
		}
		return true
	})
	return posValue, tokenReturn
}

func (p *Parser) getTokensBetween(start int, end int) []TokenSelfProvider {
	var tokens []TokenSelfProvider
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos > start && pos < end {
			tokens = append(tokens, token)
		}
		return true
	})
	return tokens
}

func (p *Parser) getTokensBetweenAsInterface(start int, end int) []interface{} {
	var tokens []interface{}
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos > start && pos < end {
			value := p.GetTokenValue(token)
			tokens = append(tokens, value)
		}
		return true
	})
	return tokens
}

func (p *Parser) findBetweenTokens(startPos int, startToken ScanToken, endToken ScanToken) (int, int, error) {
	var start int
	var end int
	var layer int
	var err error
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		if pos < startPos {
			return true
		} else {
			// no start token found til now
			// so we increase the layer and set the start position
			// the layer tells us how many brackets we have to find
			// before we can stop
			if start == 0 && token.ItsMe(startToken) {
				layer++
				start = pos
				return true // we found the start token, so we MUST stop before we increase the layer again
			}
			// increase the layer if we found a start token
			// and decrease the layer if we found an end token
			if layer > 0 {
				if token.ItsMe(startToken) {
					layer++
				}
				if token.ItsMe(endToken) {
					layer--
				}
			}

			// we found the start token and the end token
			// so we can stop
			// we have to check if the layer is 0, because we can have
			// nested brackets
			if layer == 0 && token.ItsMe(endToken) && start != 0 && end == 0 {
				end = pos
				return true
			}
		}
		return true
	})
	return start, end, err
}

func (p *Parser) Print(i ...interface{}) {
	if !p.PrintByFmt {
		return
	}
	fmt.Print(i...)
}

func (p *Parser) Println(i ...interface{}) {
	if !p.PrintByFmt {
		return
	}
	fmt.Println(i...)
}

func (p *Parser) Printf(format string, i ...interface{}) {
	if !p.PrintByFmt {
		return
	}
	fmt.Printf(format, i...)
}

func (p *Parser) lineScan(line string) []ScanToken {
	var token []ScanToken
	var scan scanner.Scanner
	scan.Init(strings.NewReader(line))
	scan.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanInts | scanner.ScanStrings | scanner.ScanRawStrings | scanner.SkipComments
	scan.IsIdentRune = func(ch rune, i int) bool {
		return ch == '$' && i == 0 || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	}

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		p.Printf("%s[%v]\t%s   \n", scan.Position, scan.Pos().Offset, scan.TokenText())
		token = append(token, p.createScanToken(scan.TokenText(), scan.Pos().Offset))
		if scan.Pos().Offset > p.maxPosition {
			p.maxPosition = scan.Pos().Offset
		}
	}
	p.Println("EOF-----------------")
	return token
}

func (p *Parser) getMaxPosition() int {
	return p.maxPosition
}

func (p *Parser) createScanToken(value string, pos int) ScanToken {
	return ScanToken{
		Pos:   pos,
		Value: value,
	}
}

func (p *Parser) parseToken(tokens []ScanToken) map[int]TokenSelfProvider {

	scanLine := make(map[int]TokenSelfProvider)
	for _, t := range tokens {
		p.Print("validate token: ", t.Value, "  @pos: ", t.Pos, "\t")
		checked := p.checkAndConvert(t)
		if checked != nil {
			scanLine[t.Pos] = checked.(TokenSelfProvider)
			p.Println(" ... added token: ", reflect.TypeOf(checked))
		}
	}
	return p.compileTokens(scanLine)
}

func (p *Parser) iteratemapInKeyorder(m map[int]TokenSelfProvider, iterFn func(pos int, token TokenSelfProvider) bool) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if iterFn == nil {
		return
	}
	for _, k := range keys {
		if cont := iterFn(k, m[k]); !cont {
			break
		}
	}
}

func (p *Parser) checkAndConvert(token ScanToken) interface{} {

	// lets check the tokens by his own, if the token is one of the checks, return the token
	// as a copy of the check
	for _, check := range p.useTokens {
		newToken := check.Copy()
		if newToken.ItsMe(token) {

			newToken.SetValue(token)
			return newToken
		}
	}

	tToken := &trashToken{
		Pos:   token.Pos,
		Value: token.Value,
	}

	return tToken
}

func (p *Parser) printDegugToken(token TokenSelfProvider) {
	p.Println("token", reflect.TypeOf(token).String(), token)
}

func (p *Parser) printDegugTokenMap(tokenMap map[int]TokenSelfProvider) {
	for pos, t := range tokenMap {
		p.Println("token", pos, reflect.TypeOf(t).String(), t)
	}
}

func (p *Parser) GetTokenValue(token TokenSelfProvider) interface{} {
	switch tkn := token.(type) {
	case *trashToken:
		return tkn.Value
	case *TString:
		// remove the quotes
		return strings.ReplaceAll(tkn.Value, "\"", "")
	case *TVariable:
		return tkn.Value
	case *TPrefixedVariable:
		if value, err := p.getVariableValue(tkn.Value); err != nil {
			return err.Error()
		} else {
			return value
		}
	}
	return nil
}

// compileTokens compile the tokens, so we can check the neighbors of the tokens
// we process the tokens by the map and return a new map.
// this new map should have any fixed tokens.
// a fixed token means for example two tAssign tokens in neighbor ([=],[=]) should be one tEqual token ([==]) and so on
func (p *Parser) compileTokens(tokens map[int]TokenSelfProvider) map[int]TokenSelfProvider {

	newTokenmap := map[int]TokenSelfProvider{}

	for pos, t := range tokens {
		switch tkn := t.(type) {
		case *trashToken:
			p.Println(" --- trashToken", pos, tkn.Value)
			p.trashTokenCount++ // count the trash tokens
			p.trashTokenTrace = append(p.trashTokenTrace, "unknow token ["+tkn.Value+"] type "+reflect.TypeOf(t).String())
			newTokenmap[pos] = t
		default:

			// check if the token is in the possibleHaveNeighbors
			if recheck, ok := t.(tokenCouldBeAnother); ok && recheck.maybeWrong() {
				p.Println("     i could be the wrong one checking ", recheck.wrongValue(), "  ", reflect.TypeOf(t).String(), tkn)
				//delete(tokens, pos)
				if nextInline, ok := tokens[pos+1]; ok {
					if recheckNext, ok := nextInline.(tokenCouldBeAnother); ok && recheckNext.maybeWrong() {
						recheckValue := recheck.wrongValue() + recheckNext.wrongValue()
						p.Println("    lets recheck with value ", recheckValue)

						// check all tokens they have 2 chars to handle. so we can check the neighbors of the tokens
						// they reporting they are maybe wrong.
						// if some of the current tokens (they reporting being maybe wrong) we look at the neighbors
						// and compare the neighbors with the wrongValue() of the token together with the current token
						// if this combined token is matching the itsMe() function of the token, we replace the current token with the new token
						for _, check := range p.neighborTokens {
							if check.ItsMe(p.createScanToken(recheckValue, pos)) {
								p.Println("    i am the right one", reflect.TypeOf(check).String(), check)
								check.SetValue(p.createScanToken(recheckValue, pos))
								newTokenmap[pos] = check
							}
						}
					}
				}
			} else {
				p.Println("     me is fine", pos, reflect.TypeOf(t).String(), tkn)
				newTokenmap[pos] = t
			}
		}
	}
	return newTokenmap
}
