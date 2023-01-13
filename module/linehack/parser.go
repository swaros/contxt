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
	PrintByFmt      bool
	trashTokenCount int
	trashTokenTrace []string
	source          map[int]tokenSelfProvider
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(line string) {
	tokens := p.lineScan(line)
	p.source = p.parseToken(tokens)
	p.Println("------------------")
	p.printDegugTokenMap(p.source)

	p.iteratemapInKeyorder(p.source, func(pos int, token tokenSelfProvider) {
		p.printDegugToken(token)
	})
}

func (p *Parser) Execute(line string) error {
	p.Parse(line)
	if p.trashTokenCount > 0 {
		return fmt.Errorf("found %d trash tokens: %s", p.trashTokenCount, strings.Join(p.trashTokenTrace, ", "))
	}
	return nil
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
