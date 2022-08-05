package tviewapp

import (
	"github.com/gdamore/tcell/v2"
)

const (
	BubbleOff  = 0 // no 'bubble' check. so any element that can hit, is hitted
	BubbleDown = 1 // only the last element on the list (what means it is visual on top) can be hit
	BubbleUp   = 2 // reverse to BubleDown. means the first element (first drawn so visual in the background) only
)

var (
	lastHoverHits []int
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
	GetBehavior() CElementBehavior
}

// Defines the behavior of the element
type CElementBehavior struct {
	selectable bool // element can being selected
	movable    bool // element can be moved
	hovers     bool // element needs hover events
	static     bool // all elements they are not affected by any checks like mouse or key events. so they don't need to be handled
}

type CEListener struct {
	OnLMouseDown func(*CellApp, int, int)
	OnLMouseUp   func(*CellApp, int, int, int, int)
}

type CellApp struct {
	screen       tcell.Screen
	exitKey      tcell.Key
	style        tcell.Style
	baseElements map[int]CElement
	actives      []int
	statics      []int
	bubbleBehave int
	elementCount int
	Listener     CEListener
}

func New() *CellApp {
	return &CellApp{
		exitKey:      tcell.KeyEscape,
		bubbleBehave: BubbleDown,
		baseElements: make(map[int]CElement),
		elementCount: 0,
	}
}

// defaultHitTest is just a reusable helper for less code duplication
func defaultHitTest(hx, hy, x, y, w, h int) bool {
	if hx >= x && hy >= y && hx <= x+w && hy <= y+h {
		return true
	}
	return false
}

// AddElement adds visiual component to the application
// because it is added last, it be displayed on top
func (c *CellApp) AddElement(el ...CElement) {
	for _, elin := range el {
		c.baseElements[c.elementCount] = elin
		if elin.GetBehavior().static {
			c.statics = append(c.statics, c.elementCount)
		} else {
			c.actives = append(c.actives, c.elementCount)
		}
		c.elementCount++
	}
}

func (c *CellApp) getElementByIndex(index int) (CElement, bool) {
	if ce, ok := c.baseElements[index]; ok {
		return ce, true
	}
	return nil, false
}

// drawElements triggers the draw for any element
func (c *CellApp) drawElements() {
	c.cleanElements()
	for i := 0; i < c.elementCount; i++ {
		if el, ok := c.getElementByIndex(i); ok {
			el.draw(c, false)
		}
	}
}

func (c *CellApp) cleanElements() {
	for i := 0; i < c.elementCount; i++ {
		if el, ok := c.getElementByIndex(i); ok {
			if el.haveChanged() {
				el.draw(c, true)
			}
		}
	}
}

// leaveElementCheck triggers the on leave event on any element that is no longer
// affected by the hit test
// here we iterate over the last elements, they we had stored as hovered before
func (c *CellApp) leaveElementCheck(x, y int) {
	lastHoverHits = c.checkPreviousHitList(x, y, lastHoverHits, func(cE CElement, xa, ya int) { cE.onMouseLeaveHndl() })
}

// hoverElementCheck tests any element if they can hover, and if
// the mouse is over the element.
// checks any element if they is hit by x and y coordinats
// depends on the Bubble behavior
func (c *CellApp) hoverElementCheck(x, y int) {
	lastHoverHits = c.hitChecker(
		x, y,
		lastHoverHits, // here we have the list that contains any element that was hovering
		func(xa, ya int) { c.leaveElementCheck(xa, ya) },                // this callback triggers the check, if the previous elements still affected
		func(check CElement) bool { return check.GetBehavior().hovers }, // thats the verify if the element can hover
		func(cE CElement, xb, yb int) { cE.onMouseOverHndl(xb, yb) })    // if all is matching,we trigger the event
}

func (c *CellApp) RunLoop(exitCallBack func()) {
	if c.screen != nil {
		c.screen.EnableMouse()
		lx, ly := -1, -1
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
				/*
					if button != tcell.ButtonNone && ox < 0 {
						ox, oy = x, y
					}*/
				switch ev.Buttons() {
				case tcell.Button1:
					if lx < 0 {
						lx, ly = x, y
					}
					if c.Listener.OnLMouseDown != nil {
						c.Listener.OnLMouseDown(c, x, y)
					}

				case tcell.ButtonNone:
					if lx >= 0 && c.Listener.OnLMouseUp != nil {
						c.Listener.OnLMouseUp(c, x, y, lx, ly)
						lx, ly = -1, -1
					}
					/*
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
					*/
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
