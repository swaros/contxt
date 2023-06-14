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
	"strings"
)

func (p *Parser) Execute(line string) (interface{}, error) {
	p.Parse(line)
	// for any execution we need to check if there are any trash tokens
	// if there are any, we need to return an error
	if p.trashTokenCount > 0 {
		return nil, fmt.Errorf("found %d trash tokens: %s", p.trashTokenCount, strings.Join(p.trashTokenTrace, ", "))
	}
	if res, err := p.runSource(); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func (p *Parser) runSource() (interface{}, error) {
	var runErr error
	var myResult interface{}
	p.iteratemapInKeyorder(p.source, func(pos int, token TokenSelfProvider) bool {
		p.printDegugToken(token)

		switch tkn := token.(type) {

		case Runable:
			p.Println("RUNABLE::: ", tkn)
			// we need to check if the next token is an assign token
			// or if we have brackets, what would mean that we have a function call
			left := p.createPatternToken(p.ConditionStartToken)
			right := p.createPatternToken(p.ConditionEndToken)
			start, end, err := p.findBetweenTokens(tkn.GetPos()+1, left, right)
			if err != nil {
				runErr = err
				return false
			}
			if start > -1 && end > -1 {
				// we have brackets
				p.Println("CMD::: ", tkn)
				argument := p.getTokensBetweenAsInterface(start, end)
				if result, err := tkn.Run(argument...); err != nil {
					runErr = err
					return false
				} else {
					p.Println("RESULT::: ", reflect.TypeOf(result), result)
					myResult = result
				}

			} else {
				// we have no brackets
				p.Println("NO CMD::: ", tkn)
			}

		// we hit a condition
		// we need to check if the condition is fulfilled
		// a condition can have a condition in brackets
		// if there is a condition in brackets, we need to check this first
		// if there is no condition in brackets, we need to check the next token
		// if the condition is fulfilled, we need to check the next token
		case ConditionalCheck:
			p.Println(" CONDITION::: ", tkn.getToken().Value, tkn)
			conditionfulfilled := true // we assume that the condition is fulfilled
			// check without condition in brackets

			// check with condition in brackets
			left := p.createPatternToken(p.ConditionStartToken)
			right := p.createPatternToken(p.ConditionEndToken)
			nextThenPos, _ := p.findNextThenToken(tkn.getToken().Pos)

			// for any condition we need to check if there is a then token
			// if there is no then token, we have an error
			if nextThenPos == -1 {
				runErr = fmt.Errorf("no then condition found")
				return false
			}

			if start, end, err := p.findBetweenTokens(tkn.getToken().Pos+1, left, right); err != nil {
				// we have an error case
				runErr = err
				return false
			} else {
				// we found a condition in brackets
				// the position must be in a position before the next "then token"
				if start > 0 && end > 0 && start < nextThenPos && end < nextThenPos {
					// we have a condition in brackets
				} else {
					// we have no condition in brackets
					// so we need to check the next token
					conditionfulfilled = p.ConditionWithChain(tkn.getToken().Pos+1, nextThenPos)

				}

			}

			myResult = conditionfulfilled
			if conditionfulfilled {
				fmt.Println("condition fulfilled")
				copyRun := p.Copy(nextThenPos+1, p.maxPosition)
				if res, err := copyRun.runSource(); err != nil {
					runErr = err
					return false

				} else {
					myResult = res
				}

			} else {
				fmt.Println("condition not fulfilled")
				return false
			}
		}
		return true
	})
	return myResult, runErr
}

type ConditionalChain struct {
	group     []TokenSelfProvider
	nextCheck ConditionChain
}

func (p *Parser) ConditionWithChain(start int, end int) bool {
	tokens := p.getTokensInBetween(start, end)
	var group []TokenSelfProvider

	var chains []ConditionalChain
	var chainGroups [][]ConditionalChain

	for _, token := range tokens {
		switch tkn := token.(type) {
		case ConditionChain:
			group = []TokenSelfProvider{}
			chains = append(chains, ConditionalChain{group: group, nextCheck: tkn})
			chainGroups = append(chainGroups, chains)
			chains = []ConditionalChain{}

		default:
			group = append(group, token)
		}
	}
	if len(group) > 0 {
		chains = append(chains, ConditionalChain{group: group, nextCheck: nil})
		chainGroups = append(chainGroups, chains)
	}
	p.Println("chainGroups", chainGroups)
	//p.printDegugToken(groups)
	boolRes := p.checkConditionalChain(chainGroups)

	return boolRes

}

func (p *Parser) checkConditionalChain(chain [][]ConditionalChain) bool {
	boolResult := true
	runResult := false
	var checkHndl ConditionChain
	var keepForNext ConditionChain

	for index, group := range chain {

		checkHndl, runResult = p.checkConditionalChainGroup(group)
		if checkHndl != nil {
			keepForNext = checkHndl
		}
		if index == 0 {
			boolResult = runResult
		} else {
			if keepForNext != nil {
				args := []bool{boolResult, runResult}
				boolResult = keepForNext.IsStillValid(args)
				// invalidate the handler
				keepForNext = nil
			}
			boolResult = boolResult && runResult
		}
	}
	return boolResult
}

func (p *Parser) checkConditionalChainGroup(group []ConditionalChain) (ConditionChain, bool) {
	boolResult := true
	var nextChecker ConditionChain
	for _, chainCondition := range group {
		boolResult = boolResult && p.checkChainGroup(chainCondition.group)
		nextChecker = chainCondition.nextCheck

	}
	return nextChecker, boolResult
}

func (p *Parser) checkChainGroup(group []TokenSelfProvider) bool {
	boolResult := false
	for index, token := range group {
		switch tkn := token.(type) {
		case Condition:
			if index == 1 && len(group) > 2 {
				a := p.GetTokenValue(group[index-1])
				b := p.GetTokenValue(group[index+1])
				boolResult = tkn.compareValues(a, b)

			}
		default:
			p.Println("condition Check", reflect.TypeOf(token), "bool result", boolResult)
		}
	}
	return boolResult
}

// createPatternToken creates a new token from a string.
// tokens of this kind is only used to find other tokens
// with the same type in the source map
func (p *Parser) createPatternToken(value string) ScanToken {
	return p.createScanToken(value, -1)
}

func (p *Parser) checkConditional(start int, end int) (bool, error) {
	tokens := p.getTokensInBetween(start, end)

	boolResult := true   // default is true. so we have to find a false condition to make it false
	inComparing := false // set the comaring state, that tells us, we found a variable of function to be checked
	var comparingElements []TokenSelfProvider
	for _, token := range tokens {
		switch tkn := token.(type) {
		case *TBool:
			boolResult = boolResult && tkn.Value
		case *TPrefixedVariable, *TVariable, *TEqual, *TNotEqual, *TString, *TLess, *TLessOrEqual, *TGreater, *TGreaterOrEqual, *TOr, *TAnd, *TNot:
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

func (p *Parser) checkConditionalObsolete(start int, end int) (bool, error) {
	tokens := p.getTokensInBetween(start, end)
	boolResult := true   // default is true. so we have to find a false condition to make it false
	inComparing := false // set the comaring state, that tells us, we found a variable of function to be checked
	var comparingElements []TokenSelfProvider
	for _, token := range tokens {
		switch tkn := token.(type) {
		case *TBool:
			boolResult = boolResult && tkn.Value
		case *TPrefixedVariable, *TVariable, *TEqual, *TNotEqual, *TString, *TLess, *TLessOrEqual, *TGreater, *TGreaterOrEqual, *TOr, *TAnd, *TNot:
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

func (p *Parser) splitConditionByAndOr(tokens []TokenSelfProvider) [][]TokenSelfProvider {
	var groups [][]TokenSelfProvider
	var group []TokenSelfProvider
	for _, token := range tokens {
		switch token.(type) {
		case *TAnd, *TOr:
			groups = append(groups, group)
			group = []TokenSelfProvider{}
		default:
			group = append(group, token)
		}
	}
	groups = append(groups, group)
	return groups
}

func (p *Parser) checkConditionGroup(tokens []TokenSelfProvider) (bool, error) {
	var boolResult bool
	var first interface{}
	var second interface{}
	var comparedBy interface{}
	hitCnt := 0
	for _, token := range tokens {
		switch tkn := token.(type) {
		case *TBool:
			boolResult = boolResult && tkn.Value
			hitCnt++
		case *TPrefixedVariable:
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

		case *TVariable:
			if hitCnt == 0 {
				first = p.stringVariables[tkn.Value]
			} else {
				second = p.stringVariables[tkn.Value]
			}
			hitCnt++

		case *TString:
			if hitCnt == 0 {
				first = p.stringVariables[tkn.Value]
			} else {
				second = p.stringVariables[tkn.Value]
			}
			hitCnt++

		case *TEqual, *TNotEqual, *TLess, *TLessOrEqual, *TGreater, *TGreaterOrEqual, *TNot:
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
	case *TEqual:
		boolResult = boolResult && (first == second)
	case *TNotEqual:
		boolResult = boolResult && (first != second)
	case *TLess:
		boolResult = boolResult && (first.(int) < second.(int))
	case *TLessOrEqual:
		boolResult = boolResult && (first.(int) <= second.(int))
	case *TGreater:
		boolResult = boolResult && (first.(int) > second.(int))
	case *TGreaterOrEqual:
		boolResult = boolResult && (first.(int) >= second.(int))
	case *TNot:
		boolResult = boolResult && !first.(bool)
	}

	return boolResult, nil
}
