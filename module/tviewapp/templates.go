package tviewapp

import (
	"fmt"

	"github.com/rivo/tview"
)

var (
	ButtonMenuPageStyle tvPageStyle = tvPageStyle{template: TplPageButtonMenu}
)

type tvPageStyle struct {
	template int
	values   []interface{}
}

var templates map[int]func(tvPageStyle) tview.Primitive = make(map[int]func(tvPageStyle) tview.Primitive)

func CreatePageByStyle(style tvPageStyle) tview.Primitive {
	if hndl, ok := templates[style.template]; ok {
		return hndl(style)
	}
	return nil
}

func RegisterTemplate(styleId int, clbck func(tvPageStyle) tview.Primitive) {
	templates[styleId] = clbck
}

func initDefaults() {
	RegisterTemplate(TplPageButtonMenu, CreateButtonBarPage)
}

func CreateButtonBarPage(style tvPageStyle) tview.Primitive {
	createFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	topBox := tview.NewForm()
	topBox.SetBorder(true)
	mainBox := tview.NewBox().SetBorder(true)
	mainBox.SetBorder(true)
	createFlex.AddItem(topBox, 0, 1, true)
	createFlex.AddItem(mainBox, 0, 9, false)
	for _, el := range style.values {
		switch t := el.(type) {
		case *TvButton:
			btn := tview.NewButton(t.Label).
				SetSelectedFunc(t.OnClick).
				SetFocusFunc(t.OnFocus)
			createFlex.AddItem(btn, 0, 1, false)

		case TvButton:
			/*btn := tview.NewButton(t.Label).
			SetSelectedFunc(t.OnClick).
			SetFocusFunc(t.OnFocus)*/

			topBox.AddButton(t.Label, t.OnClick)

		default:
			fmt.Println("invalid element submitted. only *TvButton supported.")
			panic("wrong element")
		}
	}
	return createFlex
}
