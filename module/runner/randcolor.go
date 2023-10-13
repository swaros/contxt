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

package runner

import (
	"math/rand"
	"sync"
)

// solving the problem of having different color combination for each task that is running.
// so any task can identify itself by a color.
// these colors have to readable in the combination of foreground and background.
// also we store the used colors per run in memory, with the name of the task as key.
// the colors are just the key-names from the ctxout color table.
// to use them in ctxout we have to wrap them in the ctxout markup.
// the color combinations are picked from the looksGoodCombinations list.
// the list is not complete, but it is a good start.

// example:
// randColor := runner.NewRandColorStore()
// colorPicked := randColor.GetColor("myTask")
// ctxout.PrintLn(ctxout.NewMOWrap(), colorPicked.ColorMarkup(), "myTask", ctxout.CleanTag)

var (
	lastInstace *RandColorStore
	mu          sync.Mutex
)

type RandColor struct {
	foreroundColor  string // foreground color as string
	backgroundColor string // background color as string
}

type RandColorStore struct {
	colors    sync.Map
	lastIndex int
}

// LastRandColorInstance returns the last instance of RandColorStore
// if no instance exists, a new one is created
// this is a singleton (kind of)
// the reason is just to have a global instance so you have access to the colors
// assigned to an task, from everywhere in the code.
// this logic can fail, of course if you have multiple instances of RandColorStore created.
func LastRandColorInstance() *RandColorStore {
	mu.Lock()
	defer mu.Unlock()
	if lastInstace == nil {
		lastInstace = NewRandColorStore()
	}
	return lastInstace
}

var (
	// looksGoodCombinations is a list of color combinations that are readable
	looksGoodCombinations = []RandColor{
		{foreroundColor: "white", backgroundColor: "black"},
		{foreroundColor: "green", backgroundColor: "black"},
		{foreroundColor: "yellow", backgroundColor: "black"},
		{foreroundColor: "blue", backgroundColor: "black"},
		{foreroundColor: "magenta", backgroundColor: "black"},
		{foreroundColor: "cyan", backgroundColor: "black"},
		{foreroundColor: "dark-grey", backgroundColor: "black"},
		{foreroundColor: "light-green", backgroundColor: "black"},
		{foreroundColor: "light-grey", backgroundColor: "black"},
		{foreroundColor: "light-red", backgroundColor: "black"},
		{foreroundColor: "light-blue", backgroundColor: "black"},
		{foreroundColor: "light-yellow", backgroundColor: "black"},
		{foreroundColor: "light-cyan", backgroundColor: "black"},
		{foreroundColor: "light-magenta", backgroundColor: "black"},
		{foreroundColor: "white", backgroundColor: "dark-grey"},
		{foreroundColor: "light-green", backgroundColor: "dark-grey"},
		{foreroundColor: "light-blue", backgroundColor: "dark-grey"},
		{foreroundColor: "light-yellow", backgroundColor: "dark-grey"},
		{foreroundColor: "light-cyan", backgroundColor: "dark-grey"},
		{foreroundColor: "light-magenta", backgroundColor: "dark-grey"},
		{foreroundColor: "yellow", backgroundColor: "dark-grey"},
		{foreroundColor: "green", backgroundColor: "dark-grey"},
		{foreroundColor: "cyan", backgroundColor: "dark-grey"},
		{foreroundColor: "black", backgroundColor: "green"},
		{foreroundColor: "blue", backgroundColor: "green"},
		{foreroundColor: "white", backgroundColor: "green"},
		{foreroundColor: "magenta", backgroundColor: "green"},
		{foreroundColor: "black", backgroundColor: "yellow"},
		{foreroundColor: "black", backgroundColor: "blue"},
		{foreroundColor: "black", backgroundColor: "magenta"},
		{foreroundColor: "black", backgroundColor: "cyan"},
		{foreroundColor: "black", backgroundColor: "white"},
		{foreroundColor: "white", backgroundColor: "yellow"},
		{foreroundColor: "white", backgroundColor: "blue"},
		{foreroundColor: "white", backgroundColor: "magenta"},
		{foreroundColor: "white", backgroundColor: "cyan"},
		{foreroundColor: "blue", backgroundColor: "yellow"},
		{foreroundColor: "blue", backgroundColor: "cyan"},
		{foreroundColor: "blue", backgroundColor: "white"},
		{foreroundColor: "yellow", backgroundColor: "blue"},
		{foreroundColor: "yellow", backgroundColor: "magenta"},
		{foreroundColor: "yellow", backgroundColor: "cyan"},
		{foreroundColor: "yellow", backgroundColor: "white"},
	}
)

// NewRandColorStore creates a new RandColorStore
// and stores the last instance in a global variable
func NewRandColorStore() *RandColorStore {
	rn := &RandColorStore{
		colors:    sync.Map{},
		lastIndex: 0,
	}
	lastInstace = rn
	return rn
}

// PickRandColor picks a random color combination from the looksGoodCombinations list
func PickRandColor() RandColor {
	return looksGoodCombinations[rand.Intn(len(looksGoodCombinations))]
}

// PickRandColorByIndex picks a color combination from the looksGoodCombinations list
// by the given index
// if the index is out of range, the first color combination is returned
func PickRandColorByIndex(index int) RandColor {
	// should not happen but just in case someone removes all colors
	if len(looksGoodCombinations) == 0 {
		panic("no color combinations available")
	}
	if index > len(looksGoodCombinations) {
		index = 0
	}
	return looksGoodCombinations[index]
}

// PickUnusedRandColor picks a random color combination from the looksGoodCombinations list
// that is not in usage.
// if all colors are in usage, the first color combination is returned
// so there is no guarantee that you get an unused color combination.
// the second return value is true if the color combination is unused
// the second return value is false if the color combination is in usage
func (rcs *RandColorStore) PickUnusedRandColor() (RandColor, bool) {
	tries := 0
	for {
		color := PickRandColor()
		if !rcs.IsInusage(color) {
			return color, true
		}
		tries++
		// if we failed after 10 times * max variants, we just return the color
		if tries > rcs.GetMaxVariants()*10 {
			return color, false
		}
	}
}

// GetOrSetIndexColor returns a color combination for the given taskName
func (rcs *RandColorStore) GetOrSetIndexColor(taskName string) RandColor {
	if col, ok := rcs.colors.Load(taskName); ok {
		return col.(RandColor)
	}
	col := PickRandColorByIndex(rcs.lastIndex)
	rcs.colors.Store(taskName, col)
	rcs.lastIndex++
	return col
}

// GetOrSetIndexColor returns a color combination for the given taskName
// this color is randomly picked from the looksGoodCombinations list
// if the color is not already in usage, the second return value is false
// if the color is already in usage, the second return value is true
// this is also the case,if we ask for a color that is already stored
func (rcs *RandColorStore) GetOrSetRandomColor(taskName string) (RandColor, bool) {
	if col, ok := rcs.colors.Load(taskName); ok {
		return col.(RandColor), true
	}
	col, inUse := rcs.PickUnusedRandColor()
	rcs.colors.Store(taskName, col)
	rcs.lastIndex++
	return col, inUse
}

// GetMaxVariants returns the number of color combinations in the looksGoodCombinations list
func (rcs *RandColorStore) GetMaxVariants() int {
	return len(looksGoodCombinations)
}

// IsInusage returns true if the given color is already in usage
func (rcs *RandColorStore) IsInusage(color RandColor) bool {
	var inusage bool
	rcs.colors.Range(func(key, value interface{}) bool {
		if value == color {
			inusage = true
			return false
		}
		return true
	},
	)
	return inusage
}

// GetColorAsCtxMarkup returns the color combination for the given taskName
// as ctxout markup
// the first return value is the foreground color as markup
// the second return value is the background color as markup
// the third return value is the sign color as markup. this simply the foreground color
// used as background color, so cou can draw a sign with the foreground color.
func (rcs *RandColorStore) GetColorAsCtxMarkup(taskName string) (string, string, string) {
	col, _ := rcs.GetOrSetRandomColor(taskName)
	return col.ForeColor(), col.BackColor(), col.AsSignColor()
}

// ColorMarkup returns the color combination as ctxout markup
func (r *RandColor) ColorMarkup() string {
	return "<f:" + r.foreroundColor + "><b:" + r.backgroundColor + ">"
}

// ForeColor returns the foreground color as ctxout markup
func (r *RandColor) ForeColor() string {
	return "<f:" + r.foreroundColor + ">"
}

// BackColor returns the background color as ctxout markup
func (r *RandColor) BackColor() string {
	return "<b:" + r.backgroundColor + ">"
}

// AsSignColor returns the sign color as ctxout markup
func (r *RandColor) AsSignColor() string {
	return "<f:" + r.backgroundColor + ">"
}
