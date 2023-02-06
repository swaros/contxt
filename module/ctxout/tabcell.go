package ctxout

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
