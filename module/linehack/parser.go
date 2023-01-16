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
	PrintByFmt        bool
	trashTokenCount   int
	trashTokenTrace   []string
	source            map[int]tokenSelfProvider
	stringVariables   map[string]string
	variableRequester func(string) (interface{}, error) // is is the functen that is asking for a variable value by name
}

func NewParser() *Parser {
	return &Parser{
		stringVariables: make(map[string]string),
	}
}

func (p *Parser) Parse(line string) {
	tokens := p.lineScan(line)
	p.source = p.parseToken(tokens)
	p.Println("------------------")
	p.printDegugTokenMap(p.source)

}

func (p *Parser) Execute(line string) error {
	p.Parse(line)
	if p.trashTokenCount > 0 {
		return fmt.Errorf("found %d trash tokens: %s", p.trashTokenCount, strings.Join(p.trashTokenTrace, ", "))
	}
	if err := p.runSource(); err != nil {
		return err
	}
	return nil
}

func (p *Parser) SetVariableRequester(requester func(string) (interface{}, error)) {
	p.variableRequester = requester
}

func (p *Parser) getVariableValue(name string) (interface{}, error) {
	if p.variableRequester != nil {
		return p.variableRequester(name)
	}
	return nil, fmt.Errorf("variable requester is not set")
}

func (p *Parser) runSource() error {
	var runErr error
	p.iteratemapInKeyorder(p.source, func(pos int, token tokenSelfProvider) {
		p.printDegugToken(token)

		switch tkn := token.(type) {
		// the if token is found, we need to find the brackets
		// and the tokens between them
		case *If:
			p.Println("if", tkn)

			left := p.createPatternToken("(")
			right := p.createPatternToken(")")

			if start, end, err := p.findBetweenTokens(tkn.Pos+1, left, right); err != nil {
				runErr = err
				return
			} else {
				if result, err := p.checkConditional(start, end); err != nil {
					runErr = err
					return
				} else {
					p.Println("result", result)
				}
			}
		}

	})
	return runErr
}

// createPatternToken creates a new token from a string.
// tokens of this kind is only used to find other tokens
// with the same type in the source map
func (p *Parser) createPatternToken(value string) scanToken {
	return p.createScanToken(value, -1)
}

func (p *Parser) checkConditional(start int, end int) (bool, error) {
	tokens := p.getTokensInBetween(start, end)
	boolResult := true   // default is true. so we have to find a false condition to make it false
	inComparing := false // set the comaring state, that tells us, we found a variable of function to be checked
	var comparingElements []tokenSelfProvider
	for _, token := range tokens {
		switch tkn := token.(type) {
		case *tBool:
			boolResult = boolResult && tkn.Value
		case *tPrefixedVariable, *tVariable, *tEqual, *tNotEqual, *tString, *tLess, *tLessOrEqual, *tGreater, *tGreaterOrEqual, *tOr, *tAnd, *tNot:
			if inComparing {
			} else {
				inComparing = true
			}
			comparingElements = append(comparingElements, tkn)
		default:
			return false, fmt.Errorf("unexpected token in condition Check %s", reflect.TypeOf(token))
		}
	}
	if inComparing {
		splitted := p.splitConditionByAndOr(comparingElements)
		for _, group := range splitted {
			if result, err := p.checkConditionGroup(group); err != nil {
				return false, err
			} else {
				boolResult = boolResult && result
			}
		}
	}
	return boolResult, nil
}

func (p *Parser) splitConditionByAndOr(tokens []tokenSelfProvider) [][]tokenSelfProvider {
	var groups [][]tokenSelfProvider
	var group []tokenSelfProvider
	for _, token := range tokens {
		switch token.(type) {
		case *tAnd, *tOr:
			groups = append(groups, group)
			group = []tokenSelfProvider{}
		default:
			group = append(group, token)
		}
	}
	groups = append(groups, group)
	return groups
}

func (p *Parser) checkConditionGroup(tokens []tokenSelfProvider) (bool, error) {
	var boolResult bool
	var first interface{}
	var second interface{}
	var comparedBy interface{}
	hitCnt := 0
	for _, token := range tokens {
		switch tkn := token.(type) {
		case *tBool:
			boolResult = boolResult && tkn.Value
			hitCnt++
		case *tPrefixedVariable:
			if value, err := p.getVariableValue(tkn.Value); err != nil {
				return false, err
			} else {
				if hitCnt == 0 {
					first = value
				} else {
					second = value
				}
			}
			hitCnt++

		case *tVariable:
			if hitCnt == 0 {
				first = p.stringVariables[tkn.Value]
			} else {
				second = p.stringVariables[tkn.Value]
			}
			hitCnt++

		case *tString:
			if hitCnt == 0 {
				first = p.stringVariables[tkn.Value]
			} else {
				second = p.stringVariables[tkn.Value]
			}
			hitCnt++

		case *tEqual, *tNotEqual, *tLess, *tLessOrEqual, *tGreater, *tGreaterOrEqual, *tNot:
			if first == nil {
				return false, fmt.Errorf("variable to check is nil")
			}
			// any check can only on the second place
			if hitCnt != 1 {
				return false, fmt.Errorf("unexpected token %s", reflect.TypeOf(token))
			}
			comparedBy = tkn
			hitCnt++
		}
	}

	if hitCnt < 3 {
		return false, fmt.Errorf("not enough tokens to check condition")
	}

	if hitCnt > 3 {
		return false, fmt.Errorf("too many tokens to check condition")
	}

	switch comparedBy.(type) {
	case *tEqual:
		boolResult = boolResult && (first == second)
	case *tNotEqual:
		boolResult = boolResult && (first != second)
	case *tLess:
		boolResult = boolResult && (first.(int) < second.(int))
	case *tLessOrEqual:
		boolResult = boolResult && (first.(int) <= second.(int))
	case *tGreater:
		boolResult = boolResult && (first.(int) > second.(int))
	case *tGreaterOrEqual:
		boolResult = boolResult && (first.(int) >= second.(int))
	case *tNot:
		boolResult = boolResult && !first.(bool)
	}

	return boolResult, nil
}

func (p *Parser) getTokensInBetween(start int, end int) []tokenSelfProvider {
	var tokens []tokenSelfProvider
	p.iteratemapInKeyorder(p.source, func(pos int, token tokenSelfProvider) {
		if pos > start && pos < end {
			tokens = append(tokens, token)
		}
	})
	return tokens
}

func (p *Parser) findBetweenTokens(startPos int, startToken scanToken, endToken scanToken) (int, int, error) {
	var start int
	var end int
	var layer int
	var err error
	p.iteratemapInKeyorder(p.source, func(pos int, token tokenSelfProvider) {
		if pos < startPos {
			return
		} else {
			// no start token found til now
			// so we increase the layer and set the start position
			// the layer tells us how many brackets we have to find
			// before we can stop
			if start == 0 && token.itsMe(startToken) {
				layer++
				start = pos
				return // we found the start token, so we MUST stop before we increase the layer again
			}
			// increase the layer if we found a start token
			// and decrease the layer if we found an end token
			if layer > 0 {
				if token.itsMe(startToken) {
					layer++
				}
				if token.itsMe(endToken) {
					layer--
				}
			}

			// we found the start token and the end token
			// so we can stop
			// we have to check if the layer is 0, because we can have
			// nested brackets
			if layer == 0 && token.itsMe(endToken) && start != 0 && end == 0 {
				end = pos
				return
			}
		}

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

func (p *Parser) lineScan(line string) []scanToken {
	var token []scanToken
	var scan scanner.Scanner
	scan.Init(strings.NewReader(line))
	scan.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanInts | scanner.ScanStrings | scanner.ScanRawStrings | scanner.SkipComments
	scan.IsIdentRune = func(ch rune, i int) bool {
		return ch == '$' && i == 0 || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	}

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		p.Printf("%s[%v]\t%s   \n", scan.Position, scan.Pos().Offset, scan.TokenText())
		token = append(token, p.createScanToken(scan.TokenText(), scan.Pos().Offset))
	}
	p.Println("EOF-----------------")
	return token
}

func (p *Parser) createScanToken(value string, pos int) scanToken {
	return scanToken{
		Pos:   pos,
		Value: value,
	}
}

func (p *Parser) parseToken(tokens []scanToken) map[int]tokenSelfProvider {

	scanLine := make(map[int]tokenSelfProvider)
	for _, t := range tokens {
		p.Print("validate token: ", t.Value, "  @pos: ", t.Pos, "\t")
		checked := p.checkAndConvert(t)
		if checked != nil {
			scanLine[t.Pos] = checked.(tokenSelfProvider)
			p.Println(" ... added token: ", reflect.TypeOf(checked))
		}
	}
	return p.compileTokens(scanLine)
}

func (p *Parser) iteratemapInKeyorder(m map[int]tokenSelfProvider, iterFn func(pos int, token tokenSelfProvider)) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if iterFn == nil {
		return
	}
	for _, k := range keys {
		iterFn(k, m[k])
	}
}

func (p *Parser) checkAndConvert(token scanToken) interface{} {

	var checks []tokenSelfProvider
	// add all token types to checks
	checks = append(checks,
		&If{},
		&tBracketOpen{},
		&tBracketClose{},
		&tCurlyOpen{},
		&tCurlyClose{},
		&tSemiColon{},
		&tThen{},
		&tElse{},
		&tPrint{},
		&tSet{},
		&tVariable{},
		&tString{},
		&tDivide{},
		&tMultiply{},
		&tPlus{},
		&tMinus{},
		&tEqual{},
		&tNotEqual{},
		&tLess{},
		&tLessOrEqual{},
		&tGreater{},
		&tGreaterOrEqual{},
		&tAnd{},
		&tOr{},
		&tModulo{},
		&tInt{},
		&tFloat{},
		&tBool{},
		&tVar{},
		&tNot{},
		&tAssign{},
		&tAssignPlus{},
		&tOrPrecedence{},
		&tXor{},
		&tPrefixedVariable{},
	)
	// lets check the tokens by his own, if the token is one of the checks, return the token
	for _, check := range checks {
		if check.itsMe(token) {
			check.setValue(token)
			return check
		}
	}

	tToken := &trashToken{
		Pos:   token.Pos,
		Value: token.Value,
	}

	return tToken
}

func (p *Parser) printDegugToken(token tokenSelfProvider) {
	p.Println("token", reflect.TypeOf(token).String(), token)
}

func (p *Parser) printDegugTokenMap(tokenMap map[int]tokenSelfProvider) {
	for pos, t := range tokenMap {
		p.Println("token", pos, reflect.TypeOf(t).String(), t)
	}
}

// compileTokens compile the tokens, so we can check the neighbors of the tokens
// we process the tokens by the map and return a new map.
// this new map should have any fixed tokens.
// a fixed token means for example two tAssign tokens in neighbor ([=],[=]) should be one tEqual token ([==]) and so on
func (p *Parser) compileTokens(tokens map[int]tokenSelfProvider) map[int]tokenSelfProvider {

	newTokenmap := map[int]tokenSelfProvider{}

	// some tokens may parsed wrong, so we need to check the tokens
	// we have to check neighbors of the tokens, if the token is not valid, we have to remove it
	// we have to check the token before and after the token
	// if the token is not valid, we have to remove it
	possibleHaveNeighbors := []tokenSelfProvider{}

	// add all tokens they have 2 charsto handle. so we can check the neighbors of the tokens
	// they reporting they are maybe wrong.
	// if some of the current tokens (they reporting being maybe wrong) we look at the neighbors
	// and compare the neighbors with the wrongValue() of the token together with the current token
	// if this combined token is matching the itsMe() function of the token, we replace the current token with the new token
	possibleHaveNeighbors = append(possibleHaveNeighbors,
		&tGreaterOrEqual{},
		&tLessOrEqual{},
		&tNotEqual{},
		&tAssignPlus{},
		&tEqual{},
		&tOr{},
	)

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

						for _, check := range possibleHaveNeighbors {
							if check.itsMe(p.createScanToken(recheckValue, pos)) {
								p.Println("    i am the right one", reflect.TypeOf(check).String(), check)
								check.setValue(p.createScanToken(recheckValue, pos))
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
