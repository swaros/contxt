package runner

import (
	"math/rand"
	"sync"
)

// solving the problem of having different color combination for each task that is running.
// so any task can identify itself by a color.
// these colors have to readable in the combination of foreground and background.
// also we store the used colors per run in memory, with the name of the task as key.

type RandColor struct {
	foreroundColor  string
	backgroundColor string
}

type RandColorStore struct {
	colors sync.Map
}

var (
	looksGoodCombinations = []RandColor{
		{foreroundColor: "black", backgroundColor: "green"},
		{foreroundColor: "black", backgroundColor: "yellow"},
		{foreroundColor: "black", backgroundColor: "blue"},
		{foreroundColor: "black", backgroundColor: "magenta"},
		{foreroundColor: "black", backgroundColor: "cyan"},
		{foreroundColor: "black", backgroundColor: "white"},
		{foreroundColor: "white", backgroundColor: "green"},
		{foreroundColor: "white", backgroundColor: "yellow"},
		{foreroundColor: "white", backgroundColor: "blue"},
		{foreroundColor: "white", backgroundColor: "magenta"},
		{foreroundColor: "white", backgroundColor: "cyan"},
		{foreroundColor: "white", backgroundColor: "black"},
		{foreroundColor: "green", backgroundColor: "black"},
		{foreroundColor: "yellow", backgroundColor: "black"},
		{foreroundColor: "blue", backgroundColor: "black"},
		{foreroundColor: "magenta", backgroundColor: "black"},
		{foreroundColor: "cyan", backgroundColor: "black"},
		{foreroundColor: "white", backgroundColor: "dark-grey"},
	}
)

func NewRandColorStore() *RandColorStore {
	return &RandColorStore{
		colors: sync.Map{},
	}
}

func PickRandColor() RandColor {

	return looksGoodCombinations[rand.Intn(len(looksGoodCombinations))]
}

func (rcs *RandColorStore) GetColor(taskName string) RandColor {
	if col, ok := rcs.colors.Load(taskName); ok {
		return col.(RandColor)
	}

	col := PickRandColor()
	rcs.colors.Store(taskName, col)
	return col
}

func (rcs *RandColorStore) GetColorAsCtxMarkup(taskName string) (string, string, string) {
	col := rcs.GetColor(taskName)
	fg := "<f:" + col.foreroundColor + ">"
	bg := "<b:" + col.backgroundColor + ">"
	sc := "<f:" + col.backgroundColor + ">"
	return fg, bg, sc
}
