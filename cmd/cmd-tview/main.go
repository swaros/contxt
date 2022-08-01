package main

import (
	"fmt"

	"github.com/swaros/contxt/tviewapp"
)

func main() {
	fmt.Println("tview example")

	app := tviewapp.NewApplication(true)
	app.SetHeader("tview example")
	app.NewPage("start", tviewapp.ButtonMenuPageStyle)

	/*
		demoPage := app.NewPageWithFlex("demo")
		demoPage.AddItem(tview.NewButton("exit").SetSelectedFunc(func() {
			app.Stop()
		}), 0, 1, true).AddItem(tview.NewButton("set header").SetSelectedFunc(func() {
			app.SetHeader("clicked")
		}), 0, 1, false)
	*/
	app.Start()

}
