package ctxout

import (
	"fmt"
	"strings"
)

// just shortcuts for table tags
const (
	OPEN_TABLE          = "<table>"
	OT                  = "<table>"
	CLOSE_TABLE         = "</table>"
	CT                  = "</table>"
	OPEN_ROW            = "<row>"
	OR                  = "<row>"
	CLOSE_ROW           = "</row>"
	CR                  = "</row>"
	OPEN_TAB            = "<tab>"
	OTB                 = "<tab>"
	CLOSE_TAB           = "</tab>"
	CTB                 = "</tab>"
	CLOSE_TAB_ROW       = "</tab></row>"
	CTR                 = "</tab></row>"
	CLOSE_TAB_ROW_TABLE = "</tab></row></table>"
	CTRT                = "</tab></row></table>"
	OPEN_TABLE_ROW      = "<table><row>"
	OTR                 = "<table><row>"
	CLOSE_ROW_TABLE     = "</row></table>"
	CRT                 = "</row></table>"
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
	return OPEN_ROW + strings.Join(cells, "") + CLOSE_ROW
}

func Table(rows ...string) string {
	return OPEN_TABLE + strings.Join(rows, "") + CLOSE_TABLE
}
