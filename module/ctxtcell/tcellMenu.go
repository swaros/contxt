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
	"github.com/gdamore/tcell/v2"
)

// Path: module/runner/tcellMenu.go

// Menu is en element that can be used to create a menu
// it contains a box and a list of menu items.
// The menu items can be selected by using the arrow keys
// and the enter key.

// MenuElement is an element in a menu
type ctMenu struct {
	border        *ctBox         // the border of the menu
	items         []*MenuElement // the items in the menu
	parent        *CtCell        // the parent tcell manager
	selectedStyle tcell.Style    // the style of the selected item
	regularStyle  tcell.Style    // the style of the regular items
	hoverStyle    tcell.Style    // the style of the hovered item
	haveFocus     bool           // if the menu has focus
	visible       bool           // if the menu is visible
	id            int            // the id of the menu
}

// MenuElement is an element in a menu
type MenuElement struct {
	text        *textElement       // the text of the element
	coordinates Coordinates        // the coordinates of the element
	isSelected  bool               // if the element is selected
	OnSelect    func(*MenuElement) // the function that is called when the element is selected
	reference   interface{}        // a reference to an object
}

// GetText returns the text Element of the menu element
func (m *MenuElement) GetText() *textElement {
	return m.text
}

// GetReference returns the reference of the menu element
// this can be nil. it is an interface{} so it can be anything
// also this is only valid for MenuElements created with NewMenuElementWithRef
func (m *MenuElement) GetReference() interface{} {
	return m.reference
}

// NewMenu creates a new menu and sets the default style
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

// NewMenuElement creates a new menu element
func (c *ctMenu) NewMenuElement(content string, onSelect func(*MenuElement)) *MenuElement {
	element := &MenuElement{}
	element.text = c.parent.Text(content)
	element.coordinates = Coordinates{}
	element.OnSelect = onSelect
	return element
}

// NewMenuElementWithRef creates a new menu element that contains a reference
func (c *ctMenu) NewMenuElementWithRef(content string, ref interface{}, onSelect func(*MenuElement)) *MenuElement {
	element := &MenuElement{}
	element.text = c.parent.Text(content)
	element.coordinates = Coordinates{}
	element.reference = ref
	element.OnSelect = onSelect
	return element
}

// SetID sets the id of the menu
func (c *ctMenu) SetID(id int) {
	c.id = id
}

// GetID returns the id of the menu
func (c *ctMenu) GetID() int {
	return c.id
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

// SetSelectedStyle sets the style of the selected item
func (c *ctMenu) SetSelectedStyle(style tcell.Style) *ctMenu {
	c.selectedStyle = style
	return c
}

// SetHoverStyle sets the style of the hovered item
func (c *ctMenu) SetHoverStyle(style tcell.Style) *ctMenu {
	c.hoverStyle = style
	return c
}

// SetDefaultStyle sets the default style of the menu
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

// AddItem adds a new item to the menu including an reference
func (c *ctMenu) AddItemWithRef(text string, reference interface{}, onSelect func(*MenuElement)) *ctMenu {
	item := c.parent.Text(text)
	item.SetStyle(c.regularStyle)
	c.items = append(c.items, c.NewMenuElementWithRef(text, reference, onSelect))

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
			item.isSelected = true
			item.OnSelect(item)
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

// Focus sets the focus of the menu
func (c *ctMenu) Focus(activated bool) {
	c.haveFocus = activated
	if activated {
		c.border.SetStyle(c.selectedStyle)
	} else {
		c.border.SetStyle(c.regularStyle)
	}
}

// IsSelectable returns true if the menu is selectable
func (c *ctMenu) IsSelectable() bool {
	return true
}
