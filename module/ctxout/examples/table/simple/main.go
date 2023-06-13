package main

import (
	"github.com/swaros/contxt/module/ctxout"
)

func main() {
	// first we need an logn text to show the wrapping
	text := " -just-to-fill-some-space- "
	for i := 0; i < 10; i++ {
		text += text
	}
	text = " : " + text

	// add table filter to ctxout
	ctxout.AddPostFilter(ctxout.NewTabOut())

	// create a table
	// size is the width of the table in percent of the terminal width
	table := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"hello"+text,    // the text content, must be a string and the first argument
				ctxout.Size(50), // the size of the cell in percent of the table width

			),
			ctxout.TD( // the next cell
				"world"+text,
				ctxout.Size(50),
			),
		),
		ctxout.Row( // the next row
			ctxout.TD(
				"hola"+text,
				ctxout.Size(50),
			),
			ctxout.TD(
				"mundo"+text,
				ctxout.Size(50),
			),
		), // and so on...
		ctxout.Row(
			ctxout.TD(
				"hallo"+text,
				ctxout.Size(50),
			),
			ctxout.TD(
				"welt"+text,
				ctxout.Size(50),
			),
		),
	)
	ctxout.PrintLn(table) // print the table

}
