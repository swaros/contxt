package cmdhandle

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/manout"
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
	mainScr       *tview.TextView
	statusScr     *tview.TextView
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

	status := tview.NewTextView()
	status.SetText("[blue]version [yellow]" + configure.GetVersion() + "\n [blue]build[yellow] " + configure.GetBuild())
	status.SetBorder(true)
	status.SetDynamicColors(true)
	ui.statusScr = status

	mainWindow := tview.NewTextView()
	mainWindow.SetBorder(true).
		SetBackgroundColor(tcell.ColorDarkGoldenrod)
	mainWindow.SetDynamicColors(true)

	ui.mainScr = mainWindow

	flex := tview.NewFlex().
		AddItem(menu, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(status, 0, 1, false).
			AddItem(mainWindow, 0, 5, false), 0, 3, false)

	pages.AddPage("main", flex, true, true)
	app.SetRoot(pages, true).EnableMouse(false)

	ui.startCapture()
	if err := app.Run(); err != nil {
		return ui, err
	}

	return ui, nil
}

func (ui *CtxUi) createMenu() *tview.List {

	menu := tview.NewList().AddItem("tasks", "start Task", 'r', func() {
		ui.pages.SendToFront("target")
	}).AddItem("quit", "exit interactive mode", 'q', func() {
		ui.app.Stop()
	})
	ui.pages.AddPage("target", ui.CreateRunPage(), true, true)
	ui.menu = menu
	return menu
}

// FilterOutPut parses the content and handles
// all interface depending the Type differently
func (ui *CtxUi) FilterOutPut(caseHandle func(target string, msg []interface{}), msg ...interface{}) []interface{} {
	var newMsh []interface{} // new hash for the output
	haveTarget := ""
	for _, chk := range msg {
		switch chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return newMsh
			}
			continue
		case CtxOutLabel:
			ctrl := chk.(CtxOutLabel)
			newMsh = append(newMsh, manout.Message(ctrl.Message))
			continue
		case CtxTargetOut:
			ctrl := chk.(CtxTargetOut)
			haveTarget = ctrl.Target
			newMsh = append(newMsh, ctrl.Target)
		default:
			newMsh = append(newMsh, chk)
		}

	}
	if haveTarget != "" {
		caseHandle(haveTarget, newMsh)
		var dwMsh []interface{}
		return dwMsh
	}
	return newMsh
}

// startCapture set up the output capturing.
// it is also the method that is the "tick"
// because it will be triggered on statusmessage
// So this is the place to update all components
func (ui *CtxUi) startCapture() {
	// we set the PreHook so any Message that is send to
	// CtxOut will be handled from now on by this function
	PreHook = func(msg ...interface{}) bool {
		msg = ui.FilterOutPut(func(target string, msg []interface{}) {
			if ui.outscr != nil {
				byte4main := []byte(tview.TranslateANSI(fmt.Sprintln(msg...)))
				ui.outscr.Write(byte4main)
			}
		}, msg...) // filter output depending types of the content

		if len(msg) > 0 && ui.mainScr != nil {
			byteData := []byte(tview.TranslateANSI(fmt.Sprintln(msg...)))
			ui.mainScr.Write(byteData)

			if ui.outscr != nil {
				byte4main := []byte(tview.TranslateANSI(fmt.Sprintln(msg...)))
				ui.outscr.Write(byte4main)
			}
		}

		return true
	}
	CtxOut("running target")
}

func (ui *CtxUi) CreateRunPage() *tview.Flex {
	// this list contains ans target and we use them as a menu
	list := tview.NewList()
	if targets, ok := getAllTargets(); ok {
		for index, target := range targets {
			list.AddItem(target, "run target "+target, rune(strconv.Itoa(index)[0]), nil)
		}
	}
	list.SetHighlightFullLine(true)
	list.SetSelectedFunc(func(i int, target, s2 string, r rune) {
		if r != 'q' {
			if ui.outscr != nil {
				ui.outscr.Clear()
			}
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
	ui.outscr = output
	targetflex := tview.NewFlex().
		AddItem(list, 0, 1, true).AddItem(output, 0, 4, false)
	return targetflex

}
