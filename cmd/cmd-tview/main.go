package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/swaros/contxt/tviewapp"
)

func main() {
	fmt.Println("tview example")

	app := tviewapp.New()
	app.NewScreen()

	demo := tviewapp.NewText("hello world")

	demo2 := tviewapp.NewBox()
	demo2.SetDim(2, 10, 40, 3)
	demo2.OnMouseOver = func(x, y int) {
		demo.SetText(".... HOVER " + strconv.Itoa(x) + " x " + strconv.Itoa(y) + " .......")

	}

	demo2.OnMouseLeave = func() {
		demo.SetText("the blue box is untouched right now.....")
	}

	app.AddElement(demo, demo2)
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
