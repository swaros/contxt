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

package ctxout

import (
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

// tabCell is a single cell in a row
type tabCell struct {
	Size                 int
	Origin               int
	OriginString         string
	Text                 string
	parent               *tabRow
	fillChar             string
	index                int    // reference to the index in the parent row
	drawMode             string // fixed = fixed size, relative = relative to terminal size, content = max size of content
	cutNotifier          string // if the text is cutted, then this string will be added to the end of the text
	overflow             bool   // if the text is cutted, then this will be set to true
	overflowContent      string // if the text is cutted, then this will be set to the cutted content
	overflowMode         string // this is the mode how the overflow is handled. ignore = the text is ignored, wrap = wrap the text
	anyPrefix, anySuffix string // this is the prefix and postfix for the cell that will be added all the time. is ment for colorcodes
	margin               int    // this is the margin for the cell. that means an absolute amout of space that will subtracted from the cell size
}

func NewTabCell(parent *tabRow) *tabCell {
	return &tabCell{
		Size:            0, // 0 = auto
		Origin:          0, // 0 left, 1 left cutted but align to left, 2 right
		Text:            "",
		parent:          parent,
		fillChar:        " ",
		cutNotifier:     " ...",
		overflow:        false,
		overflowMode:    "ignore",
		overflowContent: "",
	}
}

func (td *tabCell) SetSize(size int) *tabCell {
	td.Size = size
	return td
}

func (td *tabCell) SetOrigin(origin int) *tabCell {
	td.Origin = origin
	return td
}

func (td *tabCell) SetOriginString(origin string) *tabCell {
	td.OriginString = origin
	return td
}

func (td *tabCell) SetText(text string) *tabCell {
	td.Text = text
	return td
}

func (td *tabCell) SetFillChar(fillChar string) *tabCell {
	td.fillChar = fillChar
	return td
}

func (td *tabCell) SetDrawMode(drawMode string) *tabCell {
	td.drawMode = drawMode
	return td
}

func (td *tabCell) SetCutNotifier(cutNotifier string) *tabCell {
	td.cutNotifier = cutNotifier
	return td
}

func (td *tabCell) SetOverflowMode(overflowMode string) *tabCell {
	td.overflowMode = overflowMode
	td.autoFixOverflow()
	return td
}

func (td *tabCell) autoFixOverflow() {
	if td.overflowMode == OfAny {
		td.overflowMode = OfWrap
		return
	}
	if td.overflowMode != OfWordWrap && td.overflowMode != OfWrap && td.overflowMode != OfIgnore {
		td.overflowMode = OfIgnore
	}
}

func (td *tabCell) SetIndex(index int) *tabCell {
	td.index = index
	return td
}

func (td *tabCell) GetOverflow() bool {
	return td.overflow
}

func (td *tabCell) GetOverflowContent() string {
	return td.overflowContent
}

func (td *tabCell) GetText() string {
	return td.anyPrefix + td.Text + td.anySuffix
}

func (td *tabCell) GetSize() int {
	return td.Size
}

func (td *tabCell) GetOrigin() int {
	return td.Origin
}

// Copy returns a copy of the cell
func (td *tabCell) Copy() *tabCell {
	newCell := NewTabCell(td.parent)
	newCell.fillChar = td.fillChar
	newCell.Size = td.Size
	newCell.Text = td.Text
	newCell.Origin = td.Origin
	newCell.drawMode = td.drawMode
	newCell.cutNotifier = td.cutNotifier
	newCell.overflowMode = td.overflowMode
	newCell.overflow = td.overflow
	newCell.overflowContent = td.overflowContent
	newCell.margin = td.margin
	return newCell
}

// MoveToWrap moves the cell to the wrap mode and resets the text
// but also it updates the content for cells that are not in wrap mode
// so they can be drawed correctly by the row
func (td *tabCell) MoveToWrap() bool {
	if td.overflow && td.overflowContent != "" {
		td.Text = td.overflowContent
		td.overflowContent = ""
		td.overflow = false
		return true
	}
	td.Text = "" // also reset the text for cells that are not in overflow mode
	td.overflow = false
	return false
}

// addNotifiers adds the notifier to the text if we have enough space
// here it is not about detecting the overflow, but just to add the notifier.
// the detection of the overflow, and reduce the string to the expected length, must be done before.
// we expect an string that is already cutted to the expected length
func (td *tabCell) addNotifiers(str string, applyToProp bool) string {
	if td.cutNotifier != "" && VisibleLen(td.cutNotifier) < VisibleLen(str) {
		// we have enough space for applying the notifier
		switch td.Origin {
		case OriginRight:
			str = td.cutNotifier + str[VisibleLen(td.cutNotifier):]
		default:
			str = str[:VisibleLen(str)-VisibleLen(td.cutNotifier)] + td.cutNotifier
		}
	}
	if applyToProp {
		td.Text = str
	}
	return str
}

func (td *tabCell) WrapText(max int) (text string, overflow string) {
	// add margin if we have one and there are higher than 0.
	// tis reduces the max size of the cell
	if td.margin > 0 {
		max = max - td.margin
	}
	// here we get the text and the overflow content if we have any newline in the text
	// so the firstline is the text until the first newline
	// and the overflowText is the rest of the text including the newlines
	firstLine, overflowText := td.GetNewLineContext()
	// do we have an overflow?
	size := VisibleLen(firstLine)
	theIfWeHaveOverflowFlag := 0 // asume we have no overflow
	if size > max {
		theIfWeHaveOverflowFlag = 1 // we have an overflow. so the current text have to be cutted
	} else if size == max {
		theIfWeHaveOverflowFlag = 2 // we hit the excact size
	}

	switch td.overflowMode {

	// ignore the overflow and just fit the string to the max size
	// if the string is bigger than the max size, then it will be cutted
	// and the rest of the string will NOT be added to the overflow, because thats the ignore mode
	case OfIgnore:

		overflowText = ""
		td.overflow = false

		switch theIfWeHaveOverflowFlag {
		case 0: // no overflow
			td.Text = td.fillString(firstLine, max)
		case 1: // overflow. so the firstline is bigger than the max. we need to cut it, and add the rest of the text to the overflow
			var textr string
			// check again if we do not have the ignore case, so we do not handle overflow
			switch td.Origin {
			case OriginRight:
				textr, _ = StringCutFromRight(firstLine, max) // on ignore, we do not care about overflow
			default:
				textr, _ = StringCut(firstLine, max) // on ignore, we do not care about overflow
			}

			td.addNotifiers(textr, true)

		case 2: // excact size
			td.Text = td.fillString(firstLine, max)
		}

	case OfWrap:
		// wrap the text but to not care about wrapping by words
		wrapStr := wrap.String(firstLine, max)
		reslize, restStr := getNlSlice(wrapStr)
		td.Text = td.fillString(reslize, max)
		overflowText = restStr + overflowText
		td.overflow = true

	case OfWordWrap:
		// this is mostly the case for the first line
		if theIfWeHaveOverflowFlag == 1 {
			td.Text = FitWordsToMaxLen(td.Text, max)
			wrapped := wrap.String(wordwrap.String(td.Text, max), max)
			overflowText = ""

			// but now we have also taking care about the newlines into the text
			// of the first line
			clean, afterNl := getNlSlice(wrapped)
			// just fill the string with the fillChar
			td.Text = td.fillString(clean, max)

			// now lets proceed with the overflow. we have to add the rest of the text to the overflow
			// if we have any
			if afterNl != "" {
				// this is the case, there was overflow before, so we add the rest of the text to the overflow
				if overflowText != "" {
					overflowText = afterNl + " " + overflowText
				} else {
					// this is the case, there was no overflow before, so we set them now
					overflowText = afterNl
				}
			}
		} else {
			// the firstline is not bigger than the max size
			// so we just need to fill the string
			// the overflow is already handled by the overflowText
			td.Text = td.fillString(firstLine, max)
		}
		td.overflow = true
	}
	td.overflowContent = overflowText
	return td.Text, td.overflowContent
}

func (td *tabCell) GetNewLineContext() (string, string) {
	containsNewLine := strings.Contains(td.Text, "\n")
	if containsNewLine {
		return getNlSlice(td.Text)
	}
	return td.Text, ""
}

func getNlSlice(text string) (string, string) {
	containsNewLine := strings.Contains(text, "\n")
	if containsNewLine {
		rows := strings.Split(text, "\n")
		return rows[0], strings.Join(rows[1:], "\n")
	}
	return text, ""

}

// CutString cuts the string to the given max size
// if the string is bigger than the max size, then it will be cutted
// if the string is smaller than the max size, then it will be filled up with the fillChar
func (td *tabCell) CutString(max int) string {
	td.WrapText(max)
	return td.Text
}

func (td *tabCell) fillString(text string, max int) string {
	tSize := VisibleLen(text)
	diff := max - tSize
	if diff < 1 {
		return text
	}
	switch td.Origin {
	default:
		text = text + strings.Repeat(td.fillChar, diff)
	case 2:
		text = strings.Repeat(td.fillChar, diff) + text
	}
	return text
}
