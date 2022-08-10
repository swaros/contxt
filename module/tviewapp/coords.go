package tviewapp

func (c *CellApp) hitChecker(x, y int, currents []int, runBefore func(int, int), verify func(CElement) bool, runAction func(CElement, int, int)) []int {

	if runBefore != nil {
		runBefore(x, y) // checks for example if a element that we have store, are no longer affected by this coordinates
	}
	switch c.bubbleBehave {
	case BubbleOff, BubbleUp: // here we fire the event to any element that is in range
		for index, el := range c.baseElements { // then lets test all
			if verify != nil && !verify(el) { // we verify first if this element needs to be checked
				continue
			}
			if el.hitTest(x, y) {
				currents = append(currents, index) // keep track of elements that we had in focus
				if runAction != nil {
					runAction(el, x, y)
				}
				if c.bubbleBehave == BubbleUp {
					return currents // if just want the first it, by BubbleUp behavior, we get out now
				}
			}
		}
	case BubbleDown: // here we fire the event on the last one in the list (what is being drawed latest)
		var hitEl CElement = nil
		var hitIndex int = -1
		for _, index := range c.actives { // then lets test all
			if el, _ := c.getElementByIndex(index); el != nil {
				if el.hitTest(x, y) && (verify == nil || verify(el)) {
					hitEl = el
					hitIndex = index
				}
			}
		}
		if hitEl != nil {
			if runAction != nil {
				runAction(hitEl, x, y)
			} // trigger the handler
			currents = append(currents, hitIndex) // save this one
		}
	}
	return currents
}

func (c *CellApp) checkPreviousHitList(x, y int, currents []int, runAction func(CElement, int, int)) []int {
	if len(currents) < 1 { // no stored hits, we get out early
		return currents
	}

	var cleanUp []int // prepare the new list of elements they stiff affected
	switch c.bubbleBehave {
	case BubbleOff, BubbleUp:
		for _, index := range currents {
			if el, ok := c.getElementByIndex(index); ok {

				if !el.hitTest(x, y) {
					if runAction != nil {
						runAction(el, x, y)
					}
				} else {
					cleanUp = append(cleanUp, index) // memorize again this element, becasue it is not affected by hittest
				}
				if c.bubbleBehave == BubbleUp { // in case we just need the first element (again. is visual the element in the background)
					currents = cleanUp // copy the list here already
					return currents    // and get out
				}
			}
		}
	case BubbleDown: // this case is for the elements in front of the ui (visual) only. elements behind them do not count
		found := false
		for i := len(currents) - 1; i >= 0; i-- { // we look from the other way around ..last in map first
			index := currents[i]
			if el, ok := c.getElementByIndex(index); ok {

				if !el.hitTest(x, y) { // this element is even no longer hivered. so we trigger the leave
					if runAction != nil {
						runAction(el, x, y)
					}
				} else {
					if !found { // as long we did not find a hovered element, we keep memorize the elements that still hovered.
						cleanUp = append(cleanUp, index) // but this can be juts one element...
						found = true                     // ...because from now on, we ignore anything else
					} else {
						if runAction != nil { // becasue we found our top element, any other have to leave
							runAction(el, x, y)
						}
					}
				}
			}

		}
	}
	currents = cleanUp
	return currents
}
