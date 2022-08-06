package tviewapp

import "github.com/gdamore/tcell/v2"

type CECard struct {
	title        string
	dim          CeSize
	isChanged    bool
	drawStyle    tcell.Style
	headStyle    tcell.Style
	OnMouseOver  func(x, y int)
	OnMouseLeave func()
}

func NewCard(title string) *CECard {
	return &CECard{
		title:     title,
		headStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGray),
		drawStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorLightCyan),
		dim: CeSize{
			top:    5,
			left:   5,
			width:  15,
			height: 15,
		},
	}
}

func (card *CECard) draw(ca *CellApp, cleanUp bool) {
	if !cleanUp {
		ca.drawFrameWs(card.dim, card.headStyle)
		xa, ya, xb, _ := CESizeToCoords(card.dim)
		ca.drawFrame(xa, ya, xb, ya+3, card.headStyle)
		ca.drawText(xa, ya, xb, ya+1, card.drawStyle, card.title)

		//ca.drawHTopLine(xa, ya+2, card.dim.width, card.drawStyle)
	}
}

func (card *CECard) GetBehavior() CElementBehavior {
	return CElementBehavior{
		selectable: true,
		movable:    true,
		hovers:     true,
		static:     true,
	}
}

func (card *CECard) setStyle(style tcell.Style) {
	card.drawStyle = style
}

func (card *CECard) hitTest(x, y int) bool {
	xa, ya, xb, yb := CESizeToCoords(card.dim)
	return defaultHitTest(x, y, xa, ya, xb, yb)
}

func (card *CECard) onMouseOverHndl(x, y int) {
	if card.OnMouseOver != nil {
		card.OnMouseOver(x, y)
	}
}

func (card *CECard) onMouseLeaveHndl() {
	if card.OnMouseLeave != nil {
		card.OnMouseLeave()
	}
}

func (card *CECard) haveChanged() bool {
	return card.isChanged
}

func (card *CECard) SetDim(left, top, width, height int) {
	card.dim = CeSize{
		left:   left,
		width:  width,
		height: height,
		top:    top,
	}
}

func (card *CECard) SetOffset(left, top int) {
	card.dim.left = left
	card.dim.top = top
}
