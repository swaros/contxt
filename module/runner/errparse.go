package runner

type ErrParse struct {
	Err        error
	session    *CmdSession
	code       []YamlLine
	explainLib *ExplainLib
}

type YamlLine struct {
	LineNr  int
	Line    string
	IsError bool
}

func NewErrParse(err error, session *CmdSession) *ErrParse {
	return &ErrParse{
		Err:        err,
		session:    session,
		explainLib: NewDefaultExplainer(),
	}
}

func NewYamlCodeLine(lineNr int, line string, isError bool) YamlLine {
	return YamlLine{
		LineNr:  lineNr,
		Line:    line,
		IsError: isError,
	}
}

func (e *ErrParse) Error() string {
	return e.Err.Error()
}

func (e *ErrParse) Explain() string {
	if msg, ok := e.explainLib.Explain(e); ok {
		return msg
	}

	return e.Error() // fallback the error if we could not give more details
}

type ErrorReference struct {
	Found  bool
	LineNr int
}
