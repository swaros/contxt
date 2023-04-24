package ctxtcell

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type CtCell struct {
	MouseEnabled  bool
	noClearScreen bool
	screen        tcell.Screen
	regularStyles defaultStyles
	stopSign      bool
	loopTimer     time.Duration
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

func NewTcell() *CtCell {
	newct := &CtCell{}
	newct.regularStyles.normal = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	newct.regularStyles.hovered = tcell.StyleDefault.Underline(true)
	newct.regularStyles.focused = tcell.StyleDefault.Bold(true)
	newct.regularStyles.active = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorWhite)
	return newct
}

func (c *CtCell) SetMouse(mouse bool) *CtCell {
	c.MouseEnabled = mouse
	return c
}

func (c *CtCell) SetScreen(s tcell.Screen) *CtCell {
	c.screen = s
	return c
}

func (c *CtCell) GetScreen() tcell.Screen {
	return c.screen
}

func (c *CtCell) Stop() {
	c.stopSign = true
}

func (c *CtCell) SetNoClearScreen(noclear bool) *CtCell {
	c.noClearScreen = noclear
	return c
}

func (c *CtCell) AddDebugMessage(msg ...interface{}) {
	txtMsg := fmt.Sprint(msg...)
	debugMessages = append(debugMessages, txtMsg)
}

func (c *CtCell) CleanDebugMessages() {
	debugMessages = []string{}
}

func (c *CtCell) debugOut(msg string) {
	w, h := c.screen.Size()
	row := h - 2
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

func (c *CtCell) Loop() {
	ox, oy := -1, -1
	startLoopTimer := time.Now()
	for !c.stopSign {
		log.Println(".:.")
		// clear screen if not disabled
		if !c.noClearScreen {
			c.screen.Clear()
		}

		// Poll event
		ev := c.screen.PollEvent()

		var mousePos position
		var clickReleaseEventPos position
		releaseBtnCache := 0

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
			} else if ev.Key() == tcell.KeyTAB { // cycle focus by pressing tab
				c.CycleFocus()
			}
		case *tcell.EventMouse:
			c.AddDebugMessage("Mouse event")
			x, y := ev.Position()
			mousePos = CreatePosition(x, y, false)
			c.MouseHoverAll(mousePos) // trigger hover event
			//c.screen.ShowCursor(x, y)

			// show mouse coords and debug messages
			c.AddDebugMessage(fmt.Sprintf("[mouse (x: %d, y: %d)] ", x, y))
			switch ev.Buttons() {
			case tcell.Button1:
				c.MousePressAll(mousePos, 1)
				releaseBtnCache = 1
				if ox < 0 {
					ox, oy = x, y // record location when click started
				}

			case tcell.Button2:
				c.MousePressAll(mousePos, 2)
				releaseBtnCache = 2

			case tcell.Button3:
				c.MousePressAll(mousePos, 3)
				releaseBtnCache = 3

			case tcell.ButtonNone:
				if ox >= 0 {
					clickReleaseEventPos = CreatePosition(ox, oy, false)
					c.MouseReleaseAll(mousePos, clickReleaseEventPos, releaseBtnCache)
					ox, oy = -1, -1
				}
			}
		}
		// draw all elements
		c.AddDebugMessage("DRAWING")
		c.DrawAll()
		debugMsg := strings.Join(debugMessages, ",")
		c.debugOut(debugMsg)
		// show screen
		c.screen.Show()
		// remove any onscreen debug messages
		c.CleanDebugMessages()

		c.loopTimer = time.Since(startLoopTimer)
	}
}

func (c *CtCell) SendEvent(ev tcell.Event) error {
	return c.screen.PostEvent(ev)
}

func (c *CtCell) GetLastLoopTime() time.Duration {
	return c.loopTimer
}

func (c *CtCell) Run() error {

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	if c.screen == nil {
		s, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		if err := s.Init(); err != nil {
			return err
		}
		c.screen = s
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
	return nil
}
