package runner

import "fmt"

type PlainOutput struct {
}

func NewPlainOutput() *PlainOutput {
	return &PlainOutput{}
}

func (p *PlainOutput) GetName() string {
	return "plain"
}

func (p *PlainOutput) GetOutHandler(c *CmdExecutorImpl) func(msg ...interface{}) {
	return func(msg ...interface{}) {
		fmt.Println(msg...)
	}
}
