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
	ATTR_SIZE = "size"
	// defines the char if the cell should fill the remaining space of the cell
	// default is  " " (space)
	ATTR_FILL = "fill"
	// defines how the cell should be aligned
	// 1 = left, 2 = right
	ATTR_ORIGIN = "origin"
	// defines how the space is calculated.
	// fixed = if the calculated size is bigger than the cell size, then we will use the cell size
	// extend = fill the rest of the row with the space if prevoius fixed or content cells are smaller than the calculated size
	// content = we will use the max size of the content if they is smaller than the calculated size
	ATTR_DRAW = "draw"
	// if the text is cutted, then this string will be added to the end of the text
	ATTR_CUT_ADD = "cut-add"
	// this is the mode how the overflow is handled. ignore = the text is ignored, wrap = wrap the text
	ATTR_OVERFLOW = "overflow"
	// this is the suffix for the cell that will be added to the content all the time. is ment for colorcodes. here and clear code is usually used
	ATTR_SUFFIX = "suffix"
	// this is the prefix for the cell that will be placed in front of the content all the time. is ment for colorcodes
	ATTR_PREFIX = "prefix"
	// additonal margin for the cell. this will be subtracted from the cell size
	ATTR_MARGIN = "margin"
	// the content will be shown, cuttet and alligned depending the origin value
	OVERFLOW_ANY = "any"
	// we will keep the content and wrap it. this will increase the height of the row if we need more space
	OVERFLOW_WORDWRAP = "wordwrap"
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
		markup:   *NewMarkup(),
		calcSize: NewRoundSerial(),
	}
}

// Row functions

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
		tSize := LenPrintable(token.Text)
		tabCell.index = index
		if token.IsMarkup {
			if strings.HasPrefix(token.Text, "<tab") {
				tSize = 0 // markup have no text length
				tabCell.fillChar = token.GetProperty(ATTR_FILL, " ").(string)
				tabCell.Size = token.GetProperty(ATTR_SIZE, 0).(int)
				tabCell.Origin = token.GetProperty(ATTR_ORIGIN, 0).(int)
				tabCell.drawMode = token.GetProperty(ATTR_DRAW, "relative").(string)
				tabCell.cutNotifier = token.GetProperty(ATTR_CUT_ADD, "...").(string)
				tabCell.overflowMode = token.GetProperty(ATTR_OVERFLOW, "ignore").(string)
				tabCell.anySuffix = token.GetProperty(ATTR_SUFFIX, "").(string)
				tabCell.anyPrefix = token.GetProperty(ATTR_PREFIX, "").(string)
				tabCell.margin = token.GetProperty(ATTR_MARGIN, 0).(int)
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
			//intRes, _, _ := RoundHelp(orig, t.info.Width)
			intRes := t.calcSize.Round(orig)
			return intRes
		}
	}
	return orig
}

// TableParse parses a table
// If a Table is created, we also enters table mode.
// In this mode the created table is not rendered until the table is closed.
func (t *TabOut) TableParse(text string) string {
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
