package tviewapp

import (
	"github.com/gdamore/tcell/v2"
)

const (
	BubbleOff  = 0
	BubbleDown = 1
	BubbleUp   = 2
)

var (
	lastHits []CElement
)

type CeSize struct {
	width, height, left, top int
}

type CElement interface {
	draw(*CellApp, bool)
	hitTest(x, y int) bool
	setStyle(style tcell.Style)
	onMouseOverHndl(x, y int)
	onMouseLeaveHndl()
	haveChanged() bool
	SetDim(left, top, width, height int)
}

type CellApp struct {
	screen       tcell.Screen
	exitKey      tcell.Key
	style        tcell.Style
	baseElements []CElement
}

func New() *CellApp {
	return &CellApp{
		exitKey: tcell.KeyEscape,
	}
}

func defaultHitTest(hx, hy, x, y, w, h int) bool {
	if hx >= x && hy >= y && hx <= x+w && hy <= y+h {
		return true
	}
	return false
}

func (c *CellApp) AddElement(el ...CElement) {
	c.baseElements = append(c.baseElements, el...)
}

func (c *CellApp) drawElements() {
	c.cleanElements()
	for _, el := range c.baseElements {
		el.draw(c, false)
	}
}

func (c *CellApp) cleanElements() {
	for _, el := range c.baseElements {
		if el.haveChanged() {
			el.draw(c, true)
		}
	}
}

func (c *CellApp) leaveElementCheck(x, y int) {
	var cleanUp []CElement
	for _, el := range lastHits {
		if !el.hitTest(x, y) {
			el.onMouseLeaveHndl()
		} else {
			cleanUp = append(cleanUp, el)
		}
	}
	lastHits = cleanUp
}

func (c *CellApp) hoverElementCheck(x, y int) {
	c.leaveElementCheck(x, y)           // check first if some elements lost focus
	for _, el := range c.baseElements { // then lets test all
		if el.hitTest(x, y) {
			lastHits = append(lastHits, el) // keep track of elements that we had in focus
			el.onMouseOverHndl(x, y)        // trigger the handler
		}
	}
}

func (c *CellApp) RunLoop(exitCallBack func()) {
	if c.screen != nil {
		c.screen.EnableMouse()
		ox, oy := -1, -1
		for {
			// Update screen
			c.screen.Show()

			// draw all elements
			c.drawElements()

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

				// checking hovering over elements
				c.hoverElementCheck(x, y)

				button := ev.Buttons()
				// Only process button events, not wheel events
				button &= tcell.ButtonMask(0xff)

				if button != tcell.ButtonNone && ox < 0 {
					ox, oy = x, y
				}
				switch ev.Buttons() {
				case tcell.ButtonNone:
					if ox >= 0 {
						//label := fmt.Sprintf("%d,%d to %d,%d", ox, oy, x, y)
						//c.drawBox(ox, oy, x, y, boxStyle, label)
						newBox := NewBox()
						newBox.SetDim(ox, oy, x-ox, y-oy)
						c.AddElement(newBox)
						newBox.OnMouseOver = func(x, y int) {
							newBox.setStyle(tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkGreen))
						}

						newBox.OnMouseLeave = func() {
							newBox.setStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))
						}
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
		c.style = defStyle
		c.screen.Clear()
	}
}

func (c *CellApp) drawText(x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range string(text) {
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

func (c *CellApp) cleanArea(x1, y1, x2, y2 int) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			c.screen.SetContent(col, row, ' ', nil, c.style)
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
