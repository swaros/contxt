package runner

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ctCell struct {
	MouseEnabled  bool
	screen        tcell.Screen
	regularStyles defaultStyles
}

type position struct {
	X            int
	Y            int
	isProcentage bool
}

type dim struct {
	w int
	h int
}

type defaultStyles struct {
	normal  tcell.Style
	hovered tcell.Style
	focused tcell.Style
	active  tcell.Style
}

var (
	debugMessages []string
)

func NewTcell() *ctCell {

	newct := &ctCell{}
	newct.regularStyles.normal = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	newct.regularStyles.hovered = tcell.StyleDefault.Underline(true)
	newct.regularStyles.focused = tcell.StyleDefault.Bold(true)
	newct.regularStyles.active = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite)
	return newct
}

func (c *ctCell) Init(mouse bool) {
	c.MouseEnabled = mouse
}

func (c *ctCell) AddDebugMessage(msg string) {
	debugMessages = append(debugMessages, msg)
}

func (c *ctCell) CleanDebugMessages() {
	debugMessages = []string{}
}

func (c *ctCell) debugOut(msg string) {
	w, h := c.screen.Size()
	row := h - 1
	col := 1
	width := w - 1
	for _, r := range msg {
		c.screen.SetContent(col, row, r, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
		col++
		if col >= width { // wrap to next line
			row++
			col = 1
		}
		if row > w { // get out of here if we hit the bottom
			break
		}
	}
}

func (c *ctCell) Loop() {
	ox, oy := -1, -1
	for {
		// Update screen
		c.CleanDebugMessages()
		c.DrawAll()
		c.screen.Show()
		// Poll event
		ev := c.screen.PollEvent()

		var clickEventPos position
		var clickReleaseEventPos position
		releaseBtnCache := 0

		debugMsg := strings.Join(debugMessages, ",")
		c.debugOut(debugMsg)

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			c.screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				c.screen.Sync()
			} else if ev.Rune() == 'C' || ev.Rune() == 'c' {
				c.screen.Clear()
			} else if ev.Key() == tcell.KeyTAB {
				c.CycleFocus()
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			clickEventPos = position{x, y, false}
			c.MouseHoverAll(clickEventPos) // trigger hover event
			c.screen.ShowCursor(x, y)
			// show mouse coords and debug messages

			c.debugOut(fmt.Sprintf("x: %d, y: %d debug[%s]:", x, y, debugMsg))
			switch ev.Buttons() {
			case tcell.Button1:
				c.MousePressAll(clickEventPos, 1)
				releaseBtnCache = 1
				if ox < 0 {
					ox, oy = x, y // record location when click started
				}

			case tcell.Button2:
				c.MousePressAll(clickEventPos, 2)
				releaseBtnCache = 2

			case tcell.Button3:
				c.MousePressAll(clickEventPos, 3)
				releaseBtnCache = 3

			case tcell.ButtonNone:
				if ox >= 0 {
					clickReleaseEventPos = position{ox, oy, false}
					c.MouseReleaseAll(clickEventPos, clickReleaseEventPos, releaseBtnCache)
					ox, oy = -1, -1
				}
			}
		}
	}
}

func (c *ctCell) Run() {

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	s, err := tcell.NewScreen()
	c.screen = s
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	c.screen.SetStyle(defStyle)
	if c.MouseEnabled {
		c.screen.EnableMouse()
	}
	c.screen.Clear()

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		c.screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()
	c.Loop()

}
