package ctxout

import (
	"strings"

	"github.com/swaros/contxt/module/systools"
)

const (
	TableStart = "<table"
	TableEnd   = "</table>"
	RowStart   = "<row>"
	RowEnd     = "</row>"
	TabStart   = "<tab"
	TabEnd     = "</tab>"
)

type tableHandle struct {
	rows             []tabRow
	parent           *TabOut
	rowSeperator     string
	sizeCalculations map[int]int
}

// tabRow is a single row in a table
type tabRow struct {
	Cells        []tabCell
	parent       *tableHandle
	rowEndString string
	maxLengths   []int
}

// tabCell is a single cell in a row
type tabCell struct {
	Size         int
	Origin       int
	OriginString string
	Text         string
	parent       *tabRow
	fillChar     string
	index        int    // reference to the index in the parent row
	drawMode     string // fixed = fixed size, relative = relative to terminal size, content = max size of content
}

type TabOut struct {
	table       tableHandle
	tableMode   bool // if we are in the table mode, then we will not render until we get a </table> tag
	rows        []tabCell
	markup      Markup
	info        PostFilterInfo
	RowCalcMode int // 0 = size is relative to with of terminal where maxsize is 100, 1 = absolute size
}

func NewTabOut() *TabOut {
	return &TabOut{
		markup: *NewMarkup(),
	}
}

func NewTabCell(parent *tabRow) *tabCell {
	return &tabCell{
		Size:     0, // 0 = auto
		Origin:   0, // 0 left, 1 center, 2 right
		Text:     "",
		parent:   parent,
		fillChar: " ",
	}
}

func NewTabRow(parent *tableHandle) *tabRow {
	return &tabRow{
		Cells:        []tabCell{},
		parent:       parent,
		rowEndString: "",
	}
}

func NewTableHandle(parent *TabOut) *tableHandle {
	return &tableHandle{
		rows:             []tabRow{},
		parent:           parent,
		rowSeperator:     "\n",
		sizeCalculations: make(map[int]int),
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

func (tr *tabRow) getLenByIndex(index int) (int, bool) {
	if len(tr.maxLengths) > index {
		return tr.maxLengths[index], true
	}
	return 0, false
}

func (tr *tabRow) GetSize(cell *tabCell, index int) int {
	orig := cell.Size
	if tr.parent.parent.GetInfo().IsTerminal {
		if tr.parent.parent.RowCalcMode == 0 { // relative to terminal width
			calculatedSize := tr.parent.parent.GetSize(cell.Size)
			if orig > 100 {
				orig = 100
			}

			switch cell.drawMode {
			case "fixed": // fixed size. if the calculated size is bigger than the cell size, then we will use the cell size
				if calculatedSize > cell.Size {
					diff := calculatedSize - cell.Size
					if diff > 0 {
						tr.parent.SetSizeCalculation(cell.index, diff) // store the difference so we can add it to the next cell
						return cell.Size
					}
				}
				return calculatedSize
			case "extend": // fill the rest of the row with the space if prevoius fixed or content cells are smaller than the calculated size
				fillSize := tr.parent.GetSumSize(cell.index)
				if fillSize > 0 {
					return calculatedSize + fillSize
				}
				return calculatedSize
			case "content": // we will use the max size of the content if they is smaller than the calculated size
				contentSize := tr.parent.parent.getMaxLenByIndex(cell.index)
				diff := calculatedSize - contentSize
				if diff > 0 {
					tr.parent.SetSizeCalculation(cell.index, diff) // store the difference so we can add it to the next cell
					return contentSize
				}

				return calculatedSize
			default:
				return calculatedSize
			}
		}
	}

	return cell.Size
}

func (tb *tableHandle) SetSizeCalculation(index int, size int) {
	tb.sizeCalculations[index] = size
}

func (tb *tableHandle) GetSize(index int) (int, bool) {
	if size, ok := tb.sizeCalculations[index]; ok {
		return size, true
	}
	return 0, false
}

func (tb *tableHandle) GetSumSize(untilIndex int) int {
	sum := 0
	for indx, size := range tb.sizeCalculations {
		if indx > untilIndex {
			break
		}
		sum += size
	}
	return sum
}

func (tr *tabRow) Render() string {
	if len(tr.Cells) == 0 {
		return ""
	}
	var result []string
	for indx, cell := range tr.Cells {
		if cell.Size > 0 {
			rowSize := tr.GetSize(&cell, indx)
			size := tr.parent.parent.GetSize(cell.Size)
			if rowSize > 0 {
				size = rowSize
			}
			switch cell.Origin {
			case 0: // left padding
				result = append(result, PadString(cell.Text, size, cell.fillChar))
			case 1:
				tempStr := systools.StringSubRight(cell.Text, size)
				result = append(result, PadString(tempStr, size, cell.fillChar))
			case 2:
				result = append(result, PadStringToRight(cell.Text, size, cell.fillChar))
			}
		} else {
			result = append(result, cell.Text)
		}
	}
	return strings.Join(result, "") + tr.rowEndString
}

func (tb *tableHandle) Render() string {
	var result []string
	for _, row := range tb.rows {
		cnt := row.Render()
		if cnt != "" {
			result = append(result, cnt)
		}
	}
	return strings.Join(result, tb.rowSeperator)
}

func (t *TabOut) Filter(msg interface{}) interface{} {
	return msg
}

// Update is called when the context is updated
// interface fulfills the PostFilter interface
func (t *TabOut) Update(info PostFilterInfo) {
	t.info = info
}

// CanHandleThis returns true if the text is a table
// interface fulfills the PostFilter interface
func (t *TabOut) CanHandleThis(text string) bool {
	return t.IsTable(text) || t.IsRow(text) || t.IsTab(text)
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
	t.markup.BuildInnerSliceEach(tokens, "row", func(markup []Parsed) bool {
		row := t.ScanForCells(markup, table)
		table.rows = append(table.rows, *row)
		return true
	})
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
				tabCell.fillChar = token.GetProperty("fill", " ").(string)
				tabCell.Size = token.GetProperty("size", 0).(int)
				tabCell.Origin = token.GetProperty("origin", 0).(int)
				tabCell.drawMode = token.GetProperty("draw", "relative").(string)
			} else if strings.HasPrefix(token.Text, "</tab>") {
				t.rows = append(t.rows, *tabCell)
				tabCell = NewTabCell(tabRow)

			}
		} else {
			tabCell.Text = token.Text
			tabRow.Cells = append(tabRow.Cells, *tabCell)
			tabCell = NewTabCell(tabRow)
		}
		tabRow.maxLengths = append(tabRow.maxLengths, tSize) // get the maximum length of the row
	}
	return tabRow
}

func (t *TabOut) IsTable(text string) bool {
	return strings.HasPrefix(text, TableStart) || strings.HasSuffix(text, TableEnd)
}

func (t *TabOut) IsRow(text string) bool {
	return strings.HasPrefix(text, RowStart) && strings.HasSuffix(text, RowEnd)
}

func (t *TabOut) IsTab(text string) bool {
	return strings.HasPrefix(text, TabStart) && strings.HasSuffix(text, TabEnd)
}

func (t *TabOut) GetSize(orig int) int {
	if t.info.IsTerminal {
		if t.RowCalcMode == 0 { // relative to terminal width
			if orig > 100 {
				orig = 100
			}
			return (t.info.Width * orig) / 100
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
			if strings.HasPrefix(text, "<table") {
				return "" // we are already in table mode, so we return nothing
			}
			if strings.HasPrefix(text, "</table>") {
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

	} else if t.IsRow(text) {
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
// and we do not wailt for a table end
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

// PadString Returns max len string filled with spaces
func PadString(line string, max int, fillChar string) string {
	if LenPrintable(line) > max {
		lastEsc := GetLastEscapeSequence(line)
		runes := []rune(line)
		safeSubstring := string(runes[0:max]) + lastEsc
		return safeSubstring
	}
	diff := max - LenPrintable(line)
	for i := 0; i < diff; i++ {
		line = line + fillChar
	}
	return line
}

// PadStringToRight Returns max len string filled with spaces right placed
func PadStringToRight(line string, max int, fillChar string) string {
	if LenPrintable(line) > max {
		lastEsc := GetLastEscapeSequence(line[:max])
		runes := []rune(line)

		safeSubstring := string(runes[0:max]) + lastEsc
		return safeSubstring
	}
	diff := max - LenPrintable(line)
	for i := 0; i < diff; i++ {
		line = fillChar + line
	}
	return line
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
