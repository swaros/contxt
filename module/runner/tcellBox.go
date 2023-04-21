package runner

import "github.com/gdamore/tcell/v2"

type ctBox struct {
	topLeft     position
	bottomRight position
	style       tcell.Style
	filled      bool
	fillStyle   tcell.Style
}

func (c *ctCell) NewBox() *ctBox {
	box := &ctBox{}
	box.filled = false
	box.style = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	return box
}

func (c *ctBox) SetTopLeft(x, y int) *ctBox {
	c.topLeft.X = x
	c.topLeft.Y = y
	return c
}

func (c *ctBox) SetBottomRight(x, y int) *ctBox {
	c.bottomRight.X = x
	c.bottomRight.Y = y
	return c
}

func (c *ctBox) SetTopLeftProcentage(x, y int) *ctBox {
	c.topLeft.X = x
	c.topLeft.Y = y
	c.topLeft.isProcentage = true
	return c
}

func (c *ctBox) SetBottomRightProcentage(x, y int) *ctBox {
	c.bottomRight.X = x
	c.bottomRight.Y = y
	c.bottomRight.isProcentage = true
	return c
}

func (c *ctBox) GetTopLeft() position {
	return c.topLeft
}

func (c *ctBox) GetBottomRight() position {
	return c.bottomRight
}

func (c *ctBox) SetStyle(style tcell.Style) *ctBox {
	c.style = style
	return c
}

func (c *ctBox) SetColor(fg, bg tcell.Color) *ctBox {
	c.style = c.style.Foreground(fg).Background(bg)
	return c
}

func (c *ctBox) SetFillStyle(style tcell.Style) *ctBox {
	c.fillStyle = style
	c.filled = true
	return c
}

func (c *ctBox) SetFillColor(fg, bg tcell.Color) *ctBox {
	c.fillStyle = c.fillStyle.Foreground(fg).Background(bg)
	c.filled = true
	return c
}

func (c *ctBox) SetUnfilled() *ctBox {
	c.filled = false
	return c
}

// implemeting the TcElement interface
func (c *ctBox) MouseReleaseEvent(start position, end position, trigger int) {

}

func (c *ctBox) MousePressEvent(pos position, trigger int) {

}

func (c *ctBox) MouseHoverEvent(pos position) {
}

func (c *ctBox) MouseLeaveEvent() {
}

func (c *ctBox) Focus(activated bool) {
}

func (c *ctBox) Hit(pos position) bool {
	// is not selectable by default
	return false
}

func Focus(activated bool) {
}

func (c *ctBox) IsSelectable() bool {
	return false
}

func (c *ctBox) Draw(s tcell.Screen) {
	w, h := s.Size()

	x1, y1 := c.topLeft.GetXY(s)
	x2, y2 := c.bottomRight.GetXY(s)

	if w == 0 || h == 0 {
		return
	}

	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	if c.filled {
		for row := y1; row <= y2; row++ {
			for col := x1; col <= x2; col++ {
				s.SetContent(col, row, ' ', nil, c.fillStyle)
			}
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, c.style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, c.style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, c.style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, c.style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, c.style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, c.style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, c.style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, c.style)
	}
}
