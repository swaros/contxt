package runner

import (
	"github.com/gdamore/tcell/v2"
)

// Path: module/runner/tcellMenu.go

// Menu is en element that can be used to create a menu
// it contains a box and a list of menu items.
// The menu items can be selected by using the arrow keys
// and the enter key.

type ctMenu struct {
	border        *ctBox
	items         []*textElement
	parent        *ctCell
	selectedStyle tcell.Style
	regularStyle  tcell.Style
	haveFocus     bool
}

// NewMenu creates a new menu
func (c *ctCell) NewMenu() *ctMenu {
	menu := &ctMenu{}
	menu.border = c.NewBox()
	menu.border.filled = true
	menu.regularStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	menu.border.fillStyle = menu.regularStyle
	menu.items = make([]*textElement, 0)
	menu.parent = c
	return menu
}

// SetTopLeft sets the top left position of the menu
func (c *ctMenu) SetTopLeft(x, y int) *ctMenu {
	c.border.SetTopLeft(x, y)
	return c
}

// SetBottomRight sets the bottom right position of the menu
func (c *ctMenu) SetBottomRight(x, y int) *ctMenu {
	c.border.SetBottomRight(x, y)
	return c
}

// SetTopLeftProcentage sets the top left position of the menu
// as a procentage of the screen size
func (c *ctMenu) SetTopLeftProcentage(x, y int) *ctMenu {
	c.border.SetTopLeftProcentage(x, y)
	return c
}

// SetBottomRightProcentage sets the bottom right position of the menu
// as a procentage of the screen size
func (c *ctMenu) SetBottomRightProcentage(x, y int) *ctMenu {
	c.border.SetBottomRightProcentage(x, y)
	return c
}

// SetStyle sets the style of the menu
func (c *ctMenu) SetStyle(style tcell.Style) *ctMenu {
	c.regularStyle = style
	for _, item := range c.items {
		item.SetStyle(style)
	}
	c.border.SetFillStyle(style)
	return c
}

func (c *ctMenu) SetSelectedStyle(style tcell.Style) *ctMenu {
	c.selectedStyle = style
	return c
}

// SetColor sets the color of the menu
func (c *ctMenu) SetColor(fg, bg tcell.Color) *ctMenu {
	c.border.SetColor(fg, bg)
	return c
}

// AddItem adds a new item to the menu
func (c *ctMenu) AddItem(text string) *ctMenu {
	item := c.parent.Text(text)
	item.SetStyle(c.regularStyle)
	c.items = append(c.items, item)
	return c
}

// Implement the interface

// Draw draws the menu
func (c *ctMenu) Draw(s tcell.Screen) {
	c.border.Draw(s)
	offX, offY := c.border.topLeft.GetXY(s)
	width, bottom := c.border.bottomRight.GetXY(s)
	for _, item := range c.items {
		item.SetPos(offX+1, offY+1)
		item.SetDim(width, 1)
		offY++
		if offY > bottom {
			break
		}
		item.Draw(s)
	}
}

// HandleEvent handles the events for the menu
func (c *ctMenu) MouseReleaseEvent(start position, end position, trigger int) {

}

func (c *ctMenu) MousePressEvent(start position, trigger int) {

}

func (c *ctMenu) MouseHoverEvent(pos position) {

}

func (c *ctMenu) MouseLeaveEvent() {

}

func (c *ctMenu) Hit(pos position) bool {
	// Check if the position is within the menu
	// do not get the wrong wy around. topLeft is more right and down than bottomRight
	if pos.IsInBox(c.border.topLeft.GetReal(c.parent.screen), c.border.bottomRight.GetReal(c.parent.screen)) {
		c.parent.AddDebugMessage("HIT MENU")
		for _, item := range c.items {
			if item.Hit(pos) {
				item.SetStyle(c.selectedStyle)
			} else {
				item.SetStyle(c.regularStyle)
			}
		}
		return true
	}
	c.parent.AddDebugMessage("NO MENU HIT")
	return false
}

func (c *ctMenu) Focus(activated bool) {
	c.haveFocus = activated
	if activated {
		c.border.SetStyle(c.selectedStyle)
	} else {
		c.border.SetStyle(c.regularStyle)
	}
}

func (c *ctMenu) IsSelectable() bool {
	return true
}
