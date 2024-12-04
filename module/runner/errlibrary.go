package runner

import (
	"fmt"
	"strconv"
	"strings"
)

type ExplainFunf func(errParser *ErrParse, ref ErrorReference) (string, bool)
type ErrExplainerList []ErrExplainer

const (
	LineNumberIndicator = "[LN]"
	ForInfoErrMsgPlace  = "[ERRORMSG]"
	// the default errors we will try to explain
	YamlLineIndicator = "yaml: line " + LineNumberIndicator + ": "
)

type ErrExplainer struct {
	RefLine string
	Lookup  string
	Explain ExplainFunf
	Info    string
}

type ExplainLib struct {
	ErrExplainerList
}

func NewExplainLib() *ExplainLib {
	return &ExplainLib{}
}

func (e *ExplainLib) AddExplainerByArg(refLine, lookup string, explain ExplainFunf) {
	e.ErrExplainerList = append(e.ErrExplainerList, ErrExplainer{RefLine: refLine, Lookup: lookup, Explain: explain})
}

func (e *ExplainLib) AddExplainer(explain ErrExplainer) {
	e.ErrExplainerList = append(e.ErrExplainerList, explain)
}

func (e *ExplainLib) Explain(errParser *ErrParse) (string, bool) {
	for _, explainer := range e.ErrExplainerList {
		if explainer.Explain != nil {
			// did we found matching patterns? if so we will execute the assigned explainer function
			if errRef := e.ParseErrorString(explainer.RefLine, explainer.Lookup, errParser); errRef.Found {
				if result, ok := explainer.Explain(errParser, errRef); ok {
					// if we have a special info, we will replace the [ERRORMSG] with the result
					// and return them instead of the plain result
					if explainer.Info != "" {
						info := strings.ReplaceAll(explainer.Info, ForInfoErrMsgPlace, result)
						return info, ok
					}
					return result, ok
				}
			}
		}
	}
	return "", false
}

func (e *ExplainLib) ParseErrorString(pattern string, lookup string, errParser *ErrParse) ErrorReference {
	// the pattern defines how we can extract the line number
	// for example: yaml: line 1751: did not find expected '-' indicator
	// the pattern is "yaml: line [LN]: [MSG]"
	// [LN] marks the position of the line number
	// [MSG] marks the position of the message that we have to extract
	// if we have this error: template: contxt-functions:4: function "include" not defined
	// the pattern is "template: contxt-functions:[LN]: [MSG]"

	var ref ErrorReference
	if !strings.Contains(pattern, lookup) {
		return ref
	}
	// NEEDS TO BE IMPLEMENTED
	if lnStr, ok := e.extractPattern(pattern, lookup, errParser.Error()); ok {
		ln, err := strconv.Atoi(lnStr)
		if err != nil {
			return ref
		}
		ref.Found = true
		ref.LineNr = ln
	}

	return ref

}

func (e *ExplainLib) extractPattern(pattern string, lookup string, fromstring string) (string, bool) {
	if !strings.Contains(pattern, lookup) {
		return "", false
	}
	leftOfLnMarker := strings.Split(pattern, lookup)[0]
	rightOfLnMarker := strings.Split(pattern, lookup)[1]
	if !strings.Contains(fromstring, leftOfLnMarker) {
		return "", false
	}
	clearLeft := strings.Split(fromstring, leftOfLnMarker)[1]
	if !strings.Contains(clearLeft, rightOfLnMarker) {
		return "", false
	}
	return strings.Split(clearLeft, rightOfLnMarker)[0], true

}

func extractCodeHelper(source string, errParser *ErrParse, errRef ErrorReference) (string, bool) {
	parts := strings.Split(errParser.Err.Error(), ":")
	if len(parts) < 3 {
		return "yaml Error: nothing to explain for error: " + errParser.Error(), true
	} else {
		lineNr := errRef.LineNr
		lines := strings.Split(source, "\n")
		// we have the line number now we try to find the line
		if lineNr > len(lines) {
			return fmt.Sprintf("yaml Error: can not get line %d of total amout of source lines %d:", lineNr, len(lines)) + errParser.Error(), true
		}
		// we have the line now we try to find the position of the error
		defaultBeforeLines := 5
		defaultAfterLines := 5
		start := lineNr - defaultBeforeLines
		if start < 0 {
			start = 0
		}
		end := lineNr + defaultAfterLines
		if end > len(lines) {
			end = len(lines)
		}
		// we have the start and end now we try to get the lines
		code := make([]YamlLine, 0)
		for i := start; i < end; i++ {
			code = append(code, NewYamlCodeLine(i, lines[i], i == lineNr-1))
		}
		errParser.code = code
		return errParser.Error(), true
	}
}
