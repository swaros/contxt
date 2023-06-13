package main

import (
	"github.com/swaros/contxt/module/ctxout"
)

func main() {
	// add table filter to ctxout
	ctxout.AddPostFilter(ctxout.NewTabOut())

	// create a table
	// size is the width of the table in percent of the terminal width
	table := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"hello",         // the text content, must be a string and the first argument
				ctxout.Size(50), // the size of the cell in percent of the table width
			),
			ctxout.TD( // the next cell
				"world",
				ctxout.Size(50),
			),
		),
		ctxout.Row( // the next row
			ctxout.TD(
				"hola",
				ctxout.Size(50),
			),
			ctxout.TD(
				"mundo",
				ctxout.Size(50),
			),
		), // and so on...
		ctxout.Row(
			ctxout.TD(
				"hallo",
				ctxout.Size(50),
			),
			ctxout.TD(
				"welt",
				ctxout.Size(40),
			),
		),
	)
	ctxout.PrintLn(table) // print the table

}
