package ctxout

import "strings"

type tableHandle struct {
	rows             []*tabRow
	parent           *TabOut
	rowSeperator     string
	sizeCalculations map[int]int
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

// remember the calculated size of a column
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
