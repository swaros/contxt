package runner

import (
	"sort"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type TcElement interface {
	// Draw draw the element on the screen
	Draw(s tcell.Screen)
	// MouseReleaseEvent is called when mouse is released
	MouseReleaseEvent(start position, end position, trigger int)
	// MousePressEvent is called when mouse is pressed
	MousePressEvent(pos position, trigger int)
	// MouseHoverEvent is called when mouse is hovering
	MouseHoverEvent(pos position)
	// MouseLeaveEvent is called when mouse is leaving
	MouseLeaveEvent()
	// Hit check if the element is hit by the mouse
	Hit(pos position) bool
	// SetFocus set the focus of the element
	Focus(activated bool)
	// reports if the element is selectable and also can be focused
	IsSelectable() bool
}

var (
	// contains all elements
	elements sync.Map
	// is the last element that was hovered
	LastHoverElement TcElement
	// is the last element that was clicked
	LastMouseElement TcElement
	// is the element that has the focus
	FocusedElement TcElement

	// we need an incemental ID that we can use to identify the elements
	// this is used to identify the elements in the map
	ElementLastID int

	// caching the sortes ids
	sortedElementsCache []TcElement

	// messure the time we need to render the screen
	// this is only used for debugging
	RenderTime time.Duration
)

// Draws all elements
func (c *ctCell) DrawAll() {
	// recorde time for debugging
	start := time.Now()
	// we draw all elements in the order of their z-index
	// we need to sort the elements by their z-index
	elements := c.GetSortedElements()
	for _, element := range elements {
		element.Draw(c.screen)
	}
	// recorde time for debugging
	RenderTime = time.Since(start)
	c.AddDebugMessage("Render time: " + RenderTime.String())

}

// SortedCallBack will call the callback function for all elements
// the elements are sorted by their z-index
func (c *ctCell) SortedCallBack(doIt func(b *TcElement) bool) {
	elements := c.GetSortedElements()
	for _, element := range elements {
		if !doIt(&element) {
			break
		}
	}
}

// MousePressAll is called when the mouse is pressed
// it will trigger the first element that is hit
func (c *ctCell) MousePressAll(pos position, trigger int) {

	c.SortedCallBack(func(b *TcElement) bool {
		if (*b).Hit(pos) {
			(*b).MousePressEvent(pos, trigger)
			// we only want to trigger the first element
			LastMouseElement = (*b)
			return false
		}
		return true
	})

}

// MouseReleaseAll is called when the mouse is released
// it will trigger the first element that is hit by the start coordinate
func (c *ctCell) MouseReleaseAll(start position, end position, trigger int) {

	c.SortedCallBack(func(b *TcElement) bool {
		if (*b).Hit(start) {
			(*b).MouseReleaseEvent(start, end, trigger)
			// we only want to trigger the first element
			LastMouseElement = (*b)
			return false
		}
		return true
	})
}

func (c *ctCell) MouseHoverAll(pos position) {
	var nextHoverElement TcElement

	c.SortedCallBack(func(b *TcElement) bool {
		if (*b).Hit(pos) {
			nextHoverElement = (*b)
			return false
		}
		return true
	})

	if nextHoverElement != nil && LastHoverElement != nextHoverElement {
		if LastHoverElement != nil {
			LastHoverElement.MouseLeaveEvent()
		}
		nextHoverElement.MouseHoverEvent(pos)
		LastHoverElement = nextHoverElement
	} else if nextHoverElement == nil && LastHoverElement != nil {
		LastHoverElement.MouseLeaveEvent()
		LastHoverElement = nil
	}
}

func (c *ctCell) CycleFocus() {
	var nextFocusElement TcElement
	var found bool

	c.SortedCallBack(func(b *TcElement) bool {
		if found {
			if (*b).IsSelectable() {
				nextFocusElement = (*b)
				return false
			}
		}
		if (*b) == FocusedElement {
			found = true
		}
		return true

	})

	if nextFocusElement == nil {
		elements.Range(func(key, value interface{}) bool {
			nextFocusElement = value.(TcElement)
			return false
		})
	}
	c.SetFocus(nextFocusElement)
}

// SetFocus set the focus of the element
// the old focus element will be unfocused
func (c *ctCell) SetFocus(elem TcElement) {
	if FocusedElement != nil {
		FocusedElement.Focus(false)
	}
	FocusedElement = elem
	FocusedElement.Focus(true)
}

func (c *ctCell) GetSortedKeys() []int {
	var sortedKeys []int
	elements.Range(func(key, value interface{}) bool {
		sortedKeys = append(sortedKeys, key.(int))
		return true
	})
	sort.Ints(sortedKeys)
	return sortedKeys
}

func (c *ctCell) GetSortedElements() []TcElement {
	if len(sortedElementsCache) > 0 {
		return sortedElementsCache
	}

	var sortedElements []TcElement
	keys := c.GetSortedKeys()
	for _, key := range keys {
		if v, ok := elements.Load(key); ok {
			sortedElements = append(sortedElements, v.(TcElement))
		}
	}
	sortedElementsCache = sortedElements

	return sortedElements
}

func (c *ctCell) GetElementByID(id int) TcElement {
	if v, ok := elements.Load(id); ok {
		return v.(TcElement)
	}
	return nil
}

func (c *ctCell) AddElement(e TcElement) int {
	c.ResetCaches()
	ElementLastID++
	elements.Store(ElementLastID, e)
	return ElementLastID
}

func (c *ctCell) RemoveElement(e TcElement) {
	elements.Range(func(key, value interface{}) bool {
		if value == e {
			elements.Delete(key)
			c.ResetCaches()
			return false
		}
		return true
	})
}

func (c *ctCell) ResetCaches() {
	sortedElementsCache = nil
}

func (c *ctCell) RemoveElementByID(id int) {
	elements.Delete(id)
	c.ResetCaches()
}

func (c *ctCell) ClearElements() {
	elements.Range(func(key, value interface{}) bool {
		elements.Delete(key)
		return true
	})
	c.ResetCaches()
}
