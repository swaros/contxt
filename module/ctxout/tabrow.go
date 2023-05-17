package ctxout

import (
	"errors"
	"strings"
)

// tabRow is a single row in a table
type tabRow struct {
	Cells        []*tabCell
	parent       *tableHandle
	rowEndString string
	maxLengths   []int
	Err          error
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

func (tr *tabRow) getLenByIndex(index int) (int, bool) {
	if len(tr.maxLengths) > index {
		return tr.maxLengths[index], true
	}
	return 0, false
}

func (tr *tabRow) GetSize(cell *tabCell, index int) int {
	if tr.parent.parent.GetInfo().IsTerminal {
		if tr.parent.parent.RowCalcMode == 0 { // relative to terminal width
			calculatedSize := tr.parent.parent.GetSize(cell.Size)

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
	// this will be created and updated all the time, but used only if we found an overflow usage with wrap mode
	wrapRow := NewTabRow(tr.parent) // so we just create them just in case we need them

	tr.parent.parent.GetRoundTool().Next()

	for indx, cell := range tr.Cells {
		wrapRow.AddCell(cell) // update the possible wrap row
		if cell.Size > 0 {
			size := tr.GetSize(cell, indx)
			// we just ignore any cell with a size of 0
			if size > 0 {
				result = append(result, cell.anyPrefix+cell.CutString(size)+cell.anySuffix)
			}

		} else {
			result = append(result, cell.GetText())
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
