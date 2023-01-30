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
	rows         []tabRow
	parent       *TabOut
	rowSeperator string
}

// tabRow is a single row in a table
type tabRow struct {
	Cells        []tabCell
	parent       *tableHandle
	rowEndString string
}

// tabCell is a single cell in a row
type tabCell struct {
	Size         int
	Origin       int
	OriginString string
	Text         string
	parent       *tabRow
	fillChar     string
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
		fillChar: "",
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
		rows:         []tabRow{},
		parent:       parent,
		rowSeperator: "\n",
	}
}

// Row functions

func (tr *tabRow) Render() string {
	if len(tr.Cells) == 0 {
		return ""
	}
	var result []string
	for _, cell := range tr.Cells {
		if cell.Size > 0 {
			size := tr.parent.parent.GetSize(cell.Size)
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

func (t *TabOut) ScanForRows(tokens []Parsed) *tableHandle {
	table := NewTableHandle(t)
	/*
		t.markup.BuildInnerSliceEach(tokens, "row", func(markup []Parsed) bool {
			row := t.ScanForCells(markup, table)
			table.rows = append(table.rows, *row)
			return true
		})*/
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

func (t *TabOut) GetProperty(text string, propertie string, defaultValue interface{}) interface{} {
	if strings.Contains(text, propertie) {
		switch defaultValue.(type) {
		case int:
			return t.markup.GetMarkupIntValue(text, propertie)
		case string:
			return t.markup.GetMarkupStringValue(text, propertie)
		default:
			return defaultValue
		}
	} else {
		return defaultValue
	}
}

func (t *TabOut) ScanForCells(tokens []Parsed, table *tableHandle) *tabRow {
	tabRow := NewTabRow(table)
	tabCell := NewTabCell(tabRow)
	for _, token := range tokens {
		if token.IsMarkup {
			if strings.HasPrefix(token.Text, "<tab") {
				tabCell.Size = 0
				tabCell.Origin = 0
				tabCell.fillChar = t.GetProperty(token.Text, "fill", " ").(string)
				if strings.Contains(token.Text, "size=") {
					tabCell.Size = t.markup.GetMarkupIntValue(token.Text, "size")
				}
				if strings.Contains(token.Text, "origin=") {
					tabCell.Origin = t.markup.GetMarkupIntValue(token.Text, "origin")
				}
			} else if strings.HasPrefix(token.Text, "</tab>") {
				t.rows = append(t.rows, *tabCell)
				tabCell = NewTabCell(tabRow)

			}
		} else {
			tabCell.Text = token.Text
			tabRow.Cells = append(tabRow.Cells, *tabCell)
			tabCell = NewTabCell(tabRow)
		}
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
		return "How came we here?"

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
