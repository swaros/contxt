package cmdhandle

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

type CtxUi struct {
	title         string
	app           *tview.Application
	pages         *tview.Pages
	menu          *tview.List
	cmd           *cobra.Command
	outscr        *tview.TextView
	args          []string
	LogOutMessage string
}

func InitWindow(cmd *cobra.Command, args []string) (*CtxUi, error) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	ui := &CtxUi{
		title: "con.txt",
		app:   app,
		pages: pages,
		cmd:   cmd,
		args:  args,
	}
	menu := ui.createMenu()
	menu.SetBorder(true)
	menu.SetBackgroundColor(tcell.ColorDarkBlue)
	menu.SetMainTextColor(tcell.ColorLightBlue)
	flex := tview.NewFlex().
		AddItem(menu, 0, 1, true)

	pages.AddPage("main", flex, true, true)
	app.SetRoot(pages, true)

	if err := app.Run(); err != nil {
		return ui, err
	}
	return ui, nil
}

func (ui *CtxUi) createMenu() *tview.List {

	menu := tview.NewList().AddItem("tasks", "run tasks", 'r', func() {
		ui.pages.SendToFront("target")
	}).AddItem("quit", "exit interactive mode", 'q', func() {
		ui.app.Stop()
	})
	ui.pages.AddPage("target", ui.CreateTargetList(), true, true)
	ui.menu = menu
	return menu
}

func (ui *CtxUi) updateTime() {
	for {
		time.Sleep(time.Millisecond * 500)
		ui.app.QueueUpdateDraw(func() {
			ui.outscr.SetText(ui.LogOutMessage)
		})
	}
}

func (ui *CtxUi) CreateTargetList() *tview.Flex {
	// this list contains ans target and we use them as a menu
	list := tview.NewList()
	if targets, ok := getAllTargets(); ok {
		for index, target := range targets {
			list.AddItem(target, "run target "+target, rune(strconv.Itoa(index)[0]), nil)
		}
	}

	list.SetSelectedFunc(func(i int, target, s2 string, r rune) {
		if r != 'q' {
			go RunTargets(target, true)
		}
	})

	list.AddItem("Quit", "Press to exit", 'q', func() {
		ui.pages.SendToBack("target")
	})
	list.SetBorder(true)
	list.SetTitle("select target")

	output := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	output.SetBorder(true).SetTitle("log")

	PreHook = func(msg ...interface{}) bool {
		byteData := []byte(tview.TranslateANSI(fmt.Sprintln(msg...)))
		output.Write(byteData)

		//output.SetText(output.GetText(false) + tview.TranslateANSI(fmt.Sprint(msg...)))
		//ui.app.QueueUpdateDraw(func() {})
		/*
			go func(app *tview.Application) {
				app.Draw()
			}(ui.app)
		*/
		return true
	}
	CtxOut("running target")

	targetflex := tview.NewFlex().
		AddItem(list, 0, 1, true).AddItem(output, 0, 1, false)
	return targetflex

}
