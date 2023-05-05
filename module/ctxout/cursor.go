package ctxout

import (
	"strconv"
	"strings"

	"atomicgo.dev/cursor"
)

// filter that sets the cursor position

// CursorFilter is a filter that sets the cursor position
type CursorFilter struct {
	// X is the x position of the cursor
	X int
	// Y is the y position of the cursor
	Y int

	// Info is the PostFilterInfo
	Info PostFilterInfo
}

func NewCursorFilter() *CursorFilter {
	return &CursorFilter{}
}

// Filter is called when the context is updated
// interface fulfills the PostFilter interface
func (t *CursorFilter) Filter(msg interface{}) interface{} {
	t.command(msg.(string))
	return msg
}

// Update is called when the context is updated
// interface fulfills the PostFilter interface
func (t *CursorFilter) Update(info PostFilterInfo) {
	t.Info = info
}

func (t *CursorFilter) Command(str string) string {
	return t.command(str)
}

// CanHandleThis returns true if the text is requesting a cursor position
// interface fulfills the PostFilter interface
func (t *CursorFilter) CanHandleThis(text string) bool {
	return t.IsCursor(text)
}

// Command is called when the text is a cursor position
// interface fulfills the PostFilter interface
func (t *CursorFilter) IsCursor(text string) bool {
	return strings.HasPrefix(text, "cursor:")
}

// Command handler they get the paramaters
// from the text and maps it to the https://github.com/atomicgo/cursor package
func (t *CursorFilter) command(text string) string {
	// test starts allways wirth cursor:
	text = strings.TrimPrefix(text, "cursor:")
	// keep anything after the first ;
	textSplits := strings.Split(text, ";")
	textKeep := ""
	if len(textSplits) > 1 {
		textKeep = strings.Join(textSplits[1:], ";")
	}
	text = textSplits[0]
	// split the text by comma
	split := strings.Split(text, ",")
	// fists param is the command we use
	command := split[0]
	// the rest are the params
	params := split[1:]
	// switch on the command
	switch strings.ToLower(command) {
	case "up":
		cursor.Up(t.getArgAsInt(params[0]))
	case "down":
		cursor.Down(t.getArgAsInt(params[0]))
	case "left":
		cursor.Left(t.getArgAsInt(params[0]))
	case "right":
		cursor.Right(t.getArgAsInt(params[0]))
	case "move":
		cursor.Move(t.getArgAsInt(params[0]), t.getArgAsInt(params[1]))
	case "bottom":
		cursor.Bottom()
	default:
		return text

	}
	return textKeep
}

func (t *CursorFilter) getArgAsInt(arg string) int {
	// convert the string to int
	// if it fails return 0
	i, err := strconv.Atoi(arg)
	if err != nil {
		return 0
	}
	return i
}
