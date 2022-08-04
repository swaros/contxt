package tviewapp

import (
	"github.com/gdamore/tcell/v2"
)

const (
	BubbleOff  = 0 // no 'bubble' check. so any element that can bit hit, is hitted
	BubbleDown = 1 // only the last element on the list (what means it is visual on top) can be hit
	BubbleUp   = 2 // reverse to BubleDown. means the first element (first drawn so visual in the background) only
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
	bubbleBehave int
}

func New() *CellApp {
	return &CellApp{
		exitKey:      tcell.KeyEscape,
		bubbleBehave: BubbleDown,
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

// leaveElementCheck triggers the on leave event on any element that is no longer
// affected by the hit test
// here we iterate over the last elements, they we had stored as hovered before
func (c *CellApp) leaveElementCheck(x, y int) {

	if len(lastHits) < 1 { // no stored hits, we get out early
		return
	}

	var cleanUp []CElement // prepare the new list of elements they stiff affected
	switch c.bubbleBehave {
	case BubbleOff, BubbleUp:
		for _, el := range lastHits {
			if !el.hitTest(x, y) {
				el.onMouseLeaveHndl()
			} else {
				cleanUp = append(cleanUp, el) // memorize again this element, becasue it is not affected by hittest
			}
			if c.bubbleBehave == BubbleUp { // in case we just need the first element (again. is visual the element in the background)
				lastHits = cleanUp // copy the list here already
				return             // and get out
			}
		}
	case BubbleDown: // this case is for the elements in front of the ui (visual) only. elements behind them do not count
		found := false
		for i := len(lastHits) - 1; i >= 0; i-- { // we look from the other way around ..last in map first
			el := lastHits[i]

			if !el.hitTest(x, y) { // this element is even no longer hivered. so we trigger the leave
				el.onMouseLeaveHndl()
			} else {
				if !found { // as long we did not find a hovered element, we keep memorize the elements that still hovered.
					cleanUp = append(cleanUp, el) // but this can be juts one element...
					found = true                  // ...because from now on, we ignore anything else
				} else {
					el.onMouseLeaveHndl() // becasue we found our top element, any other have to leave
				}
			}

		}
	}
	lastHits = cleanUp
}

// checks any element if they is hit by x and y coordinats
// depends on the Bubble behavior
func (c *CellApp) hoverElementCheck(x, y int) {
	c.leaveElementCheck(x, y) // check first if some elements lost focus
	switch c.bubbleBehave {
	case BubbleOff, BubbleUp: // here we fire the event to any element that is in range
		for _, el := range c.baseElements { // then lets test all
			if el.hitTest(x, y) {
				lastHits = append(lastHits, el) // keep track of elements that we had in focus
				el.onMouseOverHndl(x, y)        // trigger the handler
				if c.bubbleBehave == BubbleUp {
					return // if just want the first it, by BubbleUp behavior, we get out now
				}
			}
		}
	case BubbleDown: // here we fire the event on the last one in the list (what is being drawed latest)
		var hitEl CElement = nil
		for _, el := range c.baseElements { // then lets test all
			if el.hitTest(x, y) {
				hitEl = el
			}
		}
		if hitEl != nil {
			hitEl.onMouseOverHndl(x, y)        // trigger the handler
			lastHits = append(lastHits, hitEl) // save this one
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
