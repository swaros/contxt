package tviewapp

import "github.com/rivo/tview"

var (
	ButtonMenuPageStyle tvPageStyle = tvPageStyle{template: TplPageButtonMenu}
)

type tvPageStyle struct {
	template int
	values   []interface{}
}

var templates map[int]func() tview.Primitive = make(map[int]func() tview.Primitive)

func CreatePageByStyle(style tvPageStyle) tview.Primitive {
	if hndl, ok := templates[style.template]; ok {
		return hndl()
	}
	return nil
}

func RegisterTemplate(styleId int, clbck func() tview.Primitive) {
	templates[styleId] = clbck
}

func initDefaults() {
	RegisterTemplate(TplPageButtonMenu, CreateButtonBarPage)
}

func CreateButtonBarPage() tview.Primitive {
	createFlex := tview.NewFlex()
	return createFlex
}
