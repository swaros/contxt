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
)

const (
	TableStart = "<table"
	TableEnd   = "</table>"
	RowStart   = "<row>"
	RowEnd     = "</row>"
	TabStart   = "<tab"
	TabEnd     = "</tab>"
	// defines the size of the cell in relation to the max length of the screen. 10 means 10% of the screen
	AttrSize = "size"
	// defines the char if the cell should fill the remaining space of the cell
	// default is  " " (space)
	ATTR_FILL = "fill"
	// defines how the cell should be aligned
	// 1 = left, 2 = right
	AttrOrigin = "origin"
	// defines how the space is calculated.
	// fixed = if the calculated size is bigger than the cell size, then we will use the cell size
	// extend = fill the rest of the row with the space if prevoius fixed or content cells are smaller than the calculated size
	// content = we will use the max size of the content if they is smaller than the calculated size
	AttrDraw    = "draw"
	DrawFixed   = "fixed"
	DrawExtend  = "extend"
	DrawContent = "content"

	// if the text is cutted, then this string will be added to the end of the text
	AttrCutAdd = "cut-add"
	// this is the mode how the overflow is handled. ignore = the text is ignored, wrap = wrap the text
	AttrOverflow = "overflow"
	// this is the suffix for the cell that will be added to the content all the time. is ment for colorcodes. here and clear code is usually used
	AttrSuffix = "suffix"
	// this is the prefix for the cell that will be placed in front of the content all the time. is ment for colorcodes
	AttrPrefix = "prefix"
	// additonal margin for the cell. this will be subtracted from the cell size
	AttrMargin = "margin"

	// overflow constants
	OfIgnore   = "ignore"
	OfWrap     = "wrap"
	OfWordWrap = "wordwrap"
	OfAny      = "any"

	// origin constants
	OriginLeft  = 0
	OriginRight = 2
)

type TabOut struct {
	table       tableHandle
	tableMode   bool // if we are in the table mode, then we will not render until we get a </table> tag
	rows        []tabCell
	markup      Markup
	info        PostFilterInfo
	RowCalcMode int // 0 = size is relative to with of terminal where maxsize is 100, 1 = absolute size
	calcSize    *roundSerial
}

func NewTabOut() *TabOut {
	return &TabOut{
		markup:   *NewMarkup().SetAccepptedTags([]string{"table", "row", "tab"}),
		calcSize: NewRoundSerial(),
	}
}

// Row functions

// Calculate the size of the cell depending on
// the index of the cell in the row
func (tr *TabOut) getMaxLenByIndex(index int) int {
	max := 0
	for _, row := range tr.table.rows {
		if len, ok := row.getLenByIndex(index); ok {
			if len > max {
				max = len
			}
		}
	}
	return max
}

// Filter is called when the context is updated
// interface fulfills the PostFilter interface
func (t *TabOut) Filter(msg interface{}) interface{} {
	return msg
}

// Update is called when the context is updated
// interface fulfills the PostFilter interface
func (t *TabOut) Update(info PostFilterInfo) {
	t.info = info
	if t.info.IsTerminal {
		if t.RowCalcMode == 0 {
			t.calcSize.SetMax(info.Width)
		}
	}
}

// CanHandleThis returns true if the text is a table
// interface fulfills the PostFilter interface
func (t *TabOut) CanHandleThis(text string) bool {
	return t.IsTable(text) || t.IsRow(text) || t.IsTab(text)
}

// GetRoundTool returns the round tool what is used to calculate the size of the cells
// without rounding errors
func (t *TabOut) GetRoundTool() *roundSerial {
	return t.calcSize
}

// Command is called when the text is a table
// interface fulfills the PostFilter interface
func (t *TabOut) Command(cmd string) string {
	if t.IsTable(cmd) || t.IsRow(cmd) {
		return t.TableParse(cmd)
	} else {
		return cmd
	}
}

func (t *TabOut) Clear() {
	t.rows = []tabCell{}
	t.tableMode = false
	t.table = tableHandle{}
}

func (t *TabOut) GetInfo() PostFilterInfo {
	return t.info
}

func (t *TabOut) ScanForRows(tokens []Parsed) *tableHandle {
	table := NewTableHandle(t)
	t.updateRows(table, tokens)
	return table
}

func (t *TabOut) updateRows(table *tableHandle, tokens []Parsed) {
	table.rows = append(table.rows, t.ScanForCells(tokens, table))
}

func (t *TabOut) ScanForCells(tokens []Parsed, table *tableHandle) *tabRow {
	tabRow := NewTabRow(table)
	tabCell := NewTabCell(tabRow)
	for index, token := range tokens {
		tSize := VisibleLen(token.Text)
		tabCell.index = index
		if token.IsMarkup {
			if strings.HasPrefix(token.Text, "<tab") {
				tSize = 0 // markup have no text length
				tabCell.fillChar = token.GetProperty(ATTR_FILL, " ").(string)
				tabCell.Size = token.GetProperty(AttrSize, 0).(int)
				tabCell.Origin = token.GetProperty(AttrOrigin, 0).(int)
				tabCell.drawMode = token.GetProperty(AttrDraw, "relative").(string)
				tabCell.cutNotifier = token.GetProperty(AttrCutAdd, "...").(string)
				tabCell.overflowMode = token.GetProperty(AttrOverflow, OfIgnore).(string)
				tabCell.anySuffix = token.GetProperty(AttrSuffix, "").(string)
				tabCell.anyPrefix = token.GetProperty(AttrPrefix, "").(string)
				tabCell.margin = token.GetProperty(AttrMargin, 0).(int)
			} else if strings.HasPrefix(token.Text, "</tab>") {
				t.rows = append(t.rows, *tabCell)
				tabCell = NewTabCell(tabRow)

			}
		} else {
			tabCell.Text = token.Text
			tabRow.Cells = append(tabRow.Cells, tabCell)
			tabCell = NewTabCell(tabRow)
		}
		tabRow.maxLengths = append(tabRow.maxLengths, tSize) // get the maximum length of the row
	}
	return tabRow
}

// IsTable returns true if the text is a table
func (t *TabOut) IsTable(text string) bool {
	return strings.HasPrefix(text, TableStart) || strings.HasSuffix(text, TableEnd)
}

// IsRow returns true if the text is a row
func (t *TabOut) IsRow(text string) bool {
	return strings.HasPrefix(text, RowStart) && strings.HasSuffix(text, RowEnd)
}

// IsTab returns true if the text is a tab cell
func (t *TabOut) IsTab(text string) bool {
	return strings.HasPrefix(text, TabStart) && strings.HasSuffix(text, TabEnd)
}

// GetSize returns the size of the cell
// if the cell is relative to the terminal width, the size is calculated
// if the cell is absolute, the size is returned
func (t *TabOut) GetSize(orig int) int {
	if t.info.IsTerminal {
		if t.RowCalcMode == 0 { // relative to terminal width
			if orig > 100 {
				orig = 100
			}
			// taking care about rounding errors
			intRes := t.calcSize.Round(orig)
			return intRes
		}
	}
	return orig
}

// dirty fix for empty tags. any markup must followed by a non markup tag,
// or it will be just ignored. instead of rewriting the whole parser, we just fix it here
// by adding a space between the tags
func (t *TabOut) fixEmptyTags(text string) string {
	return strings.ReplaceAll(text, "></tab>", "> </tab>")
}

// TableParse parses a table
// If a Table is created, we also enters table mode.
// In this mode the created table is not rendered until the table is closed.
func (t *TabOut) TableParse(text string) string {
	text = t.fixEmptyTags(text)
	if t.IsTable(text) {
		if t.tableMode {
			if strings.HasPrefix(text, TableStart) {
				return "" // we are already in table mode, so we return nothing
			}
			if strings.HasPrefix(text, TableEnd) {
				t.tableMode = false
				return t.table.Render()
			}

			if strings.HasSuffix(text, TableEnd) {
				t.tableMode = false
				t.updateRows(&t.table, t.markup.Parse(text))
				return t.table.Render()
			}
		} else {

			t.tableMode = true
			tokens := t.markup.Parse(text)
			tableSlices, outers := t.markup.BuildInnerSlice(tokens, "table")
			t.table = *t.ScanForRows(tableSlices)

			// look for a table end in then outer slice
			for _, outer := range outers {
				if outer.IsMarkup && strings.HasPrefix(outer.Text, "</table>") {
					t.tableMode = false
					return t.table.Render()
				}
			}
			return "" // we are in table mode, but no table end found. so we return nothing
		}
		return ""

	} else if t.IsRow(text) { // row mode. so we do not in the table mode
		if t.tableMode {
			t.updateRows(&t.table, t.markup.Parse(text))
			return ""
		}
		return t.RowParse(text)
	} else {
		return text
	}
}

// similar to TableParse, but for a single row
// and we do not wait for a table end
func (t *TabOut) RowParse(text string) string {
	if t.IsRow(text) {
		tokens := t.markup.Parse(text)
		rowSlices, _ := t.markup.BuildInnerSlice(tokens, "row")
		t.table = *t.ScanForRows(rowSlices)
		return t.table.Render()
	} else {
		return text
	}
}

// PadStrLeft is a shortcut for PadString on a cell
func PadStrLeft(line string, max int, fillChar string) string {
	cell := NewTabCell(nil)
	cell.SetText(line).SetFillChar(fillChar).SetOrigin(0)
	return cell.CutString(max)
}

// PadStrRight is a shortcut for PadString on a cell
func PadStrRight(line string, max int, fillChar string) string {
	cell := NewTabCell(nil)
	cell.SetText(line).SetFillChar(fillChar).SetOrigin(2)
	return cell.CutString(max)
}

func GetLastEscapeSequence(text string) string {
	if len(text) == 0 {
		return ""
	}
	runes := []rune(text)
	lastEscape := 0
	for i, r := range runes {
		if r == 27 {
			lastEscape = i
		}
	}
	if lastEscape == 0 {
		return ""
	}
	return string(runes[lastEscape:])
}
