package outlaw

// global variable for selected item for any used list
var selected selectResult

type selectResult struct {
	isSelected bool
	item       selectItem
}
