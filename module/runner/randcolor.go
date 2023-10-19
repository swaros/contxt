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
	foreColor string // foreground color as string
	backColor string // background color as string
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
		{foreColor: "white", backColor: "black"},
		{foreColor: "green", backColor: "black"},
		{foreColor: "yellow", backColor: "black"},
		{foreColor: "blue", backColor: "black"},
		{foreColor: "magenta", backColor: "black"},
		{foreColor: "cyan", backColor: "black"},
		{foreColor: "dark-grey", backColor: "black"},
		{foreColor: "light-green", backColor: "black"},
		{foreColor: "light-grey", backColor: "black"},
		{foreColor: "light-red", backColor: "black"},
		{foreColor: "light-blue", backColor: "black"},
		{foreColor: "light-yellow", backColor: "black"},
		{foreColor: "light-cyan", backColor: "black"},
		{foreColor: "light-magenta", backColor: "black"},
		{foreColor: "white", backColor: "dark-grey"},
		{foreColor: "light-green", backColor: "dark-grey"},
		{foreColor: "light-blue", backColor: "dark-grey"},
		{foreColor: "light-yellow", backColor: "dark-grey"},
		{foreColor: "light-cyan", backColor: "dark-grey"},
		{foreColor: "light-magenta", backColor: "dark-grey"},
		{foreColor: "yellow", backColor: "dark-grey"},
		{foreColor: "green", backColor: "dark-grey"},
		{foreColor: "cyan", backColor: "dark-grey"},
		{foreColor: "black", backColor: "green"},
		{foreColor: "blue", backColor: "green"},
		{foreColor: "white", backColor: "green"},
		{foreColor: "magenta", backColor: "green"},
		{foreColor: "black", backColor: "yellow"},
		{foreColor: "black", backColor: "blue"},
		{foreColor: "black", backColor: "magenta"},
		{foreColor: "black", backColor: "cyan"},
		{foreColor: "black", backColor: "white"},
		{foreColor: "white", backColor: "yellow"},
		{foreColor: "white", backColor: "blue"},
		{foreColor: "white", backColor: "magenta"},
		{foreColor: "white", backColor: "cyan"},
		{foreColor: "blue", backColor: "yellow"},
		{foreColor: "blue", backColor: "cyan"},
		{foreColor: "blue", backColor: "white"},
		{foreColor: "yellow", backColor: "blue"},
		{foreColor: "yellow", backColor: "magenta"},
		{foreColor: "yellow", backColor: "cyan"},
		{foreColor: "yellow", backColor: "white"},
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
	return "<f:" + r.foreColor + "><b:" + r.backColor + ">"
}

// ForeColor returns the foreground color as ctxout markup
func (r *RandColor) ForeColor() string {
	return "<f:" + r.foreColor + ">"
}

// BackColor returns the background color as ctxout markup
func (r *RandColor) BackColor() string {
	return "<b:" + r.backColor + ">"
}

// AsSignColor returns the sign color as ctxout markup
func (r *RandColor) AsSignColor() string {
	return "<f:" + r.backColor + ">"
}
