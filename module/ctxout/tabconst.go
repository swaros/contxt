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
)

func Tab(size int) string {
	return "<tab size='" + fmt.Sprintf("%v", size) + "'>"
}

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
