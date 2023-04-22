package runner

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type textElement struct {
	text       string
	pos        position
	dim        dim
	visible    bool
	style      tcell.Style
	OnClicked  func(pos position, trigger int)
	OnReleased func(start position, end position, trigger int)
	OnHover    func(pos position)
	OnLeave    func()
	FucusFn    func(activated bool)
	OnSelect   func(selected bool)
}

var (
	texts []textElement
)

func (t *textElement) SetVisible(visible bool) {
	t.visible = visible
}

func (t *textElement) IsVisible() bool {
	return t.visible
}

func (t *textElement) SetPos(x, y int) *textElement {
	t.pos.X = x
	t.pos.Y = y
	return t
}

func (t *textElement) GetPos() *position {
	return &t.pos
}

func (t *textElement) SetPosProcentage(x, y int) *textElement {
	t.pos.X = x
	t.pos.Y = y
	t.pos.isProcentage = true
	return t
}

func (t *textElement) SetDim(w, h int) *textElement {
	t.dim.w = w
	t.dim.h = h
	return t
}

func (t *textElement) SetStyle(style tcell.Style) *textElement {
	t.style = style
	return t
}

func (t *textElement) SetColor(fg, bg tcell.Color) *textElement {
	t.style = t.style.Foreground(fg).Background(bg)
	return t
}

func (t *textElement) SetBold(b bool) *textElement {
	t.style = t.style.Bold(b)
	return t
}

func (t *textElement) SetUnderline(u bool) *textElement {
	t.style = t.style.Underline(u)
	return t
}

func (t *textElement) SetReverse(r bool) *textElement {
	t.style = t.style.Reverse(r)
	return t
}

func (t *textElement) SetBlink(b bool) *textElement {
	t.style = t.style.Blink(b)
	return t
}

func (t *textElement) SetDimmed(d bool) *textElement {
	t.style = t.style.Dim(d)
	return t
}

func (t *textElement) SetItalic(i bool) *textElement {
	t.style = t.style.Italic(i)
	return t
}

func (t *textElement) SetStrikeThrough(s bool) *textElement {
	t.style = t.style.StrikeThrough(s)
	return t
}

func (t *textElement) SetAttributes(a tcell.AttrMask) *textElement {
	t.style = t.style.Attributes(a)
	return t
}

func (t *textElement) SetForeground(fg tcell.Color) *textElement {
	t.style = t.style.Foreground(fg)
	return t
}

func (t *textElement) SetBackground(bg tcell.Color) *textElement {
	t.style = t.style.Background(bg)
	return t
}

func (t *textElement) SetColor256(fg, bg int) *textElement {
	t.style = t.style.Foreground(tcell.NewRGBColor(int32(fg), 0, 0)).Background(tcell.NewRGBColor(int32(bg), 0, 0))
	return t
}

func (t *textElement) SetColorRGB(fg, bg tcell.Color) *textElement {
	t.style = t.style.Foreground(fg).Background(bg)
	return t
}

func (t *textElement) SetContent(content string) *textElement {
	t.text = content
	return t
}

func (t *textElement) SetContentf(format string, a ...interface{}) *textElement {
	t.text = fmt.Sprintf(format, a...)
	return t
}

func (t *textElement) SetContentln(a ...interface{}) *textElement {
	t.text = fmt.Sprintln(a...)
	return t
}

func (t *textElement) SetContentlnf(format string, a ...interface{}) *textElement {
	t.text = fmt.Sprintf(format, a...)
	t.text += "	"
	return t
}

func (t *textElement) ResizeByContent() *textElement {
	t.dim.w = len(t.text)
	return t
}

func (t *textElement) MouseReleaseEvent(start position, end position, trigger int) {
	if t.OnReleased != nil {
		t.OnReleased(start, end, trigger)
	}
}

func (t *textElement) MousePressEvent(pos position, trigger int) {
	if t.OnClicked != nil {
		t.OnClicked(pos, trigger)
	}
}

func (t *textElement) MouseHoverEvent(pos position) {
	if t.OnHover != nil {
		t.OnHover(pos)
	}
}

func (t *textElement) MouseLeaveEvent() {
	if t.OnLeave != nil {
		t.OnLeave()
	}
}

func (t *textElement) Focus(activated bool) {
	if t.FucusFn != nil {
		t.FucusFn(activated)
	}
}

func (t *textElement) IsSelectable() bool {
	return t.FucusFn != nil
}

func (t *textElement) Hit(pos position) bool {
	// if no event handler is set, we don't care about the hit
	if t.OnClicked == nil && t.OnReleased == nil && t.OnHover == nil && t.OnLeave == nil {
		return false
	}
	bottomRight := CreatePosition(t.pos.X+t.dim.w, t.pos.Y+t.dim.h)
	if pos.IsInBox(t.pos, bottomRight) {
		return true
	}

	// if the position is within the text element, we have a hit
	return pos.X >= t.pos.X && pos.X <= t.pos.X+t.dim.w && pos.Y >= t.pos.Y && pos.Y <= t.pos.Y+t.dim.h
}

func (t textElement) Draw(s tcell.Screen) Coordinates {

	col, row := t.pos.GetXY(s)
	width := t.dim.w + col
	height := 0
	for _, r := range t.text {
		s.SetContent(col, row, r, nil, t.style)
		col++
		height = row
		if col >= width { // wrap to next line
			row++
			col = t.pos.X
		}
		if row > t.pos.Y+t.dim.h { // get out of here if we hit the bottom
			break
		}
	}
	return *NewCoordinates(t.pos.GetReal(s), width, height)
}

// Text creates a new text element and returns a pointer to it.
func (c *ctCell) Text(content string) *textElement {
	var te textElement
	te.text = content
	te.visible = true
	te.pos.X = 0            // default is left edge
	te.pos.Y = 0            // default is top edge
	te.dim.w = len(content) // default is the length of the text
	te.dim.h = 0            // default is at least one line
	texts = append(texts, te)
	return &te
}

// Text creates a new text element and returns a pointer to it.
// and sets the default behavior of an element to be clickable, can be focuses and can be hovered
func (c *ctCell) ActiveText(content string) *textElement {
	var te textElement
	te.text = content
	te.visible = true
	te.pos.X = 0            // default is left edge
	te.pos.Y = 0            // default is top edge
	te.dim.w = len(content) // default is the length of the text
	te.dim.h = 0            // default is at least one line
	te.OnClicked = func(pos position, trigger int) {
		c.SetFocus(&te)
		if te.OnSelect != nil {
			te.OnSelect(true)
		}
	}
	te.OnReleased = func(start position, end position, trigger int) {
		// nothing yet
	}
	te.FucusFn = func(activated bool) {
		te.SetBold(activated)
	}
	te.OnHover = func(pos position) {
		te.SetUnderline(true)
	}
	te.OnLeave = func() {
		te.SetUnderline(false)
	}
	texts = append(texts, te)
	return &te
}
