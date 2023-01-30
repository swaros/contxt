package ctxout

import (
	"strings"

	"github.com/swaros/contxt/module/systools"
)

const (
	RowStart = "<row>"
	RowEnd   = "</row>"
	TabStart = "<tab"
	TabEnd   = "</tab>"
)

type table struct {
	rows []tabRow
}

// tabRow is a single row in a table
type tabRow struct {
	Cells []tabCell
}

// tabCell is a single cell in a row
type tabCell struct {
	Size         int
	Origin       int
	OriginString string
	Text         string
}

type TabOut struct {
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

func NewTabRow() *tabCell {
	return &tabCell{
		Size:   0, // 0 = auto
		Origin: 0, // 0 left, 1 center, 2 right
		Text:   "",
	}
}

func (t *TabOut) Filter(msg interface{}) interface{} {
	return msg
}

func (t *TabOut) Update(info PostFilterInfo) {
	t.info = info
}

func (t *TabOut) CanHandleThis(text string) bool {
	return strings.HasPrefix(text, RowStart) && strings.HasSuffix(text, RowEnd)
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

func (t *TabOut) Render() string {
	var result string
	for _, row := range t.rows {
		if row.Size > 0 {
			size := t.GetSize(row.Size)
			switch row.Origin {
			case 0: // left padding
				row.Text = PadString(row.Text, size)
			case 1:
				row.Text = systools.StringSubRight(row.Text, size)
				row.Text = PadString(row.Text, size)
			case 2:
				row.Text = PadStringToR(row.Text, size)
			}

			switch row.OriginString {
			case "left":
				row.Text = PadString(row.Text, size)
			case "rightReverse":
				row.Text = PadString(row.Text, size)
			case "right":
				row.Text = PadStringToR(row.Text, size)
			}
		}
		result += row.Text
	}
	t.rows = []tabCell{}
	return result
}

func (t *TabOut) Command(cmd string) string {
	if t.CanHandleThis(cmd) {
		cmd = strings.TrimPrefix(cmd, RowStart)
		cmd = strings.TrimSuffix(cmd, RowEnd)
		tokens := t.markup.Parse(cmd)
		tabRow := NewTabRow()
		for _, token := range tokens {
			if token.IsMarkup {
				if strings.HasPrefix(token.Text, "<tab") {
					tabRow.Size = 0
					tabRow.Origin = 0
					if strings.Contains(token.Text, "size=") {
						tabRow.Size = t.markup.GetMarkupIntValue(token.Text, "size")
					}
					if strings.Contains(token.Text, "origin=") {
						tabRow.Origin = t.markup.GetMarkupIntValue(token.Text, "origin")
					}
				} else if strings.HasPrefix(token.Text, "</tab>") {
					t.rows = append(t.rows, *tabRow)
					tabRow = NewTabRow()

				}
			} else {
				tabRow.Text = token.Text
				t.rows = append(t.rows, *tabRow)
				tabRow = NewTabRow()
			}
		}
		return t.Render()
	}

	return cmd
}

// PadString Returns max len string filled with spaces
func PadString(line string, max int) string {
	if len(line) > max {
		lastEsc := GetLastEscapeSequence(line)
		runes := []rune(line)
		safeSubstring := string(runes[0:max]) + lastEsc
		return safeSubstring
	}
	diff := max - len(line)
	for i := 0; i < diff; i++ {
		line = line + " "
	}
	return line
}

// PadStringToR Returns max len string filled with spaces right placed
func PadStringToR(line string, max int) string {
	if len(line) > max {
		lastEsc := GetLastEscapeSequence(line[:max])
		runes := []rune(line)

		safeSubstring := string(runes[0:max]) + lastEsc
		return safeSubstring
	}
	diff := max - len(line)
	for i := 0; i < diff; i++ {
		line = " " + line
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
