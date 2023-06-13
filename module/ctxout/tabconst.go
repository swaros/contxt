package ctxout

import (
	"fmt"
	"strings"
)

// just shortcuts for table tags
const (
	OpenTable        = "<table>"
	TO               = "<table>"
	CloseTable       = "</table>"
	TC               = "</table>"
	OpenRow          = "<row>"
	RO               = "<row>"
	CloseRow         = "</row>"
	CR               = "</row>"
	OpenTab          = "<tab>"
	OTB              = "<tab>"
	CloseTab         = "</tab>"
	CTB              = "</tab>"
	CloseTabRow      = "</tab></row>"
	CTR              = "</tab></row>"
	CloseTabRowTable = "</tab></row></table>"
	CTRT             = "</tab></row></table>"
	OpenTableRow     = "<table><row>"
	OTR              = "<table><row>"
	CloseRowTable    = "</row></table>"
	CRT              = "</row></table>"
)

// Table provides a way to create a table with size <table size='X'>
func Tab(size int) string {
	return "<tab size='" + fmt.Sprintf("%v", size) + "'>"
}

// TabF provides a way to create a tab with properties <tab prop1='val1' prop2='val2'>
func TabF(props ...string) string {
	pre := "<tab"
	for _, prop := range props {
		prps := strings.Split(prop, "=")
		if len(prps) == 2 {
			if strings.HasPrefix(prps[1], "'") && strings.HasSuffix(prps[1], "'") {
				pre += " " + prps[0] + "=" + prps[1]
			} else {
				pre += " " + prps[0] + "='" + prps[1] + "'"
			}
		}
	}
	pre += ">"
	return pre
}

// TD provides a way to create a table cell <tab size='X'>content</tab>
func TD(content interface{}, props ...string) string {
	return TabF(props...) + fmt.Sprintf("%v", content) + "</tab>"
}

// Prop provides a way to create a property for a tab <tab prop='val'>
func Prop(name string, value interface{}) string {
	return fmt.Sprintf("%s='%v'", name, value)
}

func Row(cells ...string) string {
	return OpenRow + strings.Join(cells, "") + CloseRow
}

func Table(rows ...string) string {
	return OpenTable + strings.Join(rows, "") + CloseTable
}

// little helpers to set properties
func Right() string {
	return "origin='2'"
}

func Left() string {
	return "origin='1'"
}

func Fill(with string) string {
	return "fill='" + with + "'"
}

func Size(size int) string {
	return "size='" + fmt.Sprintf("%v", size) + "'"
}

func Margin(margin int) string {
	return "margin='" + fmt.Sprintf("%v", margin) + "'"
}

func Origin(origin int) string {
	return "origin='" + fmt.Sprintf("%v", origin) + "'"
}

func Fixed() string {
	return "draw='fixed'"
}

func Relative() string {
	return "draw='relative'"
}

func Content() string {
	return "draw='content'"
}

func Extend() string {
	return "draw='extend'"
}

func CutNotifier(notifier string) string {
	return "cut='" + notifier + "'"
}

func Overflow(mode string) string {
	return "overflow='" + mode + "'"
}

func OverflowIgnore() string {
	return "overflow='ignore'"
}

func OverflowWrap() string {
	return "overflow='wrap'"
}

func OverflowContent(content string) string {
	return "overflow-content='" + content + "'"
}
