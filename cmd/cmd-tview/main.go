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

	card := tviewapp.NewCard("debug")

	demo := tviewapp.NewText("hello world")

	demo2 := tviewapp.NewBox()
	demo2.SetDim(2, 10, 40, 3)
	demo2.OnMouseOver = func(x, y int) {
		demo.SetText(".... HOVER " + strconv.Itoa(x) + " x " + strconv.Itoa(y) + " .......")

	}

	newBox := tviewapp.NewCard("hello world")

	app.Listener.OnLMouseDown = func(ca *tviewapp.CellApp, x, y int) {
		demo.SetText(".... left MOUSE DOWN" + strconv.Itoa(x) + " x " + strconv.Itoa(y) + " .......")
	}

	app.Listener.OnLMouseUp = func(ca *tviewapp.CellApp, x, y, startx, starty int) {

		newBox.SetDim(startx, starty, x-startx, y-starty)
		ca.AddElement(newBox)
		demo.SetText(".... left MOUSE LEAVE " + strconv.Itoa(x) + " x " + strconv.Itoa(y) + " .... " + strconv.Itoa(startx) + " x " + strconv.Itoa(starty) + " .......")
	}

	demo2.OnMouseLeave = func() {
		demo.SetText("the blue box is untouched right now.....")
	}

	app.AddElement(demo, demo2, card)
	app.RunLoop(func() { os.Exit(0) })

}
