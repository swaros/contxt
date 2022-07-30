package cmdhandle

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/systools"
	"github.com/swaros/manout"
)

type CtxUi struct {
	title          string
	app            *tview.Application
	pages          *tview.Pages
	menu           *tview.List
	cmd            *cobra.Command
	outscr         *tview.TextView
	args           []string
	LogOutMessage  string
	mainScr        *tview.TextView
	statusScr      *tview.TextView
	taskScr        *tview.TextView
	selectedtarget string
	targetCtrl     *tview.Form
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

	status := tview.NewTextView()
	status.SetText("[blue]version [yellow]" + configure.GetVersion() + "\n [blue]build[yellow] " + configure.GetBuild())
	status.SetBorder(true)
	status.SetDynamicColors(true)
	ui.statusScr = status

	mainWindow := tview.NewTextView()
	mainWindow.SetBorder(true)
	mainWindow.SetDynamicColors(true)

	ui.mainScr = mainWindow

	flex := tview.NewFlex().
		AddItem(menu, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(status, 0, 1, false).
			AddItem(mainWindow, 0, 5, false), 0, 3, false)

	pages.AddPage("main", flex, true, true)
	app.SetRoot(pages, true).EnableMouse(true)

	ui.startCapture()
	// register exist trigger to get the app closed before
	systools.AddExitListener("interactive", func(exitCode int) {
		app.Stop()
	})
	if err := app.Run(); err != nil {
		return ui, err
	}

	return ui, nil
}

// createMenu creates the main menu as a default tview.List
func (ui *CtxUi) createMenu() *tview.List {

	menu := tview.NewList().AddItem("Task", "tasks overview", 't', func() {
		ui.pages.SendToFront("target")
	}).AddItem("quit", "EXIT", 'q', func() {
		ui.app.Stop()
	})
	ui.pages.AddPage("target", ui.CreateRunPage(), true, true)
	menu.SetHighlightFullLine(true)
	ui.menu = menu
	return menu
}

// FilterOutPut parses the content and handles
// all interface depending the Type differently
func (ui *CtxUi) FilterOutPut(caseHandle func(target string, msg []interface{}), msg ...interface{}) []interface{} {
	var newMsh []interface{} // new hash for the output
	haveTarget := ""
	for _, chk := range msg {
		switch v := chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return newMsh
			}
			continue
		case CtxOutLabel:
			newMsh = append(newMsh, manout.Message(v.Message))
			continue
		case CtxTargetOut:
			haveTarget = v.Target
			newMsh = append(newMsh, v.Target)
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

func (ui *CtxUi) updateTaskView() {
	if ui.selectedtarget != "" && ui.taskScr != nil {
		ui.taskScr.SetText(ui.selectedtarget)

	}

	if ui.targetCtrl != nil {
		ui.targetCtrl.Clear(true)
		if ui.selectedtarget != "" {
			ui.targetCtrl.AddButton("Start "+ui.selectedtarget, func() {
				go RunTargets(ui.selectedtarget, true)
			})
		}
	}
}

// CreateRunPage builds the page that contains different elements
// to inspect and run the targets
func (ui *CtxUi) CreateRunPage() *tview.Flex {
	// this uiTaskList contains any target and we use them as a menu
	uiTaskList := tview.NewList()
	var keyList string = "abcdefghijklmnopqrstuvwyz1234567890" // shortcuts definition
	if targets, ok := getAllTargets(); ok {                    // get all targets
		for index, target := range targets {
			if index <= len(keyList) { // we just print targets until we have chars to map
				uiTaskList.AddItem(target, "", rune(keyList[index]), nil) // add the target as listitem
			}
		}
	}
	uiTaskList.SetHighlightFullLine(true)
	uiTaskList.SetSelectedFunc(func(i int, target, s2 string, r rune) {
		if r != 'x' { // we ignore the get-back button
			if ui.outscr != nil {
				ui.outscr.Clear()
			}
			ui.selectedtarget = target
			go RunTargets(target, true)
		} else {
			ui.selectedtarget = ""
		}
	})

	uiTaskList.AddItem("close", "", 'x', func() {
		ui.pages.SendToBack("target")
	})
	uiTaskList.SetBorder(true)
	uiTaskList.SetTitle("select target")
	uiTaskList.ShowSecondaryText(false)
	uiTaskList.SetChangedFunc(func(index int, target, emptyAnyway string, shortcut rune) {
		if shortcut != 'x' { // ignore get back option
			ui.selectedtarget = target
			ui.updateTaskView()
		}
	})

	// create the log output
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	output.SetBorder(true).SetTitle("log")
	ui.outscr = output

	// create a target overview
	targetControl := tview.NewForm()
	targetControl.SetBorder(true)
	ui.targetCtrl = targetControl

	// left side we have the list of task
	// and the form that we use to start a task
	leftCtrl := tview.NewFlex().SetDirection(tview.FlexRow)
	leftCtrl.AddItem(uiTaskList, 0, 6, true).
		AddItem(targetControl, 0, 1, false)
	// this is the task overview where
	// we display the current status of the task
	targetView := tview.NewTextView()
	targetView.SetDynamicColors(true).SetBorder(true)
	ui.taskScr = targetView

	// the right site of the page contains
	// the target overview and the log output
	rightCtrl := tview.NewFlex().SetDirection(tview.FlexRow)
	rightCtrl.AddItem(targetView, 0, 1, false).
		AddItem(output, 0, 1, false)

	// compose the page content
	targetflex := tview.NewFlex().
		AddItem(leftCtrl, 0, 1, true).
		AddItem(rightCtrl, 0, 4, false)
	return targetflex

}
