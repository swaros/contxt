package ctxout

import (
	"errors"
	"strconv"
	"strings"

	"atomicgo.dev/cursor"
)

// filter that sets the cursor position
// the text is in the format cursor:command,param1,param2;rest of the text
// example: cursor:down,1;hello world
// the command is the cursor command
// the params are the params for the command
// the rest of the text is the text that is returned

// for the cursor movement https://github.com/atomicgo/cursor package is used
// except for area, which is not implemented yet, but will be in the future

// CursorFilter is a filter that sets the cursor position
type CursorFilter struct {
	// Info is the PostFilterInfo
	Info PostFilterInfo
	//Last error that occured
	LastError error
}

func NewCursorFilter() *CursorFilter {
	return &CursorFilter{}
}

// Filter is called when the context is updated
// interface fulfills the PostFilter interface
func (t *CursorFilter) Filter(msg interface{}) interface{} {
	// check if a string
	if _, ok := msg.(string); !ok {
		return t.cmd(msg.(string))
	}
	return msg
}

// Update is called when the context is updated
// interface fulfills the PostFilter interface
func (t *CursorFilter) Update(info PostFilterInfo) {
	t.Info = info
}

func (t *CursorFilter) Command(str string) string {
	return t.cmd(str)
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
func (t *CursorFilter) cmd(text string) string {
	// text starts allways with cursor:
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
	// and call the cursor package
	/*
		func Bottom()
		func ClearLine()
		func ClearLinesDown(n int)
		func ClearLinesUp(n int)
		func Down(n int)
		func DownAndClear(n int)
		func Hide()
		func HorizontalAbsolute(n int)
		func Left(n int)
		func Move(x, y int)
		func Right(n int)
		func SetTarget(w Writer)
		func Show()
		func StartOfLine()
		func StartOfLineDown(n int)
		func StartOfLineUp(n int)
		func TestCustomIOWriter(t *testing.T)
		func Up(n int)
		func UpAndClear(n int)
	*/

	switch strings.ToLower(command) {
	case "up":
		if t.assertParams(params, 1) {
			cursor.Up(t.getArgAsInt(params[0]))
		}
	case "down":
		if t.assertParams(params, 1) {
			cursor.Down(t.getArgAsInt(params[0]))
		}

	case "left":
		if t.assertParams(params, 1) {
			cursor.Left(t.getArgAsInt(params[0]))
		}

	case "right":
		if t.assertParams(params, 1) {
			cursor.Right(t.getArgAsInt(params[0]))
		}

	case "move":
		if t.assertParams(params, 2) {
			cursor.Move(t.getArgAsInt(params[0]), t.getArgAsInt(params[1]))
		}

	case "bottom":
		cursor.Bottom()

	case "clearline":
		cursor.ClearLine()

	case "clearlinesdown":
		if t.assertParams(params, 1) {
			cursor.ClearLinesDown(t.getArgAsInt(params[0]))
		}

	case "clearlinesup":
		if t.assertParams(params, 1) {
			cursor.ClearLinesUp(t.getArgAsInt(params[0]))
		}

	case "downandclear":
		if t.assertParams(params, 1) {
			cursor.DownAndClear(t.getArgAsInt(params[0]))
		}

	case "hide":
		cursor.Hide()

	case "horizontalabsolute":
		if t.assertParams(params, 1) {
			cursor.HorizontalAbsolute(t.getArgAsInt(params[0]))
		}

	case "startofline":
		cursor.StartOfLine()

	case "startoflinedown":
		if t.assertParams(params, 1) {
			cursor.StartOfLineDown(t.getArgAsInt(params[0]))
		}

	case "startoflineup":
		if t.assertParams(params, 1) {
			cursor.StartOfLineUp(t.getArgAsInt(params[0]))
		}

	case "show":
		cursor.Show()

	case "upandclear":
		if t.assertParams(params, 1) {
			cursor.UpAndClear(t.getArgAsInt(params[0]))
		}

	default:
		t.LastError = errors.New("invalid command " + command)

	}

	return textKeep
}

func (t *CursorFilter) assertParams(params []string, length int) bool {
	if len(params) != length {
		t.LastError = errors.New(
			"invalid number of params. expected " + strconv.Itoa(length) + " got " + strconv.Itoa(len(params)),
		)
		return false
	}
	return true
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
