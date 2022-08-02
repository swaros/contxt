package tviewapp

import "github.com/gdamore/tcell/v2"

type CeString struct {
	isChanged                bool
	text                     string
	left, top, width, height int
	drawStyle                tcell.Style
	OnMouseOver              func(x, y int)
	OnMouseLeave             func()
	autosize                 bool
	prevState                CeSize
}

func NewText(text string) *CeString {
	return &CeString{
		isChanged: false,
		text:      text,
		width:     len(text),
		height:    1,
		left:      0,
		top:       0,
		autosize:  true,
		drawStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple),
	}

}

func NewTextLine(text string, width int) *CeString {
	return &CeString{
		isChanged: false,
		text:      text,
		width:     width,
		height:    1,
		left:      0,
		top:       0,
		autosize:  false,
		drawStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple),
	}
}

func (str *CeString) savePrevState() {
	str.isChanged = true
	str.prevState = CeSize{
		width:  str.width,
		height: str.height,
		top:    str.top,
		left:   str.left,
	}
}

func (str *CeString) SetText(text string) {
	str.text = text
	if str.autosize {
		str.savePrevState()
		str.width = len(text)
	}
}

func (str *CeString) haveChanged() bool {
	return str.isChanged
}

func (str *CeString) draw(ca *CellApp, cleanUp bool) {
	if cleanUp {
		ca.cleanArea(str.prevState.left, str.prevState.top, str.prevState.left+str.prevState.width, str.prevState.top+str.prevState.height)
		return
	}
	ca.drawText(str.left, str.top, str.left+str.width, str.top+str.height, str.drawStyle, str.text)
}

func (str *CeString) setStyle(style tcell.Style) {
	str.drawStyle = style
}

func (str *CeString) hitTest(x, y int) bool {
	return defaultHitTest(x, y, str.left, str.top, str.width, str.height)
}

func (str *CeString) onMouseOverHndl(x, y int) {
	if str.OnMouseOver != nil {
		str.OnMouseOver(x, y)
	}
}

func (str *CeString) onMouseLeaveHndl() {
	if str.OnMouseLeave != nil {
		str.OnMouseLeave()
	}
}

func (str *CeString) SetDim(left, top, width, height int) {
	str.savePrevState()
	str.left = left
	str.top = top
	str.width = width
	str.height = height
}

func (str *CeString) SetOffset(left, top int) {
	str.savePrevState()
	str.left = left
	str.top = top
}
