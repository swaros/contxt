package ctxtcell

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/swaros/contxt/module/ctxout"
)

type CtOutput struct {
	info    ctxout.CtxOutBehavior
	parent  *CtCell
	outText *textElement
	myId    int
}

func NewCtOutput(parent *CtCell) *CtOutput {
	parent.GetScreen()
	op := &CtOutput{
		outText: parent.Text("debug"),
		parent:  parent,
	}
	op.outText.SetPos(1, 10)
	op.outText.SetDim(50, 10)
	op.myId, _ = op.parent.AddElement(op.outText)
	opref := op.parent.GetElementByID(op.myId)
	// cast to textElement
	opref.(*textElement).SetContent("output")

	return op
}

func NewCtOutputNoTty() *CtOutput {
	return &CtOutput{}
}

func (c *CtOutput) Filter(msg interface{}) interface{} {
	c.Stream(msg)
	return msg
}

func (c *CtOutput) Update(info ctxout.CtxOutBehavior) {
	c.info = info
}

func (c *CtOutput) StreamLn(msg ...interface{}) {
	c.Stream(msg...)
	c.Stream("\n")
}

func (c *CtOutput) Stream(msg ...interface{}) {
	txtBuffer := ""
	opref := c.parent.GetElementByID(c.myId)
	// cast to textElement
	opref.(*textElement).SetContent("stream")
	for _, m := range msg {
		switch m.(type) {
		case string:
			txtBuffer += m.(string)
		case []byte:
			txtBuffer += string(m.([]byte))
		}
	}
	c.printText(txtBuffer)
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
	x1 := 1
	row := 1
	col := 1
	x2, y2 := c.parent.GetScreen().Size()
	c.parent.AddDebugMessage(text)
	style := tcell.StyleDefault
	for _, r := range []rune(text) {
		c.parent.GetScreen().SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}
