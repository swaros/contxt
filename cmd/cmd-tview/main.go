package main

import (
	"fmt"
	"os"

	"github.com/swaros/contxt/tviewapp"
)

func main() {
	fmt.Println("tview example")

	app := tviewapp.New()
	app.NewScreen()
	app.RunLoop(func() { os.Exit(0) })

	/*
		app := tviewapp.NewApplication(true)
		app.SetHeader("tview example")

		exitButton := tviewapp.TvButton{
			Label:   "exit app",
			OnClick: func() { app.Stop() },
		}

		if err := app.NewPage("start", tviewapp.ButtonMenuPageStyle, exitButton); err != nil {
			panic(err)
		}

		/*
			demoPage := app.NewPageWithFlex("demo")
			demoPage.AddItem(tview.NewButton("exit").SetSelectedFunc(func() {
				app.Stop()
			}), 0, 1, true).AddItem(tview.NewButton("set header").SetSelectedFunc(func() {
				app.SetHeader("clicked")
			}), 0, 1, false)
	*/
	//app.Start()

}
