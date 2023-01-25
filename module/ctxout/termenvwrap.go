package ctxout

import (
	"os"

	"github.com/muesli/termenv"
)

type termEnvWrap struct {
	output *termenv.Output
}

func NewTermEnvWrap() *termEnvWrap {
	return &termEnvWrap{}
}

func (t *termEnvWrap) Filter(msg interface{}) interface{} {
	if t.output == nil {
		t.Init()
	}
	return msg
}

func (t *termEnvWrap) Stream(msg ...interface{}) {
}

func (t *termEnvWrap) StreamLn(msg ...interface{}) {
}

func (t *termEnvWrap) ToString(msg ...interface{}) string {
	return ""
}

func (t *termEnvWrap) IsTerminal() bool {
	return false
}

func (t *termEnvWrap) IsColor() bool {
	return false
}

func (t *termEnvWrap) Init() {
	t.output = termenv.NewOutput(os.Stdout)
}
