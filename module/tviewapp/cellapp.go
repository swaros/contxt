package tviewapp

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type CellApp struct {
	screen  tcell.Screen
	exitKey tcell.Key
}

func New() *CellApp {
	return &CellApp{
		exitKey: tcell.KeyEscape,
	}
}

func (c *CellApp) RunLoop(exitCallBack func()) {
	if c.screen != nil {
		c.screen.EnableMouse()
		boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)
		c.drawBox(1, 1, 42, 7, boxStyle, "Click and drag to draw a box")
		c.drawBox(5, 9, 32, 14, boxStyle, "Press C to reset")
		ox, oy := -1, -1
		for {
			// Update screen
			c.screen.Show()

			// Poll event
			ev := c.screen.PollEvent()

			// Process event
			switch ev := ev.(type) {
			case *tcell.EventResize:
				c.screen.Sync()
			case *tcell.EventKey:
				if ev.Key() == c.exitKey || ev.Key() == tcell.KeyCtrlC {
					c.Exit()
					exitCallBack()
				}
			case *tcell.EventMouse:
				x, y := ev.Position()
				button := ev.Buttons()
				// Only process button events, not wheel events
				button &= tcell.ButtonMask(0xff)

				if button != tcell.ButtonNone && ox < 0 {
					ox, oy = x, y
				}
				switch ev.Buttons() {
				case tcell.ButtonNone:
					if ox >= 0 {
						label := fmt.Sprintf("%d,%d to %d,%d", ox, oy, x, y)
						c.drawBox(ox, oy, x, y, boxStyle, label)
						ox, oy = -1, -1
					}
				}

			}
		}
	}
}

func (c *CellApp) Exit() {
	if c.screen != nil {
		c.screen.Fini()
	}
}

func (c *CellApp) NewScreen() {
	if s, err := tcell.NewScreen(); err != nil {
		panic(err)
	} else {
		if iErr := s.Init(); iErr != nil {
			panic(iErr)
		}
		c.screen = s
		c.setupStyle()
	}
}

func (c *CellApp) setupStyle() {
	if c.screen != nil {
		defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
		c.screen.SetStyle(defStyle)
		c.screen.Clear()
	}
}

func (c *CellApp) drawText(x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		c.screen.SetContent(col, row, r, nil, style)
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

func (c *CellApp) drawBox(x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			c.screen.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		c.screen.SetContent(col, y1, tcell.RuneHLine, nil, style)
		c.screen.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		c.screen.SetContent(x1, row, tcell.RuneVLine, nil, style)
		c.screen.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		c.screen.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		c.screen.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		c.screen.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		c.screen.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}

	c.drawText(x1+1, y1+1, x2-1, y2-1, style, text)
}
