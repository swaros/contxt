package ctxout

import (
	"errors"
	"strings"
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
	rows             []*tabRow
	parent           *TabOut
	rowSeperator     string
	sizeCalculations map[int]int
}

// tabRow is a single row in a table
type tabRow struct {
	Cells        []*tabCell
	parent       *tableHandle
	rowEndString string
	maxLengths   []int
	Err          error
}

// tabCell is a single cell in a row
type tabCell struct {
	Size            int
	Origin          int
	OriginString    string
	Text            string
	parent          *tabRow
	fillChar        string
	index           int    // reference to the index in the parent row
	drawMode        string // fixed = fixed size, relative = relative to terminal size, content = max size of content
	cutNotifier     string // if the text is cutted, then this string will be added to the end of the text
	overflow        bool   // if the text is cutted, then this will be set to true
	overflowContent string // if the text is cutted, then this will be set to the cutted content
	overflowMode    string // this is the mode how the overflow is handled. ignore = the text is ignored, wrap = wrap the text
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
		Size:            0, // 0 = auto
		Origin:          0, // 0 left, 1 center, 2 right
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
	return td
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
	return td.Text
}

func (td *tabCell) GetSize() int {
	return td.Size
}

func (td *tabCell) GetOrigin() int {
	return td.Origin
}

func NewTabRow(parent *tableHandle) *tabRow {
	return &tabRow{
		Cells:        []*tabCell{},
		parent:       parent,
		rowEndString: "",
	}
}

func (tr *tabRow) SetRowEndString(rowEndString string) *tabRow {
	tr.rowEndString = rowEndString
	return tr
}

func (tr *tabRow) GetRowEndString() string {
	return tr.rowEndString
}

func (tr *tabRow) GetCells() []*tabCell {
	return tr.Cells
}

func (tr *tabRow) GetCell(index int) *tabCell {
	if len(tr.Cells) > index {
		return tr.Cells[index]
	}
	return nil
}

func (tr *tabRow) CreateRow() *tabRow {
	return NewTabRow(tr.parent)
}

func (tr *tabRow) CreateCell() *tabCell {
	return NewTabCell(tr)
}

func (tr *tabRow) AddCell(cell *tabCell) *tabRow {
	cell.parent = tr
	tr.Cells = append(tr.Cells, cell)
	return tr
}

func (tr *tabRow) AddCells(cells []*tabCell) *tabRow {
	for _, cell := range cells {
		tr.AddCell(cell)
	}
	return tr
}

func (tr *tabRow) GetMaxLengths() []int {
	return tr.maxLengths
}

func NewTableHandle(parent *TabOut) *tableHandle {
	return &tableHandle{
		rows:             []*tabRow{},
		parent:           parent,
		rowSeperator:     "\n",
		sizeCalculations: make(map[int]int),
	}
}

func (th *tableHandle) SetRowSeperator(rowSeperator string) *tableHandle {
	th.rowSeperator = rowSeperator
	return th
}

func (th *tableHandle) CreateRow() *tabRow {
	return NewTabRow(th)
}

func (th *tableHandle) GetRowSeperator() string {
	return th.rowSeperator
}

func (th *tableHandle) AddRow(row *tabRow) *tableHandle {
	row.parent = th
	th.rows = append(th.rows, row)
	return th
}

func (th *tableHandle) AddRows(rows []*tabRow) *tableHandle {
	for _, row := range rows {
		th.AddRow(row)
	}
	return th
}

func (th *tableHandle) GetRows() []*tabRow {
	return th.rows
}

func (th *tableHandle) GetRow(index int) *tabRow {
	if len(th.rows) > index {
		return th.rows[index]
	}
	return nil
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

func (tr *tabRow) Render() (string, *tabRow, error) {
	tr.Err = nil // reset error
	if len(tr.Cells) == 0 {
		tr.Err = errors.New("no cells to render")
		return "", nil, tr.Err
	}

	if tr.parent == nil {
		tr.Err = errors.New("no parent table")
		return "", nil, tr.Err
	}

	if tr.parent.parent == nil {
		tr.Err = errors.New("no parent table parent")
		return "", nil, tr.Err
	}

	var result []string
	// this is the row that will be used if we have a wrap overflow mode.
	// this will be created and updated all the time, but used only if we found an overflow usagewith wrap mode
	wrapRow := NewTabRow(tr.parent)

	for indx, cell := range tr.Cells {
		wrapRow.AddCell(cell)
		if cell.Size > 0 {
			rowSize := tr.GetSize(cell, indx)
			size := tr.parent.parent.GetSize(cell.Size)
			if rowSize > 0 {
				size = rowSize
			}
			switch cell.Origin {
			case 0: // left padding
				result = append(result, cell.PadString(size))
			case 1:
				result = append(result, cell.PadStringToRightStayLeft(size))
			case 2:
				result = append(result, cell.PadStringToRight(size))
			}

		} else {
			result = append(result, cell.Text)
		}
		// we add all the cells to the wrap row, so we can use it if we have a wrap overflow mode
		// any cell have the information about the overflow mode and it have also
		// the content of the overflow text

	}

	// now we try to wrap the cells if we have a wrap overflow mode
	// if this returns true, then we have a wrap overflow mode
	// and also print a additional row with the overflow text
	// this must be recursive, because we can have a wrap overflow mode in a wrap overflow mode
	if wrapRow.WrapCells() {
		return strings.Join(result, tr.rowEndString), wrapRow, nil
	}
	//return []string{strings.Join(result, tr.rowEndString)}
	return strings.Join(result, tr.rowEndString), nil, nil
}

func (tb *tableHandle) Render() string {
	var result []string
	for _, row := range tb.rows {
		cnt, expandRow, _ := row.Render()
		if cnt != "" {
			result = append(result, cnt)
		}
		if expandRow != nil {
			for {
				cnt, expandRow, _ := expandRow.Render()
				if cnt != "" {
					result = append(result, cnt)
				}
				if expandRow == nil {
					break
				}
			}
		}
	}
	firstRow := strings.Join(result, tb.rowSeperator)
	return firstRow
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
				tabCell.fillChar = token.GetProperty("fill", " ").(string)
				tabCell.Size = token.GetProperty("size", 0).(int)
				tabCell.Origin = token.GetProperty("origin", 0).(int)
				tabCell.drawMode = token.GetProperty("draw", "relative").(string)
				tabCell.cutNotifier = token.GetProperty("cut-add", "...").(string)
				tabCell.overflowMode = token.GetProperty("overflow", "ignore").(string)
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
	cell.SetText(line).SetFillChar(fillChar)
	return cell.PadString(max)
}

// PadStrRight is a shortcut for PadString on a cell
func PadStrRight(line string, max int, fillChar string) string {
	cell := NewTabCell(nil)
	cell.SetText(line).SetFillChar(fillChar)
	return cell.PadStringToRight(max)
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
	return newCell
}

// WrapCells wraps the cells in the row
// returns true if the row has changed
func (tr *tabRow) WrapCells() bool {
	changed := false
	for i, cell := range tr.Cells {
		if cell.MoveToWrap() {
			changed = true
		}
		tr.Cells[i] = cell
	}
	return changed
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

// PadString Returns max len string filled with spaces
func (td *tabCell) PadString(max int) string {
	if max < 1 {
		return ""
	}
	tSize := LenPrintable(td.Text)
	if tSize == max {
		return td.Text
	}
	if tSize > max {
		runes := []rune(td.Text)
		add := td.cutNotifier

		td.overflow = true
		if td.overflowMode != "ignore" {
			add = "" // if we keep the overflow, we do not add the cut notifier
			td.overflowContent = string(runes[max:])
		} else {
			max -= LenPrintable(td.cutNotifier)
			if max < 1 {
				max = 0
			}
		}
		td.Text = string(runes[0:max]) + add //+ lastEsc
		return td.Text
	}
	diff := max - tSize
	for i := 0; i < diff; i++ {
		td.Text += td.fillChar
	}
	return td.Text
}

// PadStringToRight Returns max len string filled with spaces right placed
func (td *tabCell) PadStringToRight(max int) string {
	if max < 1 {
		return ""
	}
	tSize := LenPrintable(td.Text)
	if tSize == max {
		return td.Text
	}
	if tSize > max {
		td.overflow = true
		runes := []rune(td.Text)
		add := td.cutNotifier
		if td.overflowMode != "ignore" {
			add = "" // if we keep the overflow, we do not add the cut notifier
			td.overflowContent = string(runes[max:])
		} else {
			max -= LenPrintable(td.cutNotifier)
			if max < 0 {
				max = 0
			}
		}
		safeSubstring := string(runes[0:max]) + add //+ lastEsc
		return safeSubstring
	}
	diff := max - tSize
	for i := 0; i < diff; i++ {
		td.Text = td.fillChar + td.Text
	}
	return td.Text
}

// PadStringToRight Returns max len string filled with spaces right placed
func (td *tabCell) PadStringToRightStayLeft(max int) string {
	if max < 1 {
		return ""
	}
	tSize := LenPrintable(td.Text)
	if tSize == max {
		return td.Text
	}
	if tSize > max {
		runes := []rune(td.Text)
		left := LenPrintable(td.Text) - max
		add := td.cutNotifier
		td.overflow = true
		if td.overflowMode != "ignore" {
			add = "" // if we keep the overflow, we do not add the cut notifier
			td.overflowContent = string(runes[left:])
		} else {
			max -= LenPrintable(td.cutNotifier)
			if max < 0 {
				max = 0
			}
		}
		safeSubstring := add + string(runes[left:])
		return safeSubstring
	}
	diff := max - tSize
	for i := 0; i < diff; i++ {
		td.Text += td.fillChar
	}
	return td.Text
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
