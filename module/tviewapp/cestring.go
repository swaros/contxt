package tviewapp

import "github.com/gdamore/tcell/v2"

type CeString struct {
	text                     string
	left, top, width, height int
	drawStyle                tcell.Style
	OnMouseOver              func(x, y int)
	OnMouseLeave             func()
	autosize                 bool
}

func NewText(text string) *CeString {
	return &CeString{
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
		text:      text,
		width:     width,
		height:    1,
		left:      0,
		top:       0,
		autosize:  false,
		drawStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple),
	}
}

func (str *CeString) SetText(text string) {
	str.text = text
	if str.autosize {
		str.width = len(text)
	}
}

func (str *CeString) draw(ca *CellApp) {
	ca.drawText(str.left, str.top, str.left+str.width, str.top+str.height, str.drawStyle, str.text)
}

func (str *CeString) setStyle(style tcell.Style) {
	str.drawStyle = style
	str.autosize = false
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
	str.left = left
	str.top = top
	str.width = width
	str.height = height
}

func (str *CeString) SetOffset(left, top int) {
	str.left = left
	str.top = top
}
