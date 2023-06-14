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

 package ctxtcell

import (
	"fmt"
	"strings"

	"github.com/swaros/contxt/module/ctxout"
)

// this is the output filter for the tcell module
// that will be injected to ctxout
type CtOutput struct {
	info         ctxout.CtxOutBehavior
	parent       *CtCell
	outText      *textElement
	myId         int
	stringBuffer []string
}

func NewCtOutput(parent *CtCell) *CtOutput {
	parent.GetScreen()
	op := &CtOutput{
		outText: parent.Text("debug"),
		parent:  parent,
	}
	op.outText.SetPos(0, 0)
	op.outText.SetDim(50, 10)
	op.myId, _ = op.parent.AddElement(op.outText)
	opref := op.parent.GetElementByID(op.myId)
	// cast to textElement
	opref.(*textElement).SetContent("output " + fmt.Sprintf("%v", op.myId))

	return op
}

func NewCtOutputNoTty() *CtOutput {
	return &CtOutput{}
}

func (c *CtOutput) Update(info ctxout.CtxOutBehavior) {
	c.info = info
}

func (c *CtOutput) StreamLn(msg ...interface{}) {
	c.Stream(msg...)
	c.printText(ctxout.ToString(strings.Join(c.stringBuffer, "") + "\n"))
}

func (c *CtOutput) Stream(msg ...interface{}) {
	txtBuffer := ctxout.ToString(msg...)
	opref := c.parent.GetElementByID(c.myId)
	// cast to textElement
	opref.(*textElement).SetContent("stream")
	c.stringBuffer = append(c.stringBuffer, txtBuffer)
}

// non interface functions
func (c *CtOutput) printText(text string) {
	// without an parent we can't print using screen
	// so we print to stdout
	if c.parent == nil {
		fmt.Println(text)
		return
	}
	c.outText.SetContent(text).SetDim(c.parent.GetScreen().Size())

}
