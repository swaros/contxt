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
