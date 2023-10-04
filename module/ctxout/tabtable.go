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
