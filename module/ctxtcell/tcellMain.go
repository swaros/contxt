// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

 package ctxtcell

import (
	"errors"
	"fmt"
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
	output        *CtOutput
	debug         bool
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

func (c *CtCell) SetDebug(debug bool) *CtCell {
	c.debug = debug
	return c
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
		if c.debug {
			debugMsg := strings.Join(debugMessages, ",")
			c.debugOut(debugMsg)
		}
		// show screen
		c.screen.Show()
		// remove any onscreen debug messages
		if c.debug {
			c.CleanDebugMessages()
		}

		c.loopTimer = time.Since(startLoopTimer)
	}
}

func (c *CtCell) SendEvent(ev tcell.Event) error {
	return c.screen.PostEvent(ev)
}

func (c *CtCell) GetLastLoopTime() time.Duration {
	return c.loopTimer
}

func (c *CtCell) GetOutput() *CtOutput {
	return c.output
}

func (c *CtCell) Init() error {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	if c.screen == nil {
		s, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		if err := s.Init(); err != nil {
			return err
		}
		c.SetScreen(s)
	}

	c.screen.SetStyle(defStyle)
	if c.MouseEnabled {
		c.screen.EnableMouse()
	}
	return nil
}

func (c *CtCell) Run() error {
	if c.screen == nil {
		return errors.New("screen not initialized")
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
