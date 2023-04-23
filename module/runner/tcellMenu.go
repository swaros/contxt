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
	items         []*MenuElement
	parent        *CtCell
	selectedStyle tcell.Style
	regularStyle  tcell.Style
	hoverStyle    tcell.Style
	haveFocus     bool
	visible       bool
}

type MenuElement struct {
	text        *textElement
	coordinates Coordinates
	isSelected  bool
	OnSelect    func(*MenuElement)
}

// NewMenu creates a new menu
func (c *CtCell) NewMenu() *ctMenu {
	menu := &ctMenu{}
	menu.border = c.NewBox()
	menu.border.filled = true
	menu.regularStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	menu.border.fillStyle = menu.regularStyle
	menu.items = make([]*MenuElement, 0)
	menu.SetDefaultStyle()
	menu.parent = c
	menu.visible = true
	return menu
}

func (c *ctMenu) NewMenuElement(content string, onSelect func(*MenuElement)) *MenuElement {
	element := &MenuElement{}
	element.text = c.parent.Text(content)
	element.coordinates = Coordinates{}
	element.OnSelect = onSelect
	return element
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
		item.text.SetStyle(style)
	}
	c.border.SetFillStyle(style)
	return c
}

func (c *ctMenu) SetSelectedStyle(style tcell.Style) *ctMenu {
	c.selectedStyle = style
	return c
}

func (c *ctMenu) SetHoverStyle(style tcell.Style) *ctMenu {
	c.hoverStyle = style
	return c
}

func (c *ctMenu) SetDefaultStyle() *ctMenu {
	c.regularStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	c.hoverStyle = tcell.StyleDefault.Underline(true)
	c.selectedStyle = tcell.StyleDefault.Bold(true)
	return c
}

// SetColor sets the color of the menu
func (c *ctMenu) SetColor(fg, bg tcell.Color) *ctMenu {
	c.border.SetColor(fg, bg)
	return c
}

// AddItem adds a new item to the menu
func (c *ctMenu) AddItem(text string, onSelect func(*MenuElement)) *ctMenu {
	item := c.parent.Text(text)
	item.SetStyle(c.regularStyle)
	c.items = append(c.items, c.NewMenuElement(text, onSelect))
	return c
}

// Implement the interface

// Draw draws the menu
func (c *ctMenu) Draw(s tcell.Screen) Coordinates {
	coords := c.border.Draw(s)
	offX, offY := c.border.topLeft.GetXY(s)
	width, bottom := c.border.bottomRight.GetXY(s)
	for _, item := range c.items {
		item.text.SetPos(offX+1, offY+1)
		item.text.SetDim(width, 1)
		offY++
		if offY > bottom {
			break
		}
		if item.isSelected {
			item.text.SetStyle(c.selectedStyle)
		}
		item.coordinates = item.text.Draw(s)
	}
	return coords
}

// HandleEvent handles the events for the menu
func (c *ctMenu) MouseReleaseEvent(start position, end position, trigger int) {
	if c.HitTestFn(start) {
		c.parent.AddDebugMessage("mouse HIT on menu with ", trigger, " button")
		c.TestMenuEntry(start, func(item *MenuElement) {
			// if the item is already selected, we do executes the callback
			if item.isSelected {
				item.OnSelect(item)
			}
			item.isSelected = true
		}, func(item *MenuElement) {
			item.isSelected = false
		})
	}
}

func (c *ctMenu) MousePressEvent(start position, trigger int) {

}

func (c *ctMenu) MouseHoverEvent(pos position) {

}

func (c *ctMenu) MouseLeaveEvent() {

}

func (c *ctMenu) SetVisible(visible bool) {
	c.visible = visible
}

func (c *ctMenu) IsVisible() bool {
	return c.visible
}

// shortcut for the test event
func (c *ctMenu) HitTestFn(pos position) bool {
	// do not get the wrong wy around. topLeft is more right and down than bottomRight
	return pos.IsInBox(c.border.topLeft.GetReal(c.parent.screen), c.border.bottomRight.GetReal(c.parent.screen))
}

func (c *ctMenu) TestMenuEntry(pos position, onHit func(item *MenuElement), onMiss func(item *MenuElement)) {
	for _, item := range c.items {
		// we just need to check the y coordinate
		// inside the menu we checked already that the x coordinate is within the menu
		// so we we do not need to have an excact match on the text entry
		if pos.Y == item.coordinates.TopLeft.Y {
			onHit(item)
		} else {
			onMiss(item)
		}
	}
}

func (c *ctMenu) KeyEvent(key tcell.Key, r rune) {
	if c.haveFocus {
		switch key {
		case tcell.KeyUp:
			c.parent.AddDebugMessage("UP")
		case tcell.KeyDown:
			c.parent.AddDebugMessage("DOWN")
		case tcell.KeyLeft:
			c.parent.AddDebugMessage("LEFT")
		case tcell.KeyRight:
			c.parent.AddDebugMessage("RIGHT")
		case tcell.KeyEnter:
			c.parent.AddDebugMessage("ENTER")
		}
	}
}

func (c *ctMenu) Hit(pos position, s tcell.Screen) bool {
	// Check if the position is within the menu
	if c.HitTestFn(pos) {
		c.parent.AddDebugMessage("HIT MENU")
		c.TestMenuEntry(pos, func(item *MenuElement) {
			item.text.SetStyle(c.hoverStyle)
		}, func(item *MenuElement) {
			item.text.SetStyle(c.regularStyle)
		})
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
