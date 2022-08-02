package tviewapp

import "github.com/gdamore/tcell/v2"

type CeBox struct {
	left, top, width, height int
	isChanged                bool
	drawStyle                tcell.Style
	OnMouseOver              func(x, y int)
	OnMouseLeave             func()
}

func NewBox() *CeBox {
	return &CeBox{
		width:     50,
		height:    50,
		left:      1,
		top:       1,
		drawStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue),
	}
}
func (box *CeBox) draw(ca *CellApp, cleanUp bool) {
	if cleanUp {
		ca.cleanArea(box.left, box.top, box.left+box.width, box.top+box.height)
		return
	}
	ca.drawBox(box.left, box.top, box.left+box.width, box.top+box.height, box.drawStyle, "")
}

func (box *CeBox) setStyle(style tcell.Style) {
	box.drawStyle = style
}

func (box *CeBox) hitTest(x, y int) bool {
	return defaultHitTest(x, y, box.left, box.top, box.width, box.height)
}

func (box *CeBox) onMouseOverHndl(x, y int) {
	if box.OnMouseOver != nil {
		box.OnMouseOver(x, y)
	}
}

func (box *CeBox) haveChanged() bool {
	return box.isChanged
}

func (box *CeBox) onMouseLeaveHndl() {
	if box.OnMouseLeave != nil {
		box.OnMouseLeave()
	}
}
func (box *CeBox) SetDim(left, top, width, height int) {
	box.left = left
	box.top = top
	box.width = width
	box.height = height
	box.isChanged = true
}

func (box *CeBox) SetOffset(left, top int) {
	box.isChanged = true
	box.left = left
	box.top = top
}
